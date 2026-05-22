package service

import (
	"context"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
)

// RAGService manages SOP registrations, processing chunk storage, and vector similarity queries.
type RAGService interface {
	RegisterSOP(ctx context.Context, sop *model.SOP) error
	SaveChunks(ctx context.Context, chunks []*model.SOPChunk) error
	QuerySimilarity(ctx context.Context, query model.Float32Vector, limit int) ([]*model.SOPChunk, error)
}

type ragService struct {
	sopRepo persistence.SOPRepository
}

// NewRAGService instantiates a new RAGService.
func NewRAGService(sopRepo persistence.SOPRepository) RAGService {
	return &ragService{
		sopRepo: sopRepo,
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
