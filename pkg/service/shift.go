package service

import (
	"context"
	"encoding/json"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
)

// ShiftService manages employee shifts and agent session initialization.
type ShiftService interface {
	InitializeShift(ctx context.Context, userID, shiftInstanceID string) (*model.ShiftAgentSession, error)
}

type shiftService struct {
	sessionRepo persistence.ShiftAgentSessionRepository
	userRepo    persistence.UserRepository
}

// NewShiftService instantiates a new ShiftService.
func NewShiftService(sessionRepo persistence.ShiftAgentSessionRepository, userRepo persistence.UserRepository) ShiftService {
	return &shiftService{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
	}
}

func (s *shiftService) InitializeShift(ctx context.Context, userID, shiftInstanceID string) (*model.ShiftAgentSession, error) {
	// Checks if a session context has already been provisioned for this shift
	session, err := s.sessionRepo.FindByShift(ctx, userID, shiftInstanceID)
	if err == nil && session != nil && session.ID != "" {
		return session, nil
	}

	// Fetch profile context
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build structured grounded context for the Gemini ADK Session Store
	contextMap := map[string]interface{}{
		"user_email": user.Email,
		"user_name":  user.Name,
		"roles":      user.Roles,
	}
	contextBytes, err := json.Marshal(contextMap)
	if err != nil {
		return nil, err
	}

	newSession := &model.ShiftAgentSession{
		AssigneeID:      userID,
		ShiftInstanceID: shiftInstanceID,
		MessageHistory:  model.JSONB("[]"),
		SessionContext:  model.JSONB(contextBytes),
		Status:          "ACTIVE",
	}

	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return nil, err
	}

	return newSession, nil
}
