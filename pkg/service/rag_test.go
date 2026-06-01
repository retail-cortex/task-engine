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

package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
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
		svc := NewRAGService(mockSopRepo, embeddingGen)

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
		svc := NewRAGService(mockSopRepo, embeddingGen)

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

		// Instantiate service and inject mock http client
		svcInterface := NewRAGService(mockSopRepo, embeddingGen)
		svc := svcInterface.(*ragService)
		svc.httpClient = mockHTTP

		// Run pipeline synchronously for assertions
		svc.processSOPPipeline(context.Background(), sopID, processID)

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

		svcInterface := NewRAGService(mockSopRepo, embeddingGen)
		svc := svcInterface.(*ragService)
		svc.httpClient = mockHTTP

		svc.processSOPPipeline(context.Background(), sopID, processID)

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
		svcInterface := NewRAGService(mockSopRepo, embeddingGen)
		svc := svcInterface.(*ragService)
		svc.httpClient = mockHTTP

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
		svcInterface := NewRAGService(mockSopRepo, embeddingGen)
		svc := svcInterface.(*ragService)
		svc.httpClient = mockHTTP

		changed, err := svc.CheckSOPUpdates(context.Background(), sopID)

		assert.NoError(t, err)
		assert.True(t, changed)
	})
}
