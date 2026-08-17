// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"github.com/stretchr/testify/assert"
)

// Mock SOPRepository interface inside package service block
type mockSOPRepository struct {
	persistence.SOPRepository
	CreateFunc          func(ctx context.Context, s *model.SOP) error
	FindByIDFunc        func(ctx context.Context, id string) (*model.SOP, error)
	UpdateFunc          func(ctx context.Context, s *model.SOP) error
	CreateProcessFunc   func(ctx context.Context, p *model.SOPProcess) error
	FindProcessByIDFunc func(ctx context.Context, id string) (*model.SOPProcess, error)
	UpdateProcessFunc   func(ctx context.Context, p *model.SOPProcess) error
	CreateChunksFunc    func(ctx context.Context, chunks []*model.SOPChunk) error
	QuerySimilarityFunc func(ctx context.Context, query model.Float32Vector, limit int) ([]*model.SOPChunk, error)
	ListFunc                 func(ctx context.Context) ([]*model.SOP, error)
	ListRangeFunc            func(ctx context.Context, offset, limit int) ([]*model.SOP, error)
	DeleteFunc               func(ctx context.Context, id string) error
	ListProcessesFunc        func(ctx context.Context) ([]*model.SOPProcess, error)
	ListProcessesRangeFunc   func(ctx context.Context, offset, limit int) ([]*model.SOPProcess, error)
	DeleteProcessFunc        func(ctx context.Context, id string) error
}

func (m *mockSOPRepository) Create(ctx context.Context, s *model.SOP) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, s)
	}
	return nil
}

func (m *mockSOPRepository) FindByID(ctx context.Context, id string) (*model.SOP, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockSOPRepository) Update(ctx context.Context, s *model.SOP) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, s)
	}
	return nil
}

func (m *mockSOPRepository) CreateProcess(ctx context.Context, p *model.SOPProcess) error {
	if m.CreateProcessFunc != nil {
		return m.CreateProcessFunc(ctx, p)
	}
	return nil
}

func (m *mockSOPRepository) FindProcessByID(ctx context.Context, id string) (*model.SOPProcess, error) {
	if m.FindProcessByIDFunc != nil {
		return m.FindProcessByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockSOPRepository) UpdateProcess(ctx context.Context, p *model.SOPProcess) error {
	if m.UpdateProcessFunc != nil {
		return m.UpdateProcessFunc(ctx, p)
	}
	return nil
}

func (m *mockSOPRepository) CreateChunks(ctx context.Context, chunks []*model.SOPChunk) error {
	if m.CreateChunksFunc != nil {
		return m.CreateChunksFunc(ctx, chunks)
	}
	return nil
}

func (m *mockSOPRepository) QuerySimilarity(ctx context.Context, query model.Float32Vector, limit int) ([]*model.SOPChunk, error) {
	if m.QuerySimilarityFunc != nil {
		return m.QuerySimilarityFunc(ctx, query, limit)
	}
	return nil, nil
}

func (m *mockSOPRepository) List(ctx context.Context) ([]*model.SOP, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *mockSOPRepository) ListRange(ctx context.Context, offset, limit int) ([]*model.SOP, error) {
	if m.ListRangeFunc != nil {
		return m.ListRangeFunc(ctx, offset, limit)
	}
	return nil, nil
}

func (m *mockSOPRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *mockSOPRepository) ListProcesses(ctx context.Context) ([]*model.SOPProcess, error) {
	if m.ListProcessesFunc != nil {
		return m.ListProcessesFunc(ctx)
	}
	return nil, nil
}

func (m *mockSOPRepository) ListProcessesRange(ctx context.Context, offset, limit int) ([]*model.SOPProcess, error) {
	if m.ListProcessesRangeFunc != nil {
		return m.ListProcessesRangeFunc(ctx, offset, limit)
	}
	return nil, nil
}

func (m *mockSOPRepository) DeleteProcess(ctx context.Context, id string) error {
	if m.DeleteProcessFunc != nil {
		return m.DeleteProcessFunc(ctx, id)
	}
	return nil
}

// Mock EmbeddingGenerator
type mockEmbeddingGenerator struct {
	GenerateEmbeddingsFunc func(ctx context.Context, text string) (model.Float32Vector, error)
}

func (m *mockEmbeddingGenerator) GenerateEmbeddings(ctx context.Context, text string) (model.Float32Vector, error) {
	if m.GenerateEmbeddingsFunc != nil {
		return m.GenerateEmbeddingsFunc(ctx, text)
	}
	return make(model.Float32Vector, 768), nil
}

// Mock HTTP client
type mockHTTPClient struct {
	GetFunc  func(url string) (*http.Response, error)
	HeadFunc func(url string) (*http.Response, error)
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	if m.GetFunc != nil {
		return m.GetFunc(url)
	}
	return nil, nil
}

func (m *mockHTTPClient) Head(url string) (*http.Response, error) {
	if m.HeadFunc != nil {
		return m.HeadFunc(url)
	}
	return nil, nil
}

func TestRAGService_IngestSOPAsync(t *testing.T) {
	t.Run("successfully registers and enqueues async processing run", func(t *testing.T) {
		mockSopRepo := &mockSOPRepository{
			CreateFunc: func(ctx context.Context, s *model.SOP) error {
				assert.Equal(t, "Dallas Curbside Policy SOP", s.Title)
				assert.Equal(t, "http://omnimart.com/docs/curbside.html", s.CanonicalURL)
				s.ID = "sop-dallas-pickup"
				return nil
			},
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return &model.SOP{ID: id, Title: "Dallas Curbside Policy SOP", CanonicalURL: "http://omnimart.com/docs/curbside.html"}, nil
			},
			CreateProcessFunc: func(ctx context.Context, p *model.SOPProcess) error {
				assert.Equal(t, "sop-dallas-pickup", p.SOPID)
				assert.Equal(t, "IN_PROGRESS", p.Status)
				assert.NotNil(t, p.LockedAt)
				assert.Equal(t, "DIRECT_INGEST", *p.LockedBy)
				assert.False(t, p.IsActive)
				p.ID = "run-dallas-pickup"
				return nil
			},
			FindProcessByIDFunc: func(ctx context.Context, id string) (*model.SOPProcess, error) {
				return &model.SOPProcess{ID: id, SOPID: "sop-dallas-pickup", Status: "PENDING"}, nil
			},
			UpdateProcessFunc: func(ctx context.Context, p *model.SOPProcess) error {
				return nil
			},
		}

		embeddingGen := &mockEmbeddingGenerator{}
		svc := service.NewRAGService(mockSopRepo, embeddingGen)

		sop, process, err := svc.IngestSOPAsync(context.Background(), "Dallas Curbside Policy SOP", "http://omnimart.com/docs/curbside.html")

		assert.NoError(t, err)
		assert.NotNil(t, sop)
		assert.NotNil(t, process)
		assert.Equal(t, "sop-dallas-pickup", sop.ID)
		assert.Equal(t, "run-dallas-pickup", process.ID)
	})

	t.Run("missing parameter triggers validation failure", func(t *testing.T) {
		mockSopRepo := &mockSOPRepository{}
		embeddingGen := &mockEmbeddingGenerator{}
		svc := service.NewRAGService(mockSopRepo, embeddingGen)

		sop, process, err := svc.IngestSOPAsync(context.Background(), "", "http://omnimart.com/docs/curbside.html")

		assert.Error(t, err)
		assert.Nil(t, sop)
		assert.Nil(t, process)
		assert.Contains(t, err.Error(), "title and canonicalURL are mandatory parameters")
	})
}

func TestRAGService_ProcessSOPPipeline_HTML(t *testing.T) {
	t.Run("successfully processes HTML file, culls tags, and extracts cache metadata", func(t *testing.T) {
		sopID := "sop-123"
		processID := "process-456"

		existingSop := &model.SOP{
			ID:           sopID,
			Title:        "Produce Freshness Operations SOP",
			CanonicalURL: "https://omnimart.com/freshness.html",
		}

		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				assert.Equal(t, sopID, id)
				return existingSop, nil
			},
			FindProcessByIDFunc: func(ctx context.Context, id string) (*model.SOPProcess, error) {
				assert.Equal(t, processID, id)
				return &model.SOPProcess{ID: processID, SOPID: sopID, Status: "PENDING"}, nil
			},
		}

		// Mock HTTP Get returns high-fidelity HTML response matching cache indicators
		mockHTML := "<html><head><title>Freshness</title></head><body><h1>Rotations</h1><p>Rotate produce shelf stock daily.</p></body></html>"
		lastModified := "Wed, 21 Oct 2025 07:28:00 GMT"
		mockHTTP := &mockHTTPClient{
			GetFunc: func(url string) (*http.Response, error) {
				assert.Equal(t, "https://omnimart.com/freshness.html", url)

				headers := make(http.Header)
				headers.Set("Content-Type", "text/html; charset=utf-8")
				headers.Set("Last-Modified", lastModified)
				headers.Set("ETag", "etag-html-999")

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Header:     headers,
					Body:       io.NopCloser(bytes.NewBufferString(mockHTML)),
				}
				return resp, nil
			},
		}

		// Mock embedding returns stable 768 vector
		mockVector := make(model.Float32Vector, 768)
		mockVector[0] = 0.55
		embeddingGen := &mockEmbeddingGenerator{
			GenerateEmbeddingsFunc: func(ctx context.Context, text string) (model.Float32Vector, error) {
				// Assert HTML tags like <html>, <body>, <head> are culled successfully
				assert.NotContains(t, text, "<html>")
				assert.NotContains(t, text, "<body>")
				assert.Contains(t, text, "Rotate produce shelf stock daily")
				return mockVector, nil
			},
		}

		var createdChunks []*model.SOPChunk
		mockSopRepo.CreateChunksFunc = func(ctx context.Context, chunks []*model.SOPChunk) error {
			assert.NotEmpty(t, chunks)
			assert.Equal(t, sopID, chunks[0].SOPID)
			assert.Equal(t, processID, chunks[0].SOPProcessID)
			assert.Equal(t, mockVector, chunks[0].Embedding)
			createdChunks = chunks
			return nil
		}

		var capturedSOP *model.SOP
		mockSopRepo.UpdateFunc = func(ctx context.Context, s *model.SOP) error {
			capturedSOP = s
			return nil
		}

		var processCompleted bool
		mockSopRepo.UpdateProcessFunc = func(ctx context.Context, p *model.SOPProcess) error {
			assert.Equal(t, "COMPLETED", p.Status)
			assert.True(t, p.IsActive)
			processCompleted = true
			return nil
		}

		// Instantiate service with mock HTTP client directly
		svc := service.NewRAGServiceWithClient(mockSopRepo, embeddingGen, mockHTTP)

		// Run pipeline synchronously for assertions
		svc.ProcessSOPPipeline(context.Background(), sopID, processID)

		// Verification sweeps
		assert.Len(t, createdChunks, 1)
		assert.True(t, processCompleted)
		assert.NotNil(t, capturedSOP)

		var meta model.SOPMetadata
		err := json.Unmarshal(capturedSOP.Metadata, &meta)
		assert.NoError(t, err)
		assert.Equal(t, "HTML", meta.FileType)
		assert.Equal(t, "etag-html-999", meta.ETag)
		assert.Equal(t, int64(len(mockHTML)), meta.RawContentLength)
		assert.Equal(t, 1, meta.ProcessedChunks)
		assert.Empty(t, meta.SHA256Checksum)

		expectedDate, _ := http.ParseTime(lastModified)
		assert.NotNil(t, meta.CacheEffectiveDate)
		assert.True(t, meta.CacheEffectiveDate.Equal(expectedDate))
	})
}

func TestRAGService_ProcessSOPPipeline_Binary(t *testing.T) {
	t.Run("successfully processes PDF, maps SHA fingerprint, and seeds enums", func(t *testing.T) {
		sopID := "sop-pdf-7"
		processID := "process-pdf-8"

		existingSop := &model.SOP{
			ID:           sopID,
			Title:        "Vault Ingestion & Safe Audit SOP",
			CanonicalURL: "https://omnimart.com/docs/vault.pdf",
		}

		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return existingSop, nil
			},
			FindProcessByIDFunc: func(ctx context.Context, id string) (*model.SOPProcess, error) {
				return &model.SOPProcess{ID: processID, SOPID: sopID, Status: "PENDING"}, nil
			},
		}

		mockPDF := []byte("standard simulated binary pdf structure - verify cashier drops")
		sum := sha256.Sum256(mockPDF)
		expectedChecksum := hex.EncodeToString(sum[:])

		mockHTTP := &mockHTTPClient{
			GetFunc: func(url string) (*http.Response, error) {
				headers := make(http.Header)
				headers.Set("Content-Type", "application/pdf")
				resp := &http.Response{
					StatusCode: http.StatusOK,
					Header:     headers,
					Body:       io.NopCloser(bytes.NewBuffer(mockPDF)),
				}
				return resp, nil
			},
		}

		embeddingGen := &mockEmbeddingGenerator{}

		var capturedSOP *model.SOP
		mockSopRepo.UpdateFunc = func(ctx context.Context, s *model.SOP) error {
			capturedSOP = s
			return nil
		}

		var processCompleted bool
		mockSopRepo.UpdateProcessFunc = func(ctx context.Context, p *model.SOPProcess) error {
			assert.Equal(t, "COMPLETED", p.Status)
			processCompleted = true
			return nil
		}

		svc := service.NewRAGServiceWithClient(mockSopRepo, embeddingGen, mockHTTP)

		svc.ProcessSOPPipeline(context.Background(), sopID, processID)

		assert.True(t, processCompleted)
		assert.NotNil(t, capturedSOP)

		var meta model.SOPMetadata
		err := json.Unmarshal(capturedSOP.Metadata, &meta)
		assert.NoError(t, err)
		assert.Equal(t, "PDF", meta.FileType)
		assert.Equal(t, expectedChecksum, meta.SHA256Checksum)
		assert.Nil(t, meta.CacheEffectiveDate)
		assert.Equal(t, int64(len(mockPDF)), meta.RawContentLength)
	})
}

func TestRAGService_CheckSOPUpdates(t *testing.T) {
	sopID := "sop-audit-99"

	t.Run("returns false when ETag matches standard cache grids (No changes)", func(t *testing.T) {
		lastChecked := time.Now().Add(-1 * time.Hour)
		metaStruct := model.SOPMetadata{
			FileType:         "HTML",
			ETag:             "etag-grid-0",
			RawContentLength: 100,
			LastCheckedAt:    lastChecked,
		}
		metaBytes, _ := json.Marshal(metaStruct)

		existingSop := &model.SOP{
			ID:           sopID,
			Title:        "Dallas Lane Audit Guidelines",
			CanonicalURL: "https://omnimart.com/lane-audit.html",
			Metadata:     model.JSONB(metaBytes),
		}

		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return existingSop, nil
			},
		}

		mockHTTP := &mockHTTPClient{
			HeadFunc: func(url string) (*http.Response, error) {
				headers := make(http.Header)
				headers.Set("ETag", "etag-grid-0") // Identical ETag
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     headers,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}, nil
			},
		}

		var capturedSOP *model.SOP
		mockSopRepo.UpdateFunc = func(ctx context.Context, s *model.SOP) error {
			capturedSOP = s
			return nil
		}

		embeddingGen := &mockEmbeddingGenerator{}
		svc := service.NewRAGServiceWithClient(mockSopRepo, embeddingGen, mockHTTP)

		changed, err := svc.CheckSOPUpdates(context.Background(), sopID)

		assert.NoError(t, err)
		assert.False(t, changed)
		assert.NotNil(t, capturedSOP)

		var updatedMeta model.SOPMetadata
		_ = json.Unmarshal(capturedSOP.Metadata, &updatedMeta)
		assert.True(t, updatedMeta.LastCheckedAt.After(lastChecked))
	})

	t.Run("returns true and launches process when ETag changes (Drift Detected)", func(t *testing.T) {
		metaStruct := model.SOPMetadata{
			FileType: "HTML",
			ETag:     "etag-stale",
		}
		metaBytes, _ := json.Marshal(metaStruct)

		existingSop := &model.SOP{
			ID:           sopID,
			CanonicalURL: "https://omnimart.com/lane-audit.html",
			Metadata:     model.JSONB(metaBytes),
		}

		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return existingSop, nil
			},
			CreateProcessFunc: func(ctx context.Context, p *model.SOPProcess) error {
				assert.Equal(t, sopID, p.SOPID)
				assert.Equal(t, "IN_PROGRESS", p.Status)
				assert.NotNil(t, p.LockedAt)
				assert.Equal(t, "DIRECT_INGEST", *p.LockedBy)
				return nil
			},
			FindProcessByIDFunc: func(ctx context.Context, id string) (*model.SOPProcess, error) {
				return &model.SOPProcess{ID: id, SOPID: sopID, Status: "PENDING"}, nil
			},
			UpdateProcessFunc: func(ctx context.Context, p *model.SOPProcess) error {
				return nil
			},
			UpdateFunc: func(ctx context.Context, s *model.SOP) error {
				return nil
			},
			CreateChunksFunc: func(ctx context.Context, chunks []*model.SOPChunk) error {
				return nil
			},
		}

		mockHTTP := &mockHTTPClient{
			HeadFunc: func(url string) (*http.Response, error) {
				headers := make(http.Header)
				headers.Set("ETag", "etag-fresh-value-777") // Modified ETag
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     headers,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}, nil
			},
			GetFunc: func(url string) (*http.Response, error) {
				headers := make(http.Header)
				headers.Set("Content-Type", "text/html; charset=utf-8")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     headers,
					Body:       io.NopCloser(bytes.NewBufferString("<html><body>Fresh</body></html>")),
				}, nil
			},
		}

		embeddingGen := &mockEmbeddingGenerator{}
		svc := service.NewRAGServiceWithClient(mockSopRepo, embeddingGen, mockHTTP)

		changed, err := svc.CheckSOPUpdates(context.Background(), sopID)

		assert.NoError(t, err)
		assert.True(t, changed)
	})
}

func TestRAGService_CheckSOPUpdates_ErrorPaths(t *testing.T) {
	t.Run("empty sopID fails", func(t *testing.T) {
		svc := service.NewRAGService(nil, nil)
		changed, err := svc.CheckSOPUpdates(context.Background(), "")
		assert.Error(t, err)
		assert.False(t, changed)
		assert.Contains(t, err.Error(), "sopID is mandatory")
	})

	t.Run("SOP not found in db fails", func(t *testing.T) {
		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return nil, errors.New("not found")
			},
		}
		svc := service.NewRAGService(mockSopRepo, nil)
		changed, err := svc.CheckSOPUpdates(context.Background(), "sop-123")
		assert.Error(t, err)
		assert.False(t, changed)
		assert.Contains(t, err.Error(), "failed to query standard SOP metadata")
	})

	t.Run("unmarshal metadata fails", func(t *testing.T) {
		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return &model.SOP{
					ID:       "sop-123",
					Metadata: model.JSONB("invalid-json"),
				}, nil
			},
		}
		svc := service.NewRAGService(mockSopRepo, nil)
		changed, err := svc.CheckSOPUpdates(context.Background(), "sop-123")
		assert.Error(t, err)
		assert.False(t, changed)
		assert.Contains(t, err.Error(), "failed to parse existing metadata layout")
	})

	t.Run("HEAD request fails", func(t *testing.T) {
		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return &model.SOP{
					ID:           "sop-123",
					CanonicalURL: "http://omnimart.com/docs/curbside.html",
					Metadata:     model.JSONB(`{"file_type":"HTML"}`),
				}, nil
			},
		}
		mockHTTP := &mockHTTPClient{
			HeadFunc: func(url string) (*http.Response, error) {
				return nil, errors.New("network error")
			},
		}
		svc := service.NewRAGServiceWithClient(mockSopRepo, nil, mockHTTP)
		changed, err := svc.CheckSOPUpdates(context.Background(), "sop-123")
		assert.Error(t, err)
		assert.False(t, changed)
		assert.Contains(t, err.Error(), "HTTP target validation HEAD checks failed")
	})

	t.Run("non-HTML Content-Length change triggers update", func(t *testing.T) {
		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return &model.SOP{
					ID:           "sop-123",
					CanonicalURL: "http://omnimart.com/docs/sop.pdf",
					Metadata:     model.JSONB(`{"file_type":"PDF","raw_content_length":100}`),
				}, nil
			},
			CreateProcessFunc: func(ctx context.Context, p *model.SOPProcess) error {
				return nil
			},
		}
		mockHTTP := &mockHTTPClient{
			HeadFunc: func(url string) (*http.Response, error) {
				headers := make(http.Header)
				headers.Set("Content-Length", "200")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     headers,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				}, nil
			},
		}
		embeddingGen := &mockEmbeddingGenerator{}
		svc := service.NewRAGServiceWithClient(mockSopRepo, embeddingGen, mockHTTP)
		changed, err := svc.CheckSOPUpdates(context.Background(), "sop-123")
		assert.NoError(t, err)
		assert.True(t, changed)
	})

	t.Run("non-HTML deep checksum change triggers update", func(t *testing.T) {
		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return &model.SOP{
					ID:           "sop-123",
					CanonicalURL: "http://omnimart.com/docs/sop.pdf",
					Metadata:     model.JSONB(`{"file_type":"PDF","raw_content_length":100,"sha256_checksum":"old-sha"}`),
				}, nil
			},
			CreateProcessFunc: func(ctx context.Context, p *model.SOPProcess) error {
				return nil
			},
		}
		mockHTTP := &mockHTTPClient{
			HeadFunc: func(url string) (*http.Response, error) {
				headers := make(http.Header)
				headers.Set("Content-Length", "100")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     headers,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				}, nil
			},
			GetFunc: func(url string) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("fresh pdf content")),
				}, nil
			},
		}
		embeddingGen := &mockEmbeddingGenerator{}
		svc := service.NewRAGServiceWithClient(mockSopRepo, embeddingGen, mockHTTP)
		changed, err := svc.CheckSOPUpdates(context.Background(), "sop-123")
		assert.NoError(t, err)
		assert.True(t, changed)
	})
}

func TestRAGService_ProcessSOPPipeline_ErrorPaths(t *testing.T) {
	t.Run("SOP record not found", func(t *testing.T) {
		var failedMessage string
		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return nil, errors.New("not found")
			},
			FindProcessByIDFunc: func(ctx context.Context, id string) (*model.SOPProcess, error) {
				return &model.SOPProcess{ID: id, SOPID: "sop-123"}, nil
			},
			UpdateProcessFunc: func(ctx context.Context, p *model.SOPProcess) error {
				assert.Contains(t, p.Status, "FAILED")
				assert.Contains(t, p.Status, "failed to lookup base SOP")
				failedMessage = p.Status
				return nil
			},
		}
		svc := service.NewRAGService(mockSopRepo, nil)
		svc.ProcessSOPPipeline(context.Background(), "sop-123", "process-123")
		assert.NotEmpty(t, failedMessage)
	})

	t.Run("HTTP download fails", func(t *testing.T) {
		var failedMessage string
		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return &model.SOP{ID: "sop-123", CanonicalURL: "http://badurl.com"}, nil
			},
			FindProcessByIDFunc: func(ctx context.Context, id string) (*model.SOPProcess, error) {
				return &model.SOPProcess{ID: id, SOPID: "sop-123"}, nil
			},
			UpdateProcessFunc: func(ctx context.Context, p *model.SOPProcess) error {
				assert.Contains(t, p.Status, "FAILED")
				assert.Contains(t, p.Status, "download endpoint lookup failed")
				failedMessage = p.Status
				return nil
			},
		}
		mockHTTP := &mockHTTPClient{
			GetFunc: func(url string) (*http.Response, error) {
				return nil, errors.New("dns lookup failed")
			},
		}
		svc := service.NewRAGServiceWithClient(mockSopRepo, nil, mockHTTP)
		svc.ProcessSOPPipeline(context.Background(), "sop-123", "process-123")
		assert.NotEmpty(t, failedMessage)
	})

	t.Run("HTTP download returns non-200", func(t *testing.T) {
		var failedMessage string
		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return &model.SOP{ID: "sop-123", CanonicalURL: "http://auth.com"}, nil
			},
			FindProcessByIDFunc: func(ctx context.Context, id string) (*model.SOPProcess, error) {
				return &model.SOPProcess{ID: id, SOPID: "sop-123"}, nil
			},
			UpdateProcessFunc: func(ctx context.Context, p *model.SOPProcess) error {
				assert.Contains(t, p.Status, "FAILED")
				assert.Contains(t, p.Status, "endpoint returned invalid status: 403 Forbidden")
				failedMessage = p.Status
				return nil
			},
		}
		mockHTTP := &mockHTTPClient{
			GetFunc: func(url string) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusForbidden,
					Status:     "403 Forbidden",
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				}, nil
			},
		}
		svc := service.NewRAGServiceWithClient(mockSopRepo, nil, mockHTTP)
		svc.ProcessSOPPipeline(context.Background(), "sop-123", "process-123")
		assert.NotEmpty(t, failedMessage)
	})

	t.Run("binary PDF classification and chunking success", func(t *testing.T) {
		var createdChunks []*model.SOPChunk
		var updatedSOP *model.SOP
		var processCompleted bool

		mockSopRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				return &model.SOP{ID: "sop-123", CanonicalURL: "http://pdf.com"}, nil
			},
			FindProcessByIDFunc: func(ctx context.Context, id string) (*model.SOPProcess, error) {
				return &model.SOPProcess{ID: id, SOPID: "sop-123"}, nil
			},
			CreateChunksFunc: func(ctx context.Context, chunks []*model.SOPChunk) error {
				createdChunks = chunks
				return nil
			},
			UpdateFunc: func(ctx context.Context, s *model.SOP) error {
				updatedSOP = s
				return nil
			},
			UpdateProcessFunc: func(ctx context.Context, p *model.SOPProcess) error {
				assert.Equal(t, "COMPLETED", p.Status)
				processCompleted = true
				return nil
			},
		}
		mockHTTP := &mockHTTPClient{
			GetFunc: func(url string) (*http.Response, error) {
				headers := make(http.Header)
				headers.Set("Content-Type", "application/pdf")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     headers,
					Body:       io.NopCloser(bytes.NewBufferString("dummy pdf content stream")),
				}, nil
			},
		}
		embeddingGen := &mockEmbeddingGenerator{}
		svc := service.NewRAGServiceWithClient(mockSopRepo, embeddingGen, mockHTTP)
		svc.ProcessSOPPipeline(context.Background(), "sop-123", "process-123")

		assert.True(t, processCompleted)
		assert.NotNil(t, updatedSOP)
		assert.NotEmpty(t, createdChunks)
		var meta model.SOPMetadata
		err := json.Unmarshal(updatedSOP.Metadata, &meta)
		assert.NoError(t, err)
		assert.Equal(t, "PDF", meta.FileType)
		assert.NotEmpty(t, meta.SHA256Checksum)
	})
}

func TestRAGService_CRUD(t *testing.T) {
	t.Run("SOP CRUD success", func(t *testing.T) {
		expectedSOP := &model.SOP{ID: "sop-1", Title: "Title 1"}
		expectedList := []*model.SOP{
			expectedSOP,
			{ID: "sop-2", Title: "Title 2"},
		}
		
		calledDelete := false
		mockRepo := &mockSOPRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.SOP, error) {
				assert.Equal(t, "sop-1", id)
				return expectedSOP, nil
			},
			ListFunc: func(ctx context.Context) ([]*model.SOP, error) {
				return expectedList, nil
			},
			ListRangeFunc: func(ctx context.Context, offset, limit int) ([]*model.SOP, error) {
				assert.Equal(t, 1, offset)
				assert.Equal(t, 10, limit)
				return expectedList[1:], nil
			},
			DeleteFunc: func(ctx context.Context, id string) error {
				assert.Equal(t, "sop-1", id)
				calledDelete = true
				return nil
			},
		}
		
		svc := service.NewRAGService(mockRepo, nil)
		
		resGet, err := svc.GetSOPByID(context.Background(), "sop-1")
		assert.NoError(t, err)
		assert.Equal(t, expectedSOP, resGet)
		
		resList, err := svc.ListSOPs(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedList, resList)
		
		resListRange, err := svc.ListSOPsRange(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, expectedList[1:], resListRange)
		
		err = svc.DeleteSOP(context.Background(), "sop-1")
		assert.NoError(t, err)
		assert.True(t, calledDelete)
	})

	t.Run("SOPProcess CRUD success", func(t *testing.T) {
		expectedProc := &model.SOPProcess{ID: "proc-1", Status: "COMPLETED"}
		expectedList := []*model.SOPProcess{
			expectedProc,
			{ID: "proc-2", Status: "FAILED"},
		}
		
		calledDelete := false
		mockRepo := &mockSOPRepository{
			FindProcessByIDFunc: func(ctx context.Context, id string) (*model.SOPProcess, error) {
				assert.Equal(t, "proc-1", id)
				return expectedProc, nil
			},
			ListProcessesFunc: func(ctx context.Context) ([]*model.SOPProcess, error) {
				return expectedList, nil
			},
			ListProcessesRangeFunc: func(ctx context.Context, offset, limit int) ([]*model.SOPProcess, error) {
				assert.Equal(t, 1, offset)
				assert.Equal(t, 10, limit)
				return expectedList[1:], nil
			},
			DeleteProcessFunc: func(ctx context.Context, id string) error {
				assert.Equal(t, "proc-1", id)
				calledDelete = true
				return nil
			},
		}
		
		svc := service.NewRAGService(mockRepo, nil)
		
		resGet, err := svc.GetProcessByID(context.Background(), "proc-1")
		assert.NoError(t, err)
		assert.Equal(t, expectedProc, resGet)
		
		resList, err := svc.ListProcesses(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedList, resList)
		
		resListRange, err := svc.ListProcessesRange(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, expectedList[1:], resListRange)
		
		err = svc.DeleteProcess(context.Background(), "proc-1")
		assert.NoError(t, err)
		assert.True(t, calledDelete)
	})
}

func TestRAGService_EmbeddingGeneratorAndCRUD(t *testing.T) {
	t.Run("default embedding generator generates 768 vector", func(t *testing.T) {
		gen := service.NewDefaultEmbeddingGenerator()
		vec, err := gen.GenerateEmbeddings(context.Background(), "test sop content")
		assert.NoError(t, err)
		assert.Len(t, vec, 768)
	})

	t.Run("RegisterSOP, SaveChunks, QuerySimilarity", func(t *testing.T) {
		sop := &model.SOP{ID: "sop-1"}
		chunks := []*model.SOPChunk{{ID: "chunk-1"}}
		mockRepo := &mockSOPRepository{
			CreateFunc: func(ctx context.Context, s *model.SOP) error {
				assert.Equal(t, "sop-1", s.ID)
				return nil
			},
			CreateChunksFunc: func(ctx context.Context, c []*model.SOPChunk) error {
				assert.Len(t, c, 1)
				return nil
			},
			QuerySimilarityFunc: func(ctx context.Context, embedding model.Float32Vector, limit int) ([]*model.SOPChunk, error) {
				return chunks, nil
			},
		}

		svc := service.NewRAGService(mockRepo, nil)
		err := svc.RegisterSOP(context.Background(), sop)
		assert.NoError(t, err)

		err = svc.SaveChunks(context.Background(), chunks)
		assert.NoError(t, err)

		res, err := svc.QuerySimilarity(context.Background(), model.Float32Vector{0.1}, 5)
		assert.NoError(t, err)
		assert.Equal(t, chunks, res)
	})

	t.Run("defaultHTTPClient Head", func(t *testing.T) {
		httpClient := service.NewRAGService(nil, nil)
		assert.NotNil(t, httpClient)
	})
}
