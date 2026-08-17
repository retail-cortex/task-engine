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

package model_test

import (
	"context"
	"testing"
	"time"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type dummyDialector struct{}

func (d dummyDialector) Name() string                                                { return "dummy" }
func (d dummyDialector) Initialize(db *gorm.DB) error                                { return nil }
func (d dummyDialector) Migrator(db *gorm.DB) gorm.Migrator                          { return nil }
func (d dummyDialector) DataTypeOf(field *schema.Field) string                       { return "text" }
func (d dummyDialector) DefaultValueOf(field *schema.Field) clause.Expression        { return nil }
func (d dummyDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {}
func (d dummyDialector) QuoteTo(writer clause.Writer, str string)                    { writer.WriteString(str) }
func (d dummyDialector) Explain(sql string, vars ...interface{}) string              { return sql }

func TestTaskExecution_AfterUpdate(t *testing.T) {
	db, err := gorm.Open(dummyDialector{}, &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	now := time.Now()

	tests := []struct {
		name     string
		exec     *model.TaskExecution
		withUser string
	}{
		{
			name: "TASK_STARTED transition",
			exec: &model.TaskExecution{
				ID:                 "exec-1",
				Status:             "IN_PROGRESS",
				TotalPausedSeconds: 0,
			},
			withUser: "user-123",
		},
		{
			name: "TASK_PAUSED transition",
			exec: &model.TaskExecution{
				ID:       "exec-2",
				Status:   "PAUSED",
				PausedAt: &now,
			},
		},
		{
			name: "TASK_RESUMED transition",
			exec: &model.TaskExecution{
				ID:                 "exec-3",
				Status:             "IN_PROGRESS",
				TotalPausedSeconds: 60,
			},
		},
		{
			name: "TASK_COMPLETED transition",
			exec: &model.TaskExecution{
				ID:     "exec-4",
				Status: "COMPLETED",
			},
		},
		{
			name: "STATUS_TRANSITION default",
			exec: &model.TaskExecution{
				ID:     "exec-5",
				Status: "ASSIGNED",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx := db
			if tc.withUser != "" {
				ctx := context.WithValue(context.Background(), "userID", tc.withUser)
				tx = db.WithContext(ctx)
			}
			err := tc.exec.AfterUpdate(tx)
			assert.NoError(t, err)
		})
	}
}
