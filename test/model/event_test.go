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
	"testing"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestEventStyle_Values(t *testing.T) {
	tests := []struct {
		name     string
		style    model.EventStyle
		expected string
	}{
		{"Style Batch", model.StyleBatch, "BATCH"},
		{"Style Adhoc", model.StyleAdhoc, "ADHOC"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.style))
		})
	}
}

func TestEventType_Values(t *testing.T) {
	tests := []struct {
		name     string
		enumVal  model.EventType
		expected string
	}{
		{"Retail Shift", model.EventRetailShift, "RetailShift"},
		{"Store Break", model.EventStoreBreak, "StoreBreak"},
		{"Training Session", model.EventTrainingSession, "TrainingSession"},
		{"Store Open", model.EventStoreOpen, "StoreOpenEvent"},
		{"Store Close", model.EventStoreClose, "StoreCloseEvent"},
		{"Till Drawer Drop", model.EventTillDrawerDrop, "TillDrawerDropEvent"},
		{"Register Audit", model.EventRegisterAudit, "RegisterAuditEvent"},
		{"Curbside Pickup", model.EventCurbsidePickup, "CurbsidePickupEvent"},
		{"Home Delivery", model.EventHomeDelivery, "HomeDeliveryEvent"},
		{"Return Processing", model.EventReturnProcess, "ReturnProcessEvent"},
		{"Receiving Arrival", model.EventReceivingArrival, "ReceivingArrivalEvent"},
		{"Shelving Stock", model.EventShelvingStock, "ShelvingStockEvent"},
		{"Stockout Resolution", model.EventStockoutCorrect, "StockoutCorrectEvent"},
		{"Inventory Count", model.EventInventoryCount, "InventoryCountEvent"},
		{"Price Markdown", model.EventPriceMarkdown, "PriceMarkdownEvent"},
		{"Customer Assistance", model.EventCustomerAssistance, "CustomerAssistanceEvent"},
		{"Asset Maintenance", model.EventAssetMaintenance, "AssetMaintenanceEvent"},
		{"Volt & Vine Showroom Refresh", model.EventShowroomRefresh, "ShowroomRefreshEvent"},
		{"Volt & Vine White Glove Delivery", model.EventWhiteGloveDispatch, "WhiteGloveDispatchEvent"},
		{"Volt & Vine Appliance Demo", model.EventApplianceDemo, "ApplianceDemoEvent"},
		{"OmniMart Perishable Freshness", model.EventPerishableFreshness, "PerishableFreshnessEvent"},
		{"OmniMart Prepared Hot Deli", model.EventHotFoodTransition, "HotFoodTransitionEvent"},
		{"OmniMart Direct Store Vendor", model.EventDirectStoreDelivery, "DirectStoreDeliveryEvent"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.enumVal))
		})
	}
}

func TestEventStatus_Values(t *testing.T) {
	tests := []struct {
		name     string
		status   model.EventStatus
		expected string
	}{
		{"Scheduled", model.StatusScheduled, "EventScheduled"},
		{"Cancelled", model.StatusCancelled, "EventCancelled"},
		{"Postponed", model.StatusPostponed, "EventPostponed"},
		{"Rescheduled", model.StatusRescheduled, "EventRescheduled"},
		{"Active", model.StatusActive, "EventActive"},
		{"Completed", model.StatusCompleted, "EventCompleted"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.status))
		})
	}
}

func TestEvent_Instantiation(t *testing.T) {
	siteID := "site-123"
	taskID := "task-456"

	eventBatch := model.Event{
		ID:          "event-123",
		OrganizerID: "user-organizer",
		SiteID:      &siteID,
		TaskID:      &taskID,
		Name:        "Receiving Pipeline AM",
		EventType:   model.EventReceivingArrival,
		EventStyle:  model.StyleBatch,
	}

	assert.Equal(t, "event-123", eventBatch.ID)
	assert.Equal(t, "user-organizer", eventBatch.OrganizerID)
	assert.Equal(t, "Receiving Pipeline AM", eventBatch.Name)
	assert.Equal(t, model.EventReceivingArrival, eventBatch.EventType)
	assert.Equal(t, model.StyleBatch, eventBatch.EventStyle)
	assert.Equal(t, "BATCH", string(eventBatch.EventStyle))

	eventAdhoc := model.Event{
		ID:          "event-789",
		OrganizerID: "sensor-alert-stockout",
		Name:        "Instant Aisle 4 Shelf 2 Replenish Alert",
		EventType:   model.EventStockoutCorrect,
		EventStyle:  model.StyleAdhoc,
	}
	assert.Equal(t, "event-789", eventAdhoc.ID)
	assert.Equal(t, model.StyleAdhoc, eventAdhoc.EventStyle)
	assert.Equal(t, "ADHOC", string(eventAdhoc.EventStyle))
}

func TestUserEventInstance_Instantiation(t *testing.T) {
	instance := model.UserEventInstance{
		ID:          "instance-123",
		ScheduleID:  "schedule-456",
		EventStatus: model.StatusActive,
	}

	assert.Equal(t, "instance-123", instance.ID)
	assert.Equal(t, "schedule-456", instance.ScheduleID)
	assert.Equal(t, model.StatusActive, instance.EventStatus)
	assert.Equal(t, "EventActive", string(instance.EventStatus))
}
