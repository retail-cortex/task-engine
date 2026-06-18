-- Copyright 2026 Google LLC
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

-- Dev Environment Retail Identities, Events and Shifts Seed Script
-- Programmatically maps robust operational and administrative datasets

-- TRUNCATE existing relational scheduling and RAG contexts
TRUNCATE TABLE sop_chunks CASCADE;
TRUNCATE TABLE sop_processes CASCADE;
TRUNCATE TABLE sops CASCADE;
TRUNCATE TABLE user_event_instances CASCADE;
TRUNCATE TABLE user_event_schedules CASCADE;
TRUNCATE TABLE events CASCADE;
TRUNCATE TABLE tasks CASCADE;

-- Clean specific mock user data (preserving the 553 provisioned Workspace directory profiles)
DELETE FROM user_certifications WHERE user_id IN ('00000000-0000-0000-0000-000000000000', '88888888-8888-8888-8888-888888880001', '88888888-8888-8888-8888-888888880002', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '88888888-8888-8888-8888-888888880003', '88888888-8888-8888-8888-888888880004');
DELETE FROM user_roles WHERE user_id IN ('00000000-0000-0000-0000-000000000000', '88888888-8888-8888-8888-888888880001', '88888888-8888-8888-8888-888888880002', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '88888888-8888-8888-8888-888888880003', '88888888-8888-8888-8888-888888880004');
DELETE FROM user_sites WHERE user_id IN ('00000000-0000-0000-0000-000000000000', '88888888-8888-8888-8888-888888880001', '88888888-8888-8888-8888-888888880002', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '88888888-8888-8888-8888-888888880003', '88888888-8888-8888-8888-888888880004');
DELETE FROM user_organizations WHERE user_id IN ('00000000-0000-0000-0000-000000000000', '88888888-8888-8888-8888-888888880001', '88888888-8888-8888-8888-888888880002', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '88888888-8888-8888-8888-888888880003', '88888888-8888-8888-8888-888888880004');
DELETE FROM users WHERE id IN ('00000000-0000-0000-0000-000000000000', '88888888-8888-8888-8888-888888880001', '88888888-8888-8888-8888-888888880002', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '88888888-8888-8888-8888-888888880003', '88888888-8888-8888-8888-888888880004');

-- 1. Seed Roles
INSERT INTO roles (id, name, description, created_at) VALUES
	('77777777-7777-7777-7777-777777770001', 'ADMIN', 'Global systems manager. Holds absolute access over all organization sites workloads.', NOW()),
	('77777777-7777-7777-7777-777777770002', 'REGION_MANAGER', 'Regional general operations manager. Maps dynamic multi-site selection lists.', NOW()),
	('77777777-7777-7777-7777-777777770003', 'SITE_MANAGER', 'Physical store site operational supervisor. Audits queues and reviews associate tasks.', NOW()),
	('77777777-7777-7777-7777-777777770004', 'SITE_ASSOCIATE', 'Standard store workload associate. Executes checklists and trades queue tasks.', NOW()),
	('77777777-7777-7777-7777-777777770005', 'SITE_3P', 'Third-party vendor logistics checker.', NOW())
ON CONFLICT (id) DO NOTHING;

-- 2. Seed Users
INSERT INTO users (id, o_auth_provider, o_auth_id, email, name, metadata, preferred_language_id, created_at, updated_at, version) VALUES
	-- Default Mock System User (Offline Bypass Sandbox!) -> SITE_MANAGER (Volt & Vine Seattle)
	('00000000-0000-0000-0000-000000000000', 'google', 'user-oauth-id-mock', 'hanna-mock@rmcguinness.altostrat.com', 'Hanna (Mock)', '{"title": "Volt & Vine Seattle Store Operations Manager"}', 'a0000000-0000-0000-0000-000000000001', NOW(), NOW(), 1),
	-- Floor Associate (Hanna) -> SITE_ASSOCIATE
	('88888888-8888-8888-8888-888888880001', 'google', 'user-oauth-id-hanna', 'hanna@rmcguinness.altostrat.com', 'Hanna', '{"title": "Senior Cashier & Replenishment Specialist"}', 'a0000000-0000-0000-0000-000000000001', NOW(), NOW(), 1),
	-- Shift Supervisor (Ryan) -> ADMIN (Global Master Systems Administrator)
	('b75c1a02-c884-40ed-a3f8-8b95f3ff7539', 'google', 'user-oauth-id-ryan', 'ryan@rmcguinness.altostrat.com', 'Ryan', '{"title": "Corporate Master Systems Administrator"}', 'a0000000-0000-0000-0000-000000000001', NOW(), NOW(), 1),
	-- Coworker Associate (Jenna) -> SITE_MANAGER (OmniMart Dallas)
	('88888888-8888-8888-8888-888888880003', 'google', 'user-oauth-id-jenna', 'jenna@rmcguinness.altostrat.com', 'Jenna', '{"title": "OmniMart Dallas Store Operations Manager"}', 'a0000000-0000-0000-0000-000000000001', NOW(), NOW(), 1),
	-- Southwest Regional Director (Marcus) -> REGION_MANAGER
	('88888888-8888-8888-8888-888888880004', 'google', 'user-oauth-id-marcus', 'marcus@rmcguinness.altostrat.com', 'Marcus', '{"title": "OmniMart Southwest Regional Operations Director"}', 'a0000000-0000-0000-0000-000000000002', NOW(), NOW(), 1);

-- 3. Map Users to Roles
INSERT INTO user_roles (user_id, role_id) VALUES
	('00000000-0000-0000-0000-000000000000', '77777777-7777-7777-7777-777777770001'), -- Mock User (ADMIN)
	('88888888-8888-8888-8888-888888880001', '77777777-7777-7777-7777-777777770004'), -- Hanna (SITE_ASSOCIATE)
	('b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '77777777-7777-7777-7777-777777770001'), -- Ryan (ADMIN)
	('88888888-8888-8888-8888-888888880003', '77777777-7777-7777-7777-777777770003'), -- Jenna (SITE_MANAGER)
	('88888888-8888-8888-8888-888888880004', '77777777-7777-7777-7777-777777770002'); -- Marcus (REGION_MANAGER)

-- 4. Map Users to Organizations
INSERT INTO user_organizations (organization_id, user_id) VALUES
	-- All map to the parent Gemini Nexus corporate boundary
	('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000000'),
	('11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888880001'),
	('11111111-1111-1111-1111-111111111111', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539'),
	('11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888880003'),
	('11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888880004'),
	-- Hanna, Mock, Marcus map to OmniMart (Grocery segment)
	('33333333-3333-3333-3333-333333333333', '00000000-0000-0000-0000-000000000000'),
	('33333333-3333-3333-3333-333333333333', '88888888-8888-8888-8888-888888880001'),
	('33333333-3333-3333-3333-333333333333', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539'), -- Ryan also maps to OmniMart
	('33333333-3333-3333-3333-333333333333', '88888888-8888-8888-8888-888888880003'), -- Jenna also maps to OmniMart
	('33333333-3333-3333-3333-333333333333', '88888888-8888-8888-8888-888888880004'), -- Marcus maps to OmniMart
	-- Ryan, Jenna, Marcus map to Volt & Vine (Luxury Showcase segment)
	('22222222-2222-2222-2222-222222222222', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539'),
	('22222222-2222-2222-2222-222222222222', '88888888-8888-8888-8888-888888880003'),
	('22222222-2222-2222-2222-222222222222', '88888888-8888-8888-8888-888888880004');

-- 5. Map Users to Primary Sites
INSERT INTO user_sites (id, user_id, site_id, is_primary, metadata, created_at) VALUES
	-- Mock maps to Volt & Vine Seattle (Store #2005 / Primary Manager)
	('88888888-8888-8888-8888-777777770000', '00000000-0000-0000-0000-000000000000', '44444444-4444-4444-4444-444444440000', TRUE, '{}', NOW()),
	-- Hanna maps to OmniMart Dallas (Store #1000)
	('88888888-8888-8888-8888-777777770001', '88888888-8888-8888-8888-888888880001', '55555555-5555-5555-5555-555555550000', TRUE, '{}', NOW()),
	-- Ryan maps to Volt & Vine Seattle (Primary) and OmniMart Dallas (Secondary)
	('88888888-8888-8888-8888-777777770002', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '44444444-4444-4444-4444-444444440000', TRUE, '{}', NOW()),
	('88888888-8888-8888-8888-777777770003', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '55555555-5555-5555-5555-555555550000', FALSE, '{}', NOW()),
	-- Jenna maps to OmniMart Dallas (Primary Manager) and Volt & Vine Seattle (Secondary)
	('88888888-8888-8888-8888-777777770004', '88888888-8888-8888-8888-888888880003', '55555555-5555-5555-5555-555555550000', TRUE, '{}', NOW()),
	('88888888-8888-8888-8888-777777770005', '88888888-8888-8888-8888-888888880003', '44444444-4444-4444-4444-444444440000', FALSE, '{}', NOW()),
	-- Marcus maps to OmniMart Dallas (Primary) and Volt & Vine Seattle (Secondary)
	('88888888-8888-8888-8888-777777770006', '88888888-8888-8888-8888-888888880004', '55555555-5555-5555-5555-555555550000', TRUE, '{}', NOW()),
	('88888888-8888-8888-8888-777777770007', '88888888-8888-8888-8888-888888880004', '44444444-4444-4444-4444-444444440000', FALSE, '{}', NOW());

-- 6. Seed Operational Events
-- Maps workload profiles matching standard operations across brand networks
INSERT INTO events (id, organizer_id, site_id, task_id, name, event_type, event_style, created_at) VALUES
	-- A. Global & Cashier Core Operations (OmniMart Dallas)
	('99999999-9999-9999-9999-999999990001', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '55555555-5555-5555-5555-555555550000', NULL, 'OmniMart Dallas - Morning Store Opening', 'StoreOpenEvent', 'BATCH', NOW()),
	('99999999-9999-9999-9999-999999990002', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '55555555-5555-5555-5555-555555550000', NULL, 'OmniMart Dallas - Evening Store Closing', 'StoreCloseEvent', 'BATCH', NOW()),
	('99999999-9999-9999-9999-999999990003', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '55555555-5555-5555-5555-555555550000', NULL, 'OmniMart Dallas - Front Registry Audit Sweep', 'RegisterAuditEvent', 'BATCH', NOW()),

	-- B. Brand Specific: OmniMart (Grocery Hypermarket)
	('99999999-9999-9999-9999-999999990011', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '55555555-5555-5555-5555-555555550000', NULL, 'OmniMart Dallas - Fresh Produce Rotation Audit', 'PerishableFreshnessEvent', 'BATCH', NOW()),
	('99999999-9999-9999-9999-999999990012', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '55555555-5555-5555-5555-555555550000', NULL, 'OmniMart Dallas - Kitchen Breakfast-to-Lunch Transition', 'HotFoodTransitionEvent', 'BATCH', NOW()),
	('99999999-9999-9999-9999-999999990013', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '55555555-5555-5555-5555-555555550000', NULL, 'OmniMart Dallas - Curbside Delivery Pickup Run', 'CurbsidePickupEvent', 'BATCH', NOW()),

	-- C. Brand Specific: Volt & Vine (Smart Appliance Showrooms)
	('99999999-9999-9999-9999-999999990021', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '44444444-4444-4444-4444-444444440000', NULL, 'Volt & Vine Seattle - Premium Showcase Swap', 'ShowroomRefreshEvent', 'BATCH', NOW()),
	('99999999-9999-9999-9999-999999990022', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '44444444-4444-4444-4444-444444440000', NULL, 'Volt & Vine Seattle - Luxury Home Delivery Load', 'WhiteGloveDispatchEvent', 'BATCH', NOW()),

	-- D. Base Workforce Shift Events
	-- Associates shifts are registered as Events themselves under GORM scheduling loops
	('99999999-9999-9999-9999-999999990099', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '55555555-5555-5555-5555-555555550000', NULL, 'Hanna Scheduled Shift - Associate Lane Coverage', 'RetailShift', 'BATCH', NOW()),
	('99999999-9999-9999-9999-999999990100', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '44444444-4444-4444-4444-444444440000', NULL, 'Ryan Scheduled Shift - Showroom Supervisor Shift', 'RetailShift', 'BATCH', NOW()),
	-- Mock Streaming Alert Container Events
	('00000000-0000-0000-0000-eeeeeeeeeeee', '00000000-0000-0000-0000-000000000000', '44444444-4444-4444-4444-444444440000', NULL, 'Seattle Streaming Alert Container Event', 'RetailShift', 'BATCH', NOW());

-- 7. Seed workforce recurrence schedules
INSERT INTO user_event_schedules (id, user_id, event_id, start_date, end_date, timezone, rrule, created_at) VALUES
	-- Hanna (Associate) Shift Recurrence: Mon-Fri 08:00 to 16:00America/Chicago timezone
	('55555555-5555-7777-0000-000000000001', '88888888-8888-8888-8888-888888880001', '99999999-9999-9999-9999-999999990099', '2026-05-18 08:00:00-05:00', '2026-05-22 16:00:00-05:00', 'America/Chicago', 'FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR', NOW()),
	-- Ryan (Supervisor) Shift Recurrence: Mon-Fri 09:00 to 17:00 America/Los_Angeles timezone
	('55555555-5555-7777-0000-000000000002', 'b75c1a02-c884-40ed-a3f8-8b95f3ff7539', '99999999-9999-9999-9999-999999990100', '2026-05-18 09:00:00-07:00', '2026-05-22 17:00:00-07:00', 'America/Los_Angeles', 'FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR', NOW()),
	-- Mock Streaming Alert Container Recurrence
	('00000000-0000-0000-0000-dddddddddddd', '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-eeeeeeeeeeee', '2026-05-18 00:00:00-00:00', '2026-05-25 00:00:00-00:00', 'UTC', NULL, NOW());

-- 8. Seed Materialized Event Instances
-- Materializes live shifts spanning the baseline week
INSERT INTO user_event_instances (id, schedule_id, instance_start_date, instance_end_date, event_status, created_at) VALUES
	-- Hanna's active shift occurrence context map
	('00000000-0000-0000-0000-000000000001', '55555555-5555-7777-0000-000000000001', '2026-05-22 08:00:00-05:00', '2026-05-22 16:00:00-05:00', 'EventActive', NOW()),
	-- Ryan's active shift occurrence context map
	('00000000-0000-0000-0000-000000000002', '55555555-5555-7777-0000-000000000002', '2026-05-22 09:00:00-07:00', '2026-05-22 17:00:00-07:00', 'EventActive', NOW()),
	-- Mock Streaming Alert Container Instance
	('00000000-0000-0000-0000-ffffffffffff', '00000000-0000-0000-0000-dddddddddddd', '2026-05-18 00:00:00-00:00', '2026-05-25 00:00:00-00:00', 'EventActive', NOW());

-- 9. Seed Standard Operating Procedures (SOPs)
-- Maps to local static mock assets served on local:8080 during local development
INSERT INTO sops (id, title, canonical_url, metadata, created_at) VALUES
	('11111111-aaaa-bbbb-cccc-999999990001', 'Produce Freshness and Rotation SOP', 'http://localhost:8080/static/sops/produce_freshness.html', '{}', NOW()),
	('11111111-aaaa-bbbb-cccc-999999990002', 'Vault Cash Audit Compliance Guidelines SOP', 'http://localhost:8080/static/sops/vault_audit.pdf', '{}', NOW());

-- 10. Seed Dynamic Task Templates
INSERT INTO tasks (id, name, description, task_type, priority, step_order, estimated_duration_minutes, checklist_template, metadata, created_at, updated_at, version) VALUES
	-- Template A: Front registers opening checks
	('d000fa44-0000-0000-0000-000000000000', 'Register Terminal & Cash Opening Checkout Suite', 'Mandatory cash drawer count validation and vault alignment routine.', 'STANDARD', 1, 0, 15, '[{"step": 1, "action": "Unlock terminal register drawers", "required": true}, {"step": 2, "action": "Verify cash vault count matches system drop records", "required": true}, {"step": 3, "action": "Verify receipt thermal roll status", "required": false}]', '{}', NOW(), NOW(), 1),
	
	-- Template B: Fresh perishables rotation and temperature scans
	('d000fa55-0000-0000-0000-000000000000', 'Produce Display Freshness & Rotation Sweep', 'Check freshness metrics on wet wall vegetables, cull damaged items, and log glycol temperature displays.', 'STANDARD', 3, 0, 20, '[{"step": 1, "action": "Cull wilted leafy greens and bruised fruits", "required": true}, {"step": 2, "action": "Log displays chiller temperatures (Limits: 34F to 38F)", "required": true}, {"step": 3, "action": "Rotate new stock crates behind older front-facing packages", "required": true}]', '{}', NOW(), NOW(), 1),
	
	-- Template C: Out-of-stock shelf items replenishment
	('d000fa66-0000-0000-0000-000000000000', 'Aisle 7 Empty Stockout Shelf Replenishment Run', 'Aisle 7 paper towels shelf reports stockout under spatial sensor grids. Replenish from backup inventory cage.', 'STANDARD', 2, 0, 15, '[{"step": 1, "action": "Locate pallet backup count in Receiving Cage B", "required": true}, {"step": 2, "action": "Use store jack to transport carton pallet to Aisle 7", "required": true}, {"step": 3, "action": "Stock display shelves and face front items flush", "required": true}, {"step": 4, "action": "Scan display barcode tag using handset to sign off", "required": true}]', '{}', NOW(), NOW(), 1);

-- 11. Seed Cash Drawer Asset for Ryan's Seattle Store
INSERT INTO assets (id, location_id, name, asset_tag, status, metadata, created_at, updated_at, version) VALUES
	('ca54ce11-0000-0000-0000-000000000004', '44444444-4444-4444-4444-444444440000', 'Register Terminal Cash Drawer 4', 'TAG-CASH-DRAWER-4', 'AVAILABLE', '{}', NOW(), NOW(), 1)
	ON CONFLICT (id) DO NOTHING;
