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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/stretchr/testify/assert"
)

// Helper to probe and load relative SQL scripts hermetically under Bazel sandbox.
func loadSeedScript(t *testing.T, filename string) string {
	candidates := []string{
		filepath.Join("scripts", filename),
		filepath.Join("..", "scripts", filename),
		filepath.Join("..", "..", "scripts", filename),
		filepath.Join("..", "..", "..", "scripts", filename),
	}

	// Try resolving using bazel runfiles path structure
	if runfilesDir := os.Getenv("RUNFILES_DIR"); runfilesDir != "" {
		candidates = append(candidates, filepath.Join(runfilesDir, "_main", "scripts", filename))
		candidates = append(candidates, filepath.Join(runfilesDir, "scripts", filename))
	}

	for _, p := range candidates {
		content, err := os.ReadFile(p)
		if err == nil {
			return string(content)
		}
	}

	t.Fatalf("Failed to locate database seed script %q. Checked paths: %v", filename, candidates)
	return ""
}

func TestSeedsRelationalIntegrity(t *testing.T) {
	eventsContent := loadSeedScript(t, "dev_events.sql")
	assert.NotEmpty(t, eventsContent)

	// RegEx mapping standard hex-uuid formats: e.g. '88888888-8888-8888-8888-888888880001'
	uuidRegex := regexp.MustCompile(`'[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}'`)

	t.Run("assert all seeded UUIDs are well-formed", func(t *testing.T) {
		uuids := uuidRegex.FindAllString(eventsContent, -1)
		assert.NotEmpty(t, uuids)
		for _, raw := range uuids {
			clean := strings.Trim(raw, "'")
			assert.Len(t, clean, 36)
			// Checks standard hex layout
			assert.Regexp(t, "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$", clean)
		}
	})

	t.Run("validate seeded EventType enums exist and match compiler constants", func(t *testing.T) {
		// Isolate standard 'events' insert block context to prevent greedy matches on other tables
		eventsStart := strings.Index(eventsContent, "INSERT INTO events")
		if eventsStart == -1 {
			t.Fatalf("Failed to locate 'INSERT INTO events' block in dev_events.sql")
		}
		eventsSub := eventsContent[eventsStart:]
		eventsEnd := strings.Index(eventsSub, ";")
		if eventsEnd == -1 {
			t.Fatalf("Failed to locate termination semicolon for 'INSERT INTO events' block")
		}
		eventsInsertBlock := eventsSub[:eventsEnd]

		// EventType values extraction matching 'StoreOpenEvent', 'CurbsidePickupEvent' etc.
		// matches inserts in the form: ('...', '...', '...', NULL/ID, '...', 'EventTypeString', 'StyleString', NOW())
		eventRowRegex := regexp.MustCompile(`\(\s*'[^']+'\s*,\s*'[^']+'\s*,\s*'[^']+'\s*,\s*(?:'[^']+'|NULL)\s*,\s*'[^']+'\s*,\s*'([^']+)'\s*,\s*'([^']+)'\s*,\s*NOW\(\)\s*\)`)
		rows := eventRowRegex.FindAllStringSubmatch(eventsInsertBlock, -1)
		assert.NotEmpty(t, rows)

		// Create validation lookup set containing all standard defined model EventTypes
		validEventTypes := map[model.EventType]bool{
			model.EventRetailShift:         true,
			model.EventStoreBreak:          true,
			model.EventTrainingSession:     true,
			model.EventStoreOpen:           true,
			model.EventStoreClose:          true,
			model.EventTillDrawerDrop:      true,
			model.EventRegisterAudit:       true,
			model.EventCurbsidePickup:      true,
			model.EventHomeDelivery:        true,
			model.EventReturnProcess:       true,
			model.EventReceivingArrival:    true,
			model.EventShelvingStock:       true,
			model.EventStockoutCorrect:     true,
			model.EventInventoryCount:      true,
			model.EventPriceMarkdown:       true,
			model.EventCustomerAssistance:  true,
			model.EventAssetMaintenance:    true,
			model.EventShowroomRefresh:     true,
			model.EventWhiteGloveDispatch:  true,
			model.EventApplianceDemo:       true,
			model.EventPerishableFreshness: true,
			model.EventHotFoodTransition:   true,
			model.EventDirectStoreDelivery: true,
		}

		// Validation lookup set containing all standard model EventStyles
		validEventStyles := map[model.EventStyle]bool{
			model.StyleBatch: true,
			model.StyleAdhoc: true,
		}

		for _, match := range rows {
			assert.Len(t, match, 3) // index 0: full row, index 1: event_type, index 2: event_style
			dbEventType := model.EventType(match[1])
			dbEventStyle := model.EventStyle(match[2])

			// Assert both constants compile and match structural type specs without drift
			assert.True(t, validEventTypes[dbEventType], "Database seeds contain unrecognized EventType: %q", match[1])
			assert.True(t, validEventStyles[dbEventStyle], "Database seeds contain unrecognized EventStyle: %q", match[2])
		}
	})

	t.Run("validate materialized shift occurrences map valid EventStatus enums", func(t *testing.T) {
		// Isolate standard 'user_event_instances' insert block context
		instancesStart := strings.Index(eventsContent, "INSERT INTO user_event_instances")
		if instancesStart == -1 {
			t.Fatalf("Failed to locate 'INSERT INTO user_event_instances' block in dev_events.sql")
		}
		instancesSub := eventsContent[instancesStart:]
		instancesEnd := strings.Index(instancesSub, ";")
		if instancesEnd == -1 {
			t.Fatalf("Failed to locate termination semicolon for 'INSERT INTO user_event_instances' block")
		}
		instancesInsertBlock := instancesSub[:instancesEnd]

		// Materialized occurrence inserts map in the form: ('...', '...', '...', '...', 'EventStatusString', NOW())
		instanceRowRegex := regexp.MustCompile(`\(\s*'[^']+'\s*,\s*'[^']+'\s*,\s*'[^']+'\s*,\s*'[^']+'\s*,\s*'([^']+)'\s*,\s*NOW\(\)\s*\)`)
		rows := instanceRowRegex.FindAllStringSubmatch(instancesInsertBlock, -1)
		assert.NotEmpty(t, rows)

		validStatuses := map[model.EventStatus]bool{
			model.StatusScheduled:   true,
			model.StatusCancelled:   true,
			model.StatusPostponed:   true,
			model.StatusRescheduled: true,
			model.StatusActive:      true,
			model.StatusCompleted:   true,
		}

		for _, match := range rows {
			assert.Len(t, match, 2) // index 1: status string
			dbStatus := model.EventStatus(match[1])
			assert.True(t, validStatuses[dbStatus], "Database seeds contain unrecognized EventStatus state: %q", match[1])
		}
	})

	t.Run("validate seeded SOP profiles are well-formed", func(t *testing.T) {
		sopsStart := strings.Index(eventsContent, "INSERT INTO sops")
		if sopsStart == -1 {
			t.Fatalf("Failed to locate 'INSERT INTO sops' block in dev_events.sql")
		}
		sopsSub := eventsContent[sopsStart:]
		sopsEnd := strings.Index(sopsSub, ";")
		if sopsEnd == -1 {
			t.Fatalf("Failed to locate termination semicolon for 'INSERT INTO sops' block")
		}
		sopsInsertBlock := sopsSub[:sopsEnd]

		// Extract SOP URLs and titles
		// Mappings: ('id', 'TitleString', 'URLString', '{}', NOW())
		sopRowRegex := regexp.MustCompile(`\(\s*'[^']+'\s*,\s*'([^']+)'\s*,\s*'([^']+)'\s*,\s*'[^']+'\s*,\s*NOW\(\)\s*\)`)
		rows := sopRowRegex.FindAllStringSubmatch(sopsInsertBlock, -1)
		assert.NotEmpty(t, rows)

		for _, match := range rows {
			assert.Len(t, match, 3) // 1: Title, 2: URL
			assert.NotEmpty(t, match[1])
			assert.True(t, strings.HasPrefix(match[2], "http://") || strings.HasPrefix(match[2], "https://"), "SOP URL must use http/https protocols: %q", match[2])
			assert.Contains(t, match[2], "/static/sops/", "SOP URL must point to standard static assets: %q", match[2])
		}
	})
}
