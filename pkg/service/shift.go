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
	"encoding/json"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
)

// ShiftService manages employee shifts and agent session initialization.
type ShiftService interface {
	InitializeShift(ctx context.Context, userID, shiftInstanceID string) (*model.ShiftAgentSession, error)
	UpdateSession(ctx context.Context, session *model.ShiftAgentSession) error
	ListActiveUsers(ctx context.Context) ([]*model.User, error)
	ListActiveOnShiftUsers(ctx context.Context, siteID string) ([]*model.User, error)
	GetUserProfile(ctx context.Context, userID string) (*model.User, error)
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

func (s *shiftService) UpdateSession(ctx context.Context, session *model.ShiftAgentSession) error {
	return s.sessionRepo.Update(ctx, session)
}

func (s *shiftService) ListActiveUsers(ctx context.Context) ([]*model.User, error) {
	return s.userRepo.List(ctx)
}

func (s *shiftService) ListActiveOnShiftUsers(ctx context.Context, siteID string) ([]*model.User, error) {
	return s.userRepo.ListActiveOnShiftUsers(ctx, siteID)
}


func (s *shiftService) GetUserProfile(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}
