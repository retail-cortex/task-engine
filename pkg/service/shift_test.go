package service

import (
	"context"
	"errors"
	"testing"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/stretchr/testify/assert"
)

// MockUserRepository implements persistence.UserRepository for testing.
type MockUserRepository struct {
	persistence.UserRepository
	FindByIDFunc func(ctx context.Context, id string) (*model.User, error)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

// MockShiftAgentSessionRepository implements persistence.ShiftAgentSessionRepository for testing.
type MockShiftAgentSessionRepository struct {
	persistence.ShiftAgentSessionRepository
	FindByShiftFunc func(ctx context.Context, assigneeID, shiftInstanceID string) (*model.ShiftAgentSession, error)
	CreateFunc      func(ctx context.Context, s *model.ShiftAgentSession) error
}

func (m *MockShiftAgentSessionRepository) FindByShift(ctx context.Context, assigneeID, shiftInstanceID string) (*model.ShiftAgentSession, error) {
	if m.FindByShiftFunc != nil {
		return m.FindByShiftFunc(ctx, assigneeID, shiftInstanceID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockShiftAgentSessionRepository) Create(ctx context.Context, s *model.ShiftAgentSession) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, s)
	}
	return nil
}

func TestShiftService_InitializeShift(t *testing.T) {
	t.Run("session already exists", func(t *testing.T) {
		existingSession := &model.ShiftAgentSession{
			ID:              "existing-session-123",
			AssigneeID:      "user-1",
			ShiftInstanceID: "shift-1",
			Status:          "ACTIVE",
		}

		mockSessionRepo := &MockShiftAgentSessionRepository{
			FindByShiftFunc: func(ctx context.Context, assigneeID, shiftInstanceID string) (*model.ShiftAgentSession, error) {
				assert.Equal(t, "user-1", assigneeID)
				assert.Equal(t, "shift-1", shiftInstanceID)
				return existingSession, nil
			},
		}
		mockUserRepo := &MockUserRepository{}

		svc := NewShiftService(mockSessionRepo, mockUserRepo)
		session, err := svc.InitializeShift(context.Background(), "user-1", "shift-1")

		assert.NoError(t, err)
		assert.Equal(t, existingSession, session)
	})

	t.Run("creates a new session successfully", func(t *testing.T) {
		mockSessionRepo := &MockShiftAgentSessionRepository{
			FindByShiftFunc: func(ctx context.Context, assigneeID, shiftInstanceID string) (*model.ShiftAgentSession, error) {
				return nil, errors.New("not found")
			},
			CreateFunc: func(ctx context.Context, s *model.ShiftAgentSession) error {
				assert.Equal(t, "user-1", s.AssigneeID)
				assert.Equal(t, "shift-1", s.ShiftInstanceID)
				assert.Equal(t, "ACTIVE", s.Status)
				assert.Contains(t, string(s.SessionContext), "user-1@example.com")
				return nil
			},
		}

		mockUserRepo := &MockUserRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
				assert.Equal(t, "user-1", id)
				return &model.User{
					ID:    "user-1",
					Email: "user-1@example.com",
					Name:  "John Doe",
				}, nil
			},
		}

		svc := NewShiftService(mockSessionRepo, mockUserRepo)
		session, err := svc.InitializeShift(context.Background(), "user-1", "shift-1")

		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "user-1", session.AssigneeID)
		assert.Equal(t, "shift-1", session.ShiftInstanceID)
	})

	t.Run("user repo error", func(t *testing.T) {
		mockSessionRepo := &MockShiftAgentSessionRepository{
			FindByShiftFunc: func(ctx context.Context, assigneeID, shiftInstanceID string) (*model.ShiftAgentSession, error) {
				return nil, errors.New("not found")
			},
		}

		mockUserRepo := &MockUserRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
				return nil, errors.New("user not found database error")
			},
		}

		svc := NewShiftService(mockSessionRepo, mockUserRepo)
		session, err := svc.InitializeShift(context.Background(), "user-1", "shift-1")

		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "user not found database error")
	})
}
