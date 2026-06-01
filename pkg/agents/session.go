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

package agents

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// AgentSessionState maps stateful ADK memory, planning states, and variables persistently to GORM tables.
type AgentSessionState struct {
	SessionID string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID   string    `gorm:"type:varchar(50);not null;index"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	Variables string    `gorm:"type:text;default:null"` // JSON serialised session variables (planning maps, telemetry metrics)
	History   string    `gorm:"type:text;default:null"` // JSON serialised message history list
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;index"`
}

// SessionService orchestrates state recovery, message logging, and persistent variables persistence for stateful ADK agents.
type SessionService interface {
	GetOrCreateSession(ctx context.Context, sessionID, agentID, userID string) (*AgentSessionState, error)
	SaveSession(ctx context.Context, session *AgentSessionState) error
	ClearSession(ctx context.Context, sessionID string) error
}

type sessionService struct {
	db *gorm.DB
}

// NewSessionService instantiates a self-bootstrapping Postgres-backed ADK SessionService.
// Automatically ensures dynamic GORM tables migrations are executed inside Postgres/AlloyDB on startup!
func NewSessionService(db *gorm.DB) (SessionService, error) {
	if err := db.AutoMigrate(&AgentSessionState{}); err != nil {
		return nil, err
	}
	return &sessionService{db: db}, nil
}

func (s *sessionService) GetOrCreateSession(ctx context.Context, sessionID, agentID, userID string) (*AgentSessionState, error) {
	if sessionID == "" || agentID == "" || userID == "" {
		return nil, errors.New("sessionID, agentID, and userID are mandatory parameters")
	}

	var sess AgentSessionState
	err := s.db.WithContext(ctx).First(&sess, "session_id = ?", sessionID).Error
	if err == nil && sess.SessionID != "" {
		return &sess, nil
	}

	// Session not found: materialize a new stateful session persistently!
	newSess := &AgentSessionState{
		SessionID: sessionID,
		AgentID:   agentID,
		UserID:    userID,
		Variables: "{}",
		History:   "[]",
	}

	if err := s.db.WithContext(ctx).Create(newSess).Error; err != nil {
		return nil, err
	}

	return newSess, nil
}

func (s *sessionService) SaveSession(ctx context.Context, session *AgentSessionState) error {
	return s.db.WithContext(ctx).Save(session).Error
}

func (s *sessionService) ClearSession(ctx context.Context, sessionID string) error {
	return s.db.WithContext(ctx).Delete(&AgentSessionState{}, "session_id = ?", sessionID).Error
}

type inMemorySessionService struct {
	sessions map[string]*AgentSessionState
}

// NewInMemorySessionService instantiates an in-memory, mock-compatible ADK SessionService for testing sandboxes.
func NewInMemorySessionService() SessionService {
	return &inMemorySessionService{sessions: make(map[string]*AgentSessionState)}
}

func (s *inMemorySessionService) GetOrCreateSession(ctx context.Context, sessionID, agentID, userID string) (*AgentSessionState, error) {
	if sessionID == "" || agentID == "" || userID == "" {
		return nil, errors.New("sessionID, agentID, and userID are mandatory parameters")
	}

	sess, ok := s.sessions[sessionID]
	if ok {
		return sess, nil
	}

	newSess := &AgentSessionState{
		SessionID: sessionID,
		AgentID:   agentID,
		UserID:    userID,
		Variables: "{}",
		History:   "[]",
	}
	s.sessions[sessionID] = newSess
	return newSess, nil
}

func (s *inMemorySessionService) SaveSession(ctx context.Context, session *AgentSessionState) error {
	s.sessions[session.SessionID] = session
	return nil
}

func (s *inMemorySessionService) ClearSession(ctx context.Context, sessionID string) error {
	delete(s.sessions, sessionID)
	return nil
}
