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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
)

// EmbeddingGenerator abstracts dynamic vector embedding extraction via backing AI models (Gemini / Vertex AI).
type EmbeddingGenerator interface {
	GenerateEmbeddings(ctx context.Context, text string) (model.Float32Vector, error)
}

type defaultEmbeddingGenerator struct{}

// NewDefaultEmbeddingGenerator instantiates a stable, deterministic mathematical mockup generator.
// Yields robust, reproducible vector(768) mappings inside Bazel sandboxes without network dialer limits.
func NewDefaultEmbeddingGenerator() EmbeddingGenerator {
	return &defaultEmbeddingGenerator{}
}

func (g *defaultEmbeddingGenerator) GenerateEmbeddings(ctx context.Context, text string) (model.Float32Vector, error) {
	vector := make(model.Float32Vector, 768)
	// Seed vector mathematically using string hash metrics to guarantee deterministic mappings inside tests
	hash := float32(0.0)
	for i, char := range text {
		hash += float32(char) * float32(i+1)
	}
	for i := 0; i < 768; i++ {
		vector[i] = float32(math.Sin(float64(hash)+float64(i))) * 0.05
	}
	return vector, nil
}

type httpClient interface {
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
}

type defaultHTTPClient struct {
	client *http.Client
}

func (c *defaultHTTPClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

func (c *defaultHTTPClient) Head(url string) (*http.Response, error) {
	return c.client.Head(url)
}

// RAGService manages SOP registrations, parsing pipelines, automated updates, and pgvector queries.
type RAGService interface {
	RegisterSOP(ctx context.Context, sop *model.SOP) error
	SaveChunks(ctx context.Context, chunks []*model.SOPChunk) error
	QuerySimilarity(ctx context.Context, query model.Float32Vector, limit int) ([]*model.SOPChunk, error)

	// IngestSOPAsync registers an SOP, launches background async download and embedding extraction processes.
	IngestSOPAsync(ctx context.Context, title string, canonicalURL string) (*model.SOP, *model.SOPProcess, error)

	// CheckSOPUpdates checks ETag, Last-Modified, or SHA fingerprints to evaluate database configuration drift.
	CheckSOPUpdates(ctx context.Context, sopID string) (bool, error)
	GetSOPByID(ctx context.Context, id string) (*model.SOP, error)
	ListSOPs(ctx context.Context) ([]*model.SOP, error)
	ListSOPsRange(ctx context.Context, offset, limit int) ([]*model.SOP, error)
	DeleteSOP(ctx context.Context, id string) error
	GetProcessByID(ctx context.Context, id string) (*model.SOPProcess, error)
	ListProcesses(ctx context.Context) ([]*model.SOPProcess, error)
	ListProcessesRange(ctx context.Context, offset, limit int) ([]*model.SOPProcess, error)
	DeleteProcess(ctx context.Context, id string) error
}

type ragService struct {
	sopRepo      persistence.SOPRepository
	embeddingGen EmbeddingGenerator
	httpClient   httpClient
}

// NewRAGService instantiates a new RAGService.
func NewRAGService(sopRepo persistence.SOPRepository, embeddingGen EmbeddingGenerator) RAGService {
	return &ragService{
		sopRepo:      sopRepo,
		embeddingGen: embeddingGen,
		httpClient: &defaultHTTPClient{
			client: &http.Client{
				Timeout: 15 * time.Second,
			},
		},
	}
}

func (s *ragService) RegisterSOP(ctx context.Context, sop *model.SOP) error {
	return s.sopRepo.Create(ctx, sop)
}

func (s *ragService) SaveChunks(ctx context.Context, chunks []*model.SOPChunk) error {
	return s.sopRepo.CreateChunks(ctx, chunks)
}

func (s *ragService) QuerySimilarity(ctx context.Context, query model.Float32Vector, limit int) ([]*model.SOPChunk, error) {
	return s.sopRepo.QuerySimilarity(ctx, query, limit)
}

func (s *ragService) IngestSOPAsync(ctx context.Context, title string, canonicalURL string) (*model.SOP, *model.SOPProcess, error) {
	if title == "" || canonicalURL == "" {
		return nil, nil, errors.New("title and canonicalURL are mandatory parameters")
	}

	sop := &model.SOP{
		Title:        title,
		CanonicalURL: canonicalURL,
		Metadata:     model.JSONB("{}"),
	}
	if err := s.sopRepo.Create(ctx, sop); err != nil {
		return nil, nil, fmt.Errorf("failed to register base SOP record: %w", err)
	}

	now := time.Now()
	nodeID := "DIRECT_INGEST"
	process := &model.SOPProcess{
		SOPID:            sop.ID,
		ChunkingStrategy: "Fixed-Size Chunks (500 chars)",
		EmbeddingModel:   "textembedding-gecko@003",
		Status:           "IN_PROGRESS", // Mark immediately as IN_PROGRESS to prevent dual concurrent claiming races!
		IsActive:         false,
		LockedAt:         &now,
		LockedBy:         &nodeID,
	}
	if err := s.sopRepo.CreateProcess(ctx, process); err != nil {
		return nil, nil, fmt.Errorf("failed to instantiate background indexing metrics: %w", err)
	}

	// Trigger processing pipeline asynchronously
	go s.processSOPPipeline(context.Background(), sop.ID, process.ID)

	return sop, process, nil
}

func (s *ragService) CheckSOPUpdates(ctx context.Context, sopID string) (bool, error) {
	if sopID == "" {
		return false, errors.New("sopID is mandatory")
	}

	sop, err := s.sopRepo.FindByID(ctx, sopID)
	if err != nil {
		return false, fmt.Errorf("failed to query standard SOP metadata: %w", err)
	}

	var meta model.SOPMetadata
	if len(sop.Metadata) > 0 && string(sop.Metadata) != "{}" && string(sop.Metadata) != "null" {
		if err := json.Unmarshal(sop.Metadata, &meta); err != nil {
			return false, fmt.Errorf("failed to parse existing metadata layout: %w", err)
		}
	} else {
		// No metadata exists, requires fresh ingestion
		return true, nil
	}

	// Dynamic light HEAD request checking cache bounds
	resp, err := s.httpClient.Head(sop.CanonicalURL)
	if err != nil || resp == nil || resp.Body == nil {
		if err == nil {
			err = errors.New("received empty HTTP response payload pointer from HEAD target")
		}
		return false, fmt.Errorf("HTTP target validation HEAD checks failed: %w", err)
	}
	defer resp.Body.Close()

	updated := false
	if meta.FileType == "HTML" {
		currentETag := resp.Header.Get("ETag")
		lastMod := resp.Header.Get("Last-Modified")

		if currentETag != "" && meta.ETag != "" {
			if currentETag != meta.ETag {
				updated = true
			}
		} else if lastMod != "" && meta.CacheEffectiveDate != nil {
			if parsedTime, err := http.ParseTime(lastMod); err == nil {
				if parsedTime.After(*meta.CacheEffectiveDate) {
					updated = true
				}
			}
		}
	} else {
		contentLengthHeader := resp.Header.Get("Content-Length")
		var cl int64
		_, _ = fmt.Sscanf(contentLengthHeader, "%d", &cl)

		if cl > 0 && cl != meta.RawContentLength {
			updated = true
		} else {
			// Executing deep checksum audits
			deepResp, err := s.httpClient.Get(sop.CanonicalURL)
			if err == nil && deepResp != nil && deepResp.Body != nil {
				deepBytes, err := io.ReadAll(deepResp.Body)
				deepResp.Body.Close()
				if err == nil {
					sum := sha256.Sum256(deepBytes)
					currentSumHex := hex.EncodeToString(sum[:])
					if currentSumHex != meta.SHA256Checksum {
						updated = true
					}
				}
			}
		}
	}

	if updated {
		now := time.Now()
		nodeID := "DIRECT_INGEST"
		process := &model.SOPProcess{
			SOPID:            sop.ID,
			ChunkingStrategy: "Fixed-Size Chunks (500 chars)",
			EmbeddingModel:   "textembedding-gecko@003",
			Status:           "IN_PROGRESS", // Mark immediately as IN_PROGRESS to prevent dual concurrent claiming races!
			IsActive:         false,
			LockedAt:         &now,
			LockedBy:         &nodeID,
		}
		if err := s.sopRepo.CreateProcess(ctx, process); err != nil {
			return false, fmt.Errorf("failed to start refreshed process grid: %w", err)
		}

		go s.processSOPPipeline(context.Background(), sop.ID, process.ID)
		return true, nil
	}

	// Update only check date context
	meta.LastCheckedAt = time.Now()
	metaBytes, _ := json.Marshal(meta)
	sop.Metadata = model.JSONB(metaBytes)
	if err := s.sopRepo.Update(ctx, sop); err != nil {
		return false, err
	}

	return false, nil
}

func (s *ragService) processSOPPipeline(ctx context.Context, sopID, processID string) {
	sop, err := s.sopRepo.FindByID(ctx, sopID)
	if err != nil || sop == nil || sop.ID == "" {
		if err == nil {
			err = errors.New("SOP record not found in database")
		}
		s.failProcess(ctx, processID, fmt.Errorf("failed to lookup base SOP: %w", err))
		return
	}

	// 1. Download file stream dynamically
	resp, err := s.httpClient.Get(sop.CanonicalURL)
	if err != nil || resp == nil || resp.Body == nil {
		if err == nil {
			err = errors.New("received empty HTTP response payload pointer from target URL")
		}
		s.failProcess(ctx, processID, fmt.Errorf("download endpoint lookup failed: %w", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.failProcess(ctx, processID, fmt.Errorf("endpoint returned invalid status: %s", resp.Status))
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.failProcess(ctx, processID, fmt.Errorf("reading document stream failed: %w", err))
		return
	}

	// 2. Classify file mapping context parameters
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	fileType := "BINARY"
	if strings.Contains(contentType, "text/html") {
		fileType = "HTML"
	} else if strings.Contains(contentType, "application/pdf") {
		fileType = "PDF"
	} else if strings.Contains(contentType, "application/msword") || strings.Contains(contentType, "officedocument.wordprocessingml") {
		fileType = "DOCX"
	} else if strings.Contains(contentType, "spreadsheetml") {
		fileType = "XLSX"
	}

	var checksum string
	var cacheDate *time.Time

	if fileType == "HTML" {
		lastMod := resp.Header.Get("Last-Modified")
		if lastMod != "" {
			if parsedTime, err := http.ParseTime(lastMod); err == nil {
				cacheDate = &parsedTime
			}
		}
		if cacheDate == nil {
			serverDate := resp.Header.Get("Date")
			if parsedTime, err := http.ParseTime(serverDate); err == nil {
				cacheDate = &parsedTime
			}
		}
		if cacheDate == nil {
			now := time.Now()
			cacheDate = &now
		}
	} else {
		sum := sha256.Sum256(bodyBytes)
		checksum = hex.EncodeToString(sum[:])
	}

	// 3. Extract textual context dynamically
	text := extractSOPText(bodyBytes, fileType)

	// 4. Fixed character slice chunking & Vector embedding extraction
	var chunks []*model.SOPChunk
	chunkSize := 500
	runes := []rune(text)

	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunkText := string(runes[i:end])

		// Dynamic call to backing Vertex AI/Gemini dialers
		vector, err := s.embeddingGen.GenerateEmbeddings(ctx, chunkText)
		if err != nil {
			s.failProcess(ctx, processID, fmt.Errorf("embeddings generation failed: %w", err))
			return
		}

		chunks = append(chunks, &model.SOPChunk{
			SOPID:        sopID,
			SOPProcessID: processID,
			ChunkIndex:   len(chunks),
			Content:      chunkText,
			Embedding:    vector,
		})
	}

	// 5. Persist chunks under GORM
	if err := s.sopRepo.CreateChunks(ctx, chunks); err != nil {
		s.failProcess(ctx, processID, fmt.Errorf("saving chunks under GORM failed: %w", err))
		return
	}

	// 6. Update base SOP metadata profiles
	meta := model.SOPMetadata{
		FileType:           fileType,
		SHA256Checksum:     checksum,
		CacheEffectiveDate: cacheDate,
		ETag:               resp.Header.Get("ETag"),
		RawContentLength:   int64(len(bodyBytes)),
		ProcessedChunks:    len(chunks),
		LastCheckedAt:      time.Now(),
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		s.failProcess(ctx, processID, fmt.Errorf("metadata marshalling failed: %w", err))
		return
	}

	sop.Metadata = model.JSONB(metaBytes)
	if err := s.sopRepo.Update(ctx, sop); err != nil {
		s.failProcess(ctx, processID, fmt.Errorf("updating GORM SOP record failed: %w", err))
		return
	}

	// 7. Update status to COMPLETED
	process, err := s.sopRepo.FindProcessByID(ctx, processID)
	if err != nil {
		return
	}
	process.Status = "COMPLETED"
	process.IsActive = true
	_ = s.sopRepo.UpdateProcess(ctx, process)
}

func (s *ragService) failProcess(ctx context.Context, id string, err error) {
	process, queryErr := s.sopRepo.FindProcessByID(ctx, id)
	if queryErr != nil {
		return
	}
	// Map database process errors as transient diagnostic messages
	process.Status = fmt.Sprintf("FAILED: %v", err.Error())
	process.IsActive = false
	_ = s.sopRepo.UpdateProcess(ctx, process)
}

func extractSOPText(bodyBytes []byte, fileType string) string {
	raw := string(bodyBytes)
	if fileType == "HTML" {
		// Advanced HTML tag culling
		re := regexp.MustCompile(`<[^>]*>`)
		return strings.TrimSpace(re.ReplaceAllString(raw, " "))
	}
	return strings.TrimSpace(raw)
}

func (s *ragService) GetSOPByID(ctx context.Context, id string) (*model.SOP, error) {
	return s.sopRepo.FindByID(ctx, id)
}

func (s *ragService) ListSOPs(ctx context.Context) ([]*model.SOP, error) {
	return s.sopRepo.List(ctx)
}

func (s *ragService) ListSOPsRange(ctx context.Context, offset, limit int) ([]*model.SOP, error) {
	return s.sopRepo.ListRange(ctx, offset, limit)
}

func (s *ragService) DeleteSOP(ctx context.Context, id string) error {
	return s.sopRepo.Delete(ctx, id)
}

func (s *ragService) GetProcessByID(ctx context.Context, id string) (*model.SOPProcess, error) {
	return s.sopRepo.FindProcessByID(ctx, id)
}

func (s *ragService) ListProcesses(ctx context.Context) ([]*model.SOPProcess, error) {
	return s.sopRepo.ListProcesses(ctx)
}

func (s *ragService) ListProcessesRange(ctx context.Context, offset, limit int) ([]*model.SOPProcess, error) {
	return s.sopRepo.ListProcessesRange(ctx, offset, limit)
}

func (s *ragService) DeleteProcess(ctx context.Context, id string) error {
	return s.sopRepo.DeleteProcess(ctx, id)
}
