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

package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

// AfterUpdate is a standard GORM hook triggered automatically after updating a TaskExecution record.
// It extracts the active system user from the transactional context and inserts an explicit audit trail record.
func (e *TaskExecution) AfterUpdate(tx *gorm.DB) (err error) {
	var changedByID *string
	if val := tx.Statement.Context.Value("userID"); val != nil {
		if uid, ok := val.(string); ok && uid != "" {
			changedByID = &uid
		}
	}

	// Serializes the new state of the execution
	newStateBytes, err := json.Marshal(e)
	if err != nil {
		return err
	}

	audit := TaskExecutionAudit{
		TaskExecutionID: e.ID,
		ChangedByID:     changedByID,
		ActionType:      "STATUS_TRANSITION",
		NewState:        JSONB(newStateBytes),
	}

	// Persist within the same GORM transaction context
	return tx.Create(&audit).Error
}
