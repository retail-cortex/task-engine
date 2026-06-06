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
	"time"
)

// EventStyle maps how operational events are executed in the automation engine, distinguishing scheduled windows from dynamic triggers.
type EventStyle string

const (
	// StyleBatch events are scheduled systematically around a 24-hour window, matching standard employee shifts.
	StyleBatch EventStyle = "BATCH"

	// StyleAdhoc events are triggered streaming, dynamically in response to on-demand triggers, alarms, or sensors.
	StyleAdhoc EventStyle = "ADHOC"
)

// EventType represents the operational retail category of the event, aligned with ARTS and Schema.org.
type EventType string

const (
	// Shift & Workforce Events
	EventRetailShift     EventType = "RetailShift"
	EventStoreBreak      EventType = "StoreBreak"
	EventTrainingSession EventType = "TrainingSession"

	// Store & Cashier Core Operations (ARTS Business Day Open/Close & Till Drops / Schema.org: BusinessEvent / Action)
	EventStoreOpen      EventType = "StoreOpenEvent"       // vault drops, cashier startups, physical door unlocking (Batch)
	EventStoreClose     EventType = "StoreCloseEvent"      // security sweeps, cash drops, door locking registers off (Batch)
	EventTillDrawerDrop EventType = "TillDrawerDropEvent"  // cash drawer limit reached, requesting cash drop transfer (Adhoc)
	EventRegisterAudit  EventType = "RegisterAuditEvent"   // scheduled till reconciliations, manager audit verification (Batch/Adhoc)

	// Fulfillment & Customer Action Operations (Schema.org: DeliveryEvent / OrderAction / ReturnAction)
	EventCurbsidePickup EventType = "CurbsidePickupEvent" // drive-up, click-and-collect fulfillment workload (Batch)
	EventHomeDelivery   EventType = "HomeDeliveryEvent"   // dispatching localized deliveries (Batch)
	EventReturnProcess  EventType = "ReturnProcessEvent"  // parsing customer product returns (Adhoc)

	// Inventory & Logistical Operations (ARTS Inventory schemas / Schema.org: ReceiveAction / OrganizeAction / UpdateAction)
	EventReceivingArrival EventType = "ReceivingArrivalEvent" // delivery trucks arriving at docks (Batch)
	EventShelvingStock    EventType = "ShelvingStockEvent"    // shelf replenishments / backroom transfers (Batch)
	EventStockoutCorrect  EventType = "StockoutCorrectEvent"  // empty shelf replenishment alerts (Adhoc)
	EventInventoryCount   EventType = "InventoryCountEvent"   // standard inventory auditing count events (Batch)
	EventPriceMarkdown    EventType = "PriceMarkdownEvent"    // pricing label auditing and markdowns (Batch)

	// Support & Maintenance Operations (ARTS Store Environment / Schema.org: ContactAction / ControlAction)
	EventCustomerAssistance EventType = "CustomerAssistanceEvent" // on-demand register/department customer support (Adhoc)
	EventAssetMaintenance   EventType = "AssetMaintenanceEvent"   // equipment liveness, sensor calibrations (Batch/Adhoc)

	// Brand-Specific Events: Volt & Vine (Smart Appliances Showcase - Premium Showrooms)
	EventShowroomRefresh   EventType = "ShowroomRefreshEvent"   // display appliance swaps, digital showcase calibration (Batch)
	EventWhiteGloveDispatch EventType = "WhiteGloveDispatchEvent" // delivery preparation & manifest checks for luxury setups (Batch)
	EventApplianceDemo      EventType = "ApplianceDemoEvent"      // customer testing slots for premium cooktops/ovens (Batch/Adhoc)

	// Brand-Specific Events: OmniMart (Hypermarket Grocery & General Retail - Fast Ingest)
	EventPerishableFreshness EventType = "PerishableFreshnessEvent" // high-frequency produce rotations & temp logging (Batch)
	EventHotFoodTransition  EventType = "HotFoodTransitionEvent"  // deli counter swaps (breakfast, lunch, dinner setups) (Batch)
	EventDirectStoreDelivery EventType = "DirectStoreDeliveryEvent" // streaming dock arrivals (bread, soda direct vendors) (Adhoc)
)

// EventStatus defines standard lifecycle states of scheduled operational events (Schema.org / ARTS).
type EventStatus string

const (
	// Schema.org standard EventStatusType values
	StatusScheduled   EventStatus = "EventScheduled"
	StatusCancelled   EventStatus = "EventCancelled"
	StatusPostponed   EventStatus = "EventPostponed"
	StatusRescheduled EventStatus = "EventRescheduled"

	// Operational runtime extensions
	StatusActive    EventStatus = "EventActive"
	StatusCompleted EventStatus = "EventCompleted"
)

// Event defines a scheduled operational target or workload boundaries.
type Event struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizerID string     `gorm:"type:uuid;not null;index"`
	SiteID      *string    `gorm:"type:uuid;index;default:null"`
	TaskID      *string    `gorm:"type:uuid;index;default:null"`
	Name        string     `gorm:"type:varchar(255);not null"`
	EventType   EventType  `gorm:"type:varchar(100);not null;default:'RetailShift'"`
	EventStyle  EventStyle `gorm:"type:varchar(20);not null;default:'BATCH'"` // BATCH or ADHOC mapping
	CreatedAt   time.Time  `gorm:"not null;default:now()"`
}

// UserEventSchedule holds individual user recurring shift or workload schedules.
type UserEventSchedule struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `gorm:"type:uuid;not null;uniqueIndex:idx_user_event,priority:1"`
	EventID   string    `gorm:"type:uuid;not null;index;uniqueIndex:idx_user_event,priority:2"`
	StartDate time.Time `gorm:"not null"`
	EndDate   time.Time `gorm:"not null"`
	Timezone  string    `gorm:"type:varchar(50);not null"`
	Rrule     *string   `gorm:"type:text;default:null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

// UserEventInstance represents materialized instances of user schedules.
type UserEventInstance struct {
	ID                string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ScheduleID        string      `gorm:"type:uuid;not null;uniqueIndex:idx_schedule_instance,priority:1"`
	InstanceStartDate time.Time   `gorm:"not null;uniqueIndex:idx_schedule_instance,priority:2"`
	InstanceEndDate   time.Time   `gorm:"not null"`
	EventStatus       EventStatus `gorm:"type:varchar(50);not null;default:'EventScheduled';index"`
	CreatedAt         time.Time   `gorm:"not null;default:now()"`
}
