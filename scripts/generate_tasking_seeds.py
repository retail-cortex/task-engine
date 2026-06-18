#!/usr/bin/env python3
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Script to systematically generate tasking, event scheduling, and task execution records

for all 109 storefront test stores and their associates to populate standard seed SQL logs.
"""

import json
import os
import re
import uuid
from typing import Dict, List, Tuple, Any

# Define Paths
WORKSPACE_ROOT: str = "/Users/rmcguinness/Projects/internal/gemini_task_engine"
TFVARS_PATH: str = os.path.join(WORKSPACE_ROOT, "scripts/terraform/terraform.tfvars.json")
USERS_SEED_PATH: str = os.path.join(WORKSPACE_ROOT, "scripts/terraform/app_users_seed.sql")
OUTPUT_SQL_PATH: str = os.path.join(WORKSPACE_ROOT, "scripts/seed_store_tasks.sql")

# Static Global References (Matching dev_events.sql and models)
ROLE_SITE_ASSOCIATE: str = "77777777-7777-7777-7777-777777770004"
SYSTEM_ADMIN_USER_ID: str = "b75c1a02-c884-40ed-a3f8-8b95f3ff7539" # Ryan's Master Administrator ID

# Task Template Definitions
TEMPLATE_REGISTER_OPEN: str = "d000fa44-0000-0000-0000-000000000000"
TEMPLATE_PRODUCE_FRESHNESS: str = "d000fa55-0000-0000-0000-000000000000"
TEMPLATE_SHELF_REPLENISH: str = "d000fa66-0000-0000-0000-000000000000"
TEMPLATE_SHOWROOM_REFRESH: str = "d000fa77-0000-0000-0000-000000000000" # New Premium Showcase Template

# Checklist Step Templates
STEPS_REGISTER_OPEN: List[Dict[str, Any]] = [
    {"step": 1, "action": "Unlock terminal register drawers", "required": True},
    {"step": 2, "action": "Verify cash vault count matches system drop records", "required": True},
    {"step": 3, "action": "Verify receipt thermal roll status", "required": False}
]

STEPS_PRODUCE_FRESHNESS: List[Dict[str, Any]] = [
    {"step": 1, "action": "Cull wilted leafy greens and bruised fruits", "required": True},
    {"step": 2, "action": "Log displays chiller temperatures (Limits: 34F to 38F)", "required": True},
    {"step": 3, "action": "Rotate new stock crates behind older front-facing packages", "required": True}
]

STEPS_SHELF_REPLENISH: List[Dict[str, Any]] = [
    {"step": 1, "action": "Locate pallet backup count in Receiving Cage B", "required": True},
    {"step": 2, "action": "Use store jack to transport carton pallet to Aisle 7", "required": True},
    {"step": 3, "action": "Stock display shelves and face front items flush", "required": True},
    {"step": 4, "action": "Scan display barcode tag using handset to sign off", "required": True}
]

STEPS_SHOWROOM_REFRESH: List[Dict[str, Any]] = [
    {"step": 1, "action": "Wipe down display cooktops and refrigerators", "required": True},
    {"step": 2, "action": "Verify interactive tablet demos are online and responsive", "required": True},
    {"step": 3, "action": "Ensure price tags and feature sheets match current promos", "required": True}
]

def make_completed_checklist(steps: List[Dict[str, Any]], completed_timestamps: List[str]) -> str:
    res = []
    for idx, step in enumerate(steps):
        completed_at = completed_timestamps[idx] if idx < len(completed_timestamps) else "2026-06-02T08:22:14Z"
        started_at = "2026-06-02T08:00:00Z" if idx == 0 else completed_timestamps[idx-1]
        res.append({
            **step,
            "completed": True,
            "status": "COMPLETED",
            "started_at": started_at,
            "completed_at": completed_at,
            "total_paused_seconds": 0,
            "completed_by_id": "b75c1a02-c884-40ed-a3f8-8b95f3ff7539",
            "slo_seconds": 180,
            "slo_delta_seconds": -60
        })
    return json.dumps(res).replace("'", "''")

def make_inprogress_checklist(steps: List[Dict[str, Any]], completed_count: int, completed_timestamps: List[str]) -> str:
    res = []
    for idx, step in enumerate(steps):
        if idx < completed_count:
            completed_at = completed_timestamps[idx] if idx < len(completed_timestamps) else "2026-06-02T10:00:00Z"
            started_at = "2026-06-02T09:50:00Z" if idx == 0 else completed_timestamps[idx-1]
            res.append({
                **step,
                "completed": True,
                "status": "COMPLETED",
                "started_at": started_at,
                "completed_at": completed_at,
                "total_paused_seconds": 0,
                "completed_by_id": "b75c1a02-c884-40ed-a3f8-8b95f3ff7539",
                "slo_seconds": 180,
                "slo_delta_seconds": -60
            })
        elif idx == completed_count:
            res.append({
                **step,
                "completed": False,
                "status": "IN_PROGRESS",
                "started_at": "2026-06-02T10:05:00Z",
                "total_paused_seconds": 0,
                "slo_seconds": 180
            })
        else:
            res.append({
                **step,
                "completed": False,
                "status": "PENDING",
                "slo_seconds": 180
            })
    return json.dumps(res).replace("'", "''")

def make_pending_checklist(steps: List[Dict[str, Any]]) -> str:
    res = []
    for step in steps:
        res.append({
            **step,
            "completed": False,
            "status": "PENDING",
            "slo_seconds": 180
        })
    return json.dumps(res).replace("'", "''")


def determine_timezone(slug: str, region: str) -> str:
    """Resolves the standard timezone string based on city names or geographical regions."""
    slug_lower = slug.lower()
    
    # Explicit city checks
    if any(city in slug_lower for city in ["new-york", "boston", "philadelphia", "miami", "atlanta", "jacksonville", "charlotte", "durham", "orlando", "richmond"]):
        return "America/New_York"
    if any(city in slug_lower for city in ["chicago", "dallas", "houston", "san-antonio", "columbus", "indianapolis", "austin", "nashville", "memphis", "minneapolis", "saint-paul"]):
        return "America/Chicago"
    if "denver" in slug_lower:
        return "America/Denver"
    if "phoenix" in slug_lower:
        return "America/Phoenix"
    if any(city in slug_lower for city in ["seattle", "portland", "san-francisco", "los-angeles", "san-diego", "san-jose", "oakland", "sacramento", "fresno"]):
        return "America/Los_Angeles"
    
    # Fallbacks based on region configuration
    if region in ["northeast", "southeast"]:
        return "America/New_York"
    if region in ["northcentral", "southcentral"]:
        return "America/Chicago"
    if region in ["northwest", "southwest"]:
        return "America/Los_Angeles"
        
    return "America/Chicago"

def parse_app_users(sql_path: str) -> Tuple[Dict[str, str], Dict[str, str], Dict[str, str]]:
    """Parses app_users_seed.sql to extract associate and cashier users, their names, emails, and mapped site_ids."""
    user_emails: Dict[str, str] = {}
    user_names: Dict[str, str] = {}
    user_sites: Dict[str, str] = {}
    
    user_regex = re.compile(
        r"INSERT\s+INTO\s+users\s+\([^)]+\)\s+VALUES\s*\(\s*'([^']*)'\s*,\s*'([^']*)'\s*,\s*'([^']*)'\s*,\s*'([^']*)'\s*,\s*'([^']*)'\s*,"
    )
    site_regex = re.compile(
        r"INSERT\s+INTO\s+user_sites\s+\([^)]+\)\s+VALUES\s*\(\s*'([^']*)'\s*,\s*'([^']*)'\s*,\s*'([^']*)'\s*,"
    )
    
    with open(sql_path, "r", encoding="utf-8") as f:
        for line in f:
            # Parse user insertions
            user_match = user_regex.search(line)
            if user_match:
                uid = user_match.group(1)
                email = user_match.group(4)
                name = user_match.group(5)
                if email.startswith("associate-") or email.startswith("cashier-"):
                    user_emails[uid] = email
                    user_names[uid] = name
                    
            # Parse site mappings
            site_match = site_regex.search(line)
            if site_match:
                user_id = site_match.group(2)
                site_id = site_match.group(3)
                user_sites[user_id] = site_id
                
    return user_emails, user_names, user_sites

def main() -> None:
    print("Starting systematic task and schedule seed generation for 109 stores (double coverage mode)...")
    
    # 1. Load test stores variables
    with open(TFVARS_PATH, "r", encoding="utf-8") as f:
        tfvars_data = json.load(f)
    test_stores: Dict[str, Dict[str, str]] = tfvars_data.get("test_stores", {})
    print(f"Loaded {len(test_stores)} stores from terraform.tfvars.json")
    
    # 2. Parse seeded users and locations
    user_emails, user_names, user_sites = parse_app_users(WORKSPACE_ROOT + "/scripts/terraform/app_users_seed.sql")
    print(f"Parsed {len(user_emails)} associate/cashier users from app_users_seed.sql")
    
    # Map store IDs to their corresponding associate user details (list support)
    store_associates: Dict[str, List[Tuple[str, str, str]]] = {} # site_id -> [(user_id, name, email)]
    for user_id, site_id in user_sites.items():
        if user_id in user_emails:
            if site_id not in store_associates:
                store_associates[site_id] = []
            store_associates[site_id].append((user_id, user_names[user_id], user_emails[user_id]))
            
    print(f"Successfully mapped {len(store_associates)} stores to their active personnel lists")
    
    sql_statements: List[str] = [
        "-- ==============================================================================",
        "-- AUTOMATICALLY GENERATED TEST STORE TASKING & SCHEDULING SEEDS",
        "-- Regenerates calendar configurations, recurrences, and task executions",
        "-- ==============================================================================",
        "BEGIN;",
        "",
        "-- 1. Clear existing tasking, instances, schedules, and workload shift events",
        "DELETE FROM task_executions;",
        "DELETE FROM user_event_instances;",
        "DELETE FROM user_event_schedules;",
        "DELETE FROM events WHERE event_type IN ('RetailShift', 'ShowroomRefreshEvent', 'PerishableFreshnessEvent');",
        "",
        "-- 2. Ensure the new premium showcase calibration template exists for Volt & Vine stores",
        f"INSERT INTO tasks (id, name, description, task_type, priority, step_order, estimated_duration_minutes, checklist_template, metadata, created_at, updated_at, version) "
        f"VALUES ('{TEMPLATE_SHOWROOM_REFRESH}', 'Volt & Vine - Premium Showcase Refresh & Calibration', "
        f"'Inspect, clean, and calibrate the smart home premium showcase devices.', 'STANDARD', 2, 0, 30, "
        f"'[{{\"step\": 1, \"action\": \"Wipe down display cooktops and refrigerators\", \"required\": true}}, "
        f"{{\"step\": 2, \"action\": \"Verify interactive tablet demos are online and responsive\", \"required\": true}}, "
        f"{{\"step\": 3, \"action\": \"Ensure price tags and feature sheets match current promos\", \"required\": true}}]', "
        f"'{{}}', NOW(), NOW(), 1) "
        f"ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, checklist_template = EXCLUDED.checklist_template;",
        ""
    ]
    
    # Counters
    events_count = 0
    schedules_count = 0
    instances_count = 0
    executions_count = 0
    
    # 3. Generate scheduling and tasking records for each mapped store
    for site_id, store_info in test_stores.items():
        associate_list = store_associates.get(site_id, [])
        if not associate_list:
            print(f"WARNING: No associate users found mapped to site ID: {site_id} ({store_info['name']})")
            continue
            
        store_slug = store_info["slug"]
        store_name = store_info["name"]
        region = store_info["region"]
        
        # Resolve store timezone
        tz = determine_timezone(store_slug, region)
        clean_store_name = store_name.replace("'", "''")
        
        for idx, associate_info in enumerate(associate_list):
            user_id, user_name, user_email = associate_info
            clean_user_name = user_name.replace("'", "''")
            
            # Unique event ID and schedule ID for EACH associate
            event_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"event-shift-{store_slug}-{user_id}"))
            schedule_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"schedule-{store_slug}-{user_id}"))
            
            # A. Generate Shift Event
            sql_statements.append(f"-- Operational Shift Event for {store_name} - {user_name}")
            sql_statements.append(
                f"INSERT INTO events (id, organizer_id, site_id, task_id, name, event_type, event_style, created_at) "
                f"VALUES ('{event_id}', '{SYSTEM_ADMIN_USER_ID}', '{site_id}', NULL, "
                f"'{clean_store_name} - Associate Daily Shift ({clean_user_name})', 'RetailShift', 'BATCH', NOW()) "
                f"ON CONFLICT DO NOTHING;"
            )
            events_count += 1
            
            # B. Generate User Event Schedule
            # Starts Monday Jun 1st 2026 to Tuesday Jun 30th 2026
            sql_statements.append(
                f"INSERT INTO user_event_schedules (id, user_id, event_id, start_date, end_date, timezone, rrule, created_at) "
                f"VALUES ('{schedule_id}', '{user_id}', '{event_id}', '2026-06-01 08:00:00-05:00', '2026-06-30 16:00:00-05:00', "
                f"'{tz}', 'FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR', NOW()) "
                f"ON CONFLICT DO NOTHING;"
            )
            schedules_count += 1
            
            # C. Materialize Event Instances for current week (June 1 to June 5, 2026)
            instances_by_date = {}
            for day in range(1, 6):
                date_str = f"2026-06-0{day}"
                instance_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"instance-{store_slug}-{user_id}-{date_str}"))
                instances_by_date[date_str] = instance_id
                
                status = "EventScheduled"
                if day == 1:
                    status = "EventCompleted"
                elif day == 2:
                    status = "EventActive"
                    
                sql_statements.append(
                    f"INSERT INTO user_event_instances (id, schedule_id, instance_start_date, instance_end_date, event_status, created_at) "
                    f"VALUES ('{instance_id}', '{schedule_id}', '{date_str} 08:00:00-05:00', '{date_str} 16:00:00-05:00', "
                    f"'{status}', NOW()) "
                    f"ON CONFLICT DO NOTHING;"
                )
                instances_count += 1
                
            # D. Generate Task Executions for the active instance (June 2nd, 2026)
            active_instance_id = instances_by_date["2026-06-02"]
            
            # We assign tasking differently to the first vs second associate to enable trading scenarios
            if idx % 2 == 0:
                # Associate 1:
                # Task 1: Register Terminal & Cash Opening (COMPLETED)
                t1_exec_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"task-register-{store_slug}-{user_id}-2026-06-02"))
                checklist_completed = make_completed_checklist(
                    STEPS_REGISTER_OPEN,
                    ["2026-06-02T08:05:12Z", "2026-06-02T08:12:45Z", "2026-06-02T08:22:14Z"]
                )
                sql_statements.append(
                    f"INSERT INTO task_executions (id, task_template_id, parent_execution_id, execution_type, "
                    f"subject_execution_id, initiator_id, assignee_id, event_instance_id, description, status, priority, "
                    f"due_at, prerequisite_execution_id, decision, completed_at, checklist_state, override_flags, locked_at, "
                    f"locked_by, retry_count, max_retries, last_error, created_at, updated_at, version, started_at, paused_at, total_paused_seconds) "
                    f"VALUES ('{t1_exec_id}', '{TEMPLATE_REGISTER_OPEN}', NULL, 'STANDARD', NULL, '{SYSTEM_ADMIN_USER_ID}', "
                    f"'{user_id}', '{active_instance_id}', 'Opening shift register setup and till cash drop checks.', "
                    f"'COMPLETED', 1, '2026-06-02 09:00:00-05:00', NULL, NULL, '2026-06-02 08:22:14-05:00', "
                    f"'{checklist_completed}', '{{}}', NULL, NULL, 0, 3, NULL, NOW(), NOW(), 1, "
                    f"'2026-06-02 08:00:00-05:00', NULL, 0) "
                    f"ON CONFLICT DO NOTHING;"
                )
                executions_count += 1
                
                # Task 2: Brand-specific (IN_PROGRESS)
                t2_exec_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"task-brand-{store_slug}-{user_id}-2026-06-02"))
                if "volt-and-vine" in store_slug:
                    checklist_state = make_inprogress_checklist(STEPS_SHOWROOM_REFRESH, 1, ["2026-06-02T11:15:00Z"])
                    sql_statements.append(
                        f"INSERT INTO task_executions (id, task_template_id, parent_execution_id, execution_type, "
                        f"subject_execution_id, initiator_id, assignee_id, event_instance_id, description, status, priority, "
                        f"due_at, prerequisite_execution_id, decision, completed_at, checklist_state, override_flags, locked_at, "
                        f"locked_by, retry_count, max_retries, last_error, created_at, updated_at, version, started_at, paused_at, total_paused_seconds) "
                        f"VALUES ('{t2_exec_id}', '{TEMPLATE_SHOWROOM_REFRESH}', NULL, 'STANDARD', NULL, '{SYSTEM_ADMIN_USER_ID}', "
                        f"'{user_id}', '{active_instance_id}', 'Calibrate interactive demo products and refresh showcase displays.', "
                        f"'IN_PROGRESS', 2, '2026-06-02 12:00:00-05:00', NULL, NULL, NULL, "
                        f"'{checklist_state}', '{{}}', "
                        f"'2026-06-02 11:00:00-05:00', '{clean_user_name}', 0, 3, NULL, NOW(), NOW(), 1, "
                        f"'2026-06-02 11:00:00-05:00', NULL, 0) "
                        f"ON CONFLICT DO NOTHING;"
                    )
                else:
                    checklist_state = make_inprogress_checklist(STEPS_PRODUCE_FRESHNESS, 1, ["2026-06-02T10:45:10Z"])
                    sql_statements.append(
                        f"INSERT INTO task_executions (id, task_template_id, parent_execution_id, execution_type, "
                        f"subject_execution_id, initiator_id, assignee_id, event_instance_id, description, status, priority, "
                        f"due_at, prerequisite_execution_id, decision, completed_at, checklist_state, override_flags, locked_at, "
                        f"locked_by, retry_count, max_retries, last_error, created_at, updated_at, version, started_at, paused_at, total_paused_seconds) "
                        f"VALUES ('{t2_exec_id}', '{TEMPLATE_PRODUCE_FRESHNESS}', NULL, 'STANDARD', NULL, '{SYSTEM_ADMIN_USER_ID}', "
                        f"'{user_id}', '{active_instance_id}', 'Verify chiller metrics and cull spoiled greens/wilted produce.', "
                        f"'IN_PROGRESS', 3, '2026-06-02 13:00:00-05:00', NULL, NULL, NULL, "
                        f"'{checklist_state}', '{{}}', "
                        f"'2026-06-02 10:30:00-05:00', '{clean_user_name}', 0, 3, NULL, NOW(), NOW(), 1, "
                        f"'2026-06-02 10:30:00-05:00', NULL, 0) "
                        f"ON CONFLICT DO NOTHING;"
                    )
                executions_count += 1
                
            else:
                # Associate 2:
                # Task 1: Empty Stockout shelf replenishment (IN_PROGRESS)
                t1_exec_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"task-stock-{store_slug}-{user_id}-2026-06-02"))
                checklist_state = make_inprogress_checklist(STEPS_SHELF_REPLENISH, 0, [])
                sql_statements.append(
                    f"INSERT INTO task_executions (id, task_template_id, parent_execution_id, execution_type, "
                    f"subject_execution_id, initiator_id, assignee_id, event_instance_id, description, status, priority, "
                    f"due_at, prerequisite_execution_id, decision, completed_at, checklist_state, override_flags, locked_at, "
                    f"locked_by, retry_count, max_retries, last_error, created_at, updated_at, version, started_at, paused_at, total_paused_seconds) "
                    f"VALUES ('{t1_exec_id}', '{TEMPLATE_SHELF_REPLENISH}', NULL, 'STANDARD', NULL, '{SYSTEM_ADMIN_USER_ID}', "
                    f"'{user_id}', '{active_instance_id}', 'Stock out correction run for core inventory aisle shelves.', "
                    f"'IN_PROGRESS', 2, '2026-06-02 15:30:00-05:00', NULL, NULL, NULL, '{checklist_state}', '{{}}', "
                    f"'2026-06-02 14:00:00-05:00', '{clean_user_name}', 0, 3, NULL, NOW(), NOW(), 1, "
                    f"'2026-06-02 14:00:00-05:00', NULL, 0) "
                    f"ON CONFLICT DO NOTHING;"
                )
                executions_count += 1
                
                # Task 2: Brand-specific (PENDING)
                t2_exec_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"task-brand-{store_slug}-{user_id}-2026-06-02"))
                if "volt-and-vine" in store_slug:
                    checklist_state = make_pending_checklist(STEPS_SHOWROOM_REFRESH)
                    sql_statements.append(
                        f"INSERT INTO task_executions (id, task_template_id, parent_execution_id, execution_type, "
                        f"subject_execution_id, initiator_id, assignee_id, event_instance_id, description, status, priority, "
                        f"due_at, prerequisite_execution_id, decision, completed_at, checklist_state, override_flags, locked_at, "
                        f"locked_by, retry_count, max_retries, last_error, created_at, updated_at, version, started_at, paused_at, total_paused_seconds) "
                        f"VALUES ('{t2_exec_id}', '{TEMPLATE_SHOWROOM_REFRESH}', NULL, 'STANDARD', NULL, '{SYSTEM_ADMIN_USER_ID}', "
                        f"'{user_id}', '{active_instance_id}', 'Calibrate interactive demo products and refresh showcase displays.', "
                        f"'PENDING', 2, '2026-06-02 16:00:00-05:00', NULL, NULL, NULL, '{checklist_state}', '{{}}', NULL, NULL, 0, 3, NULL, NOW(), NOW(), 1, "
                        f"NULL, NULL, 0) "
                        f"ON CONFLICT DO NOTHING;"
                    )
                else:
                    checklist_state = make_pending_checklist(STEPS_PRODUCE_FRESHNESS)
                    sql_statements.append(
                        f"INSERT INTO task_executions (id, task_template_id, parent_execution_id, execution_type, "
                        f"subject_execution_id, initiator_id, assignee_id, event_instance_id, description, status, priority, "
                        f"due_at, prerequisite_execution_id, decision, completed_at, checklist_state, override_flags, locked_at, "
                        f"locked_by, retry_count, max_retries, last_error, created_at, updated_at, version, started_at, paused_at, total_paused_seconds) "
                        f"VALUES ('{t2_exec_id}', '{TEMPLATE_PRODUCE_FRESHNESS}', NULL, 'STANDARD', NULL, '{SYSTEM_ADMIN_USER_ID}', "
                        f"'{user_id}', '{active_instance_id}', 'Verify chiller metrics and cull spoiled greens/wilted produce.', "
                        f"'PENDING', 3, '2026-06-02 16:00:00-05:00', NULL, NULL, NULL, '{checklist_state}', '{{}}', NULL, NULL, 0, 3, NULL, NOW(), NOW(), 1, "
                        f"NULL, NULL, 0) "
                        f"ON CONFLICT DO NOTHING;"
                    )
                executions_count += 1
                
        # Blank line between stores
        sql_statements.append("")
        
    # 3.5. Explicitly append Ryan's Supervisor shift, schedule, active instance, and task executions
    sql_statements.append("-- ==============================================================================")
    sql_statements.append("-- EXPLICIT SEEDS FOR SUPERVISOR RYAN (ryan@rmcguinness.altostrat.com)")
    sql_statements.append("-- ==============================================================================")
    
    ryan_user_id = "b75c1a02-c884-40ed-a3f8-8b95f3ff7539"
    ryan_event_id = "99999999-9999-9999-9999-999999990100"
    ryan_schedule_id = "55555555-5555-7777-0000-000000000002"
    ryan_seattle_site_id = "44444444-4444-4444-4444-444444440000"
    
    # Generate Ryan's shift event (which also maps to Seattle site)
    sql_statements.append(
        f"INSERT INTO events (id, organizer_id, site_id, task_id, name, event_type, event_style, created_at) "
        f"VALUES ('{ryan_event_id}', '{ryan_user_id}', '{ryan_seattle_site_id}', NULL, "
        f"'Ryan Scheduled Shift - Showroom Supervisor Shift', 'RetailShift', 'BATCH', NOW()) "
        f"ON CONFLICT DO NOTHING;"
    )
    events_count += 1
    
    # Generate Ryan's schedule (starts June 1st to June 30th 2026)
    sql_statements.append(
        f"INSERT INTO user_event_schedules (id, user_id, event_id, start_date, end_date, timezone, rrule, created_at) "
        f"VALUES ('{ryan_schedule_id}', '{ryan_user_id}', '{ryan_event_id}', '2026-06-01 09:00:00-07:00', '2026-06-30 17:00:00-07:00', "
        f"'America/Los_Angeles', 'FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR', NOW()) "
        f"ON CONFLICT DO NOTHING;"
    )
    schedules_count += 1
    
    # Materialize Ryan's instances for June 1 to June 5
    ryan_instances_by_date = {}
    for day in range(1, 6):
        date_str = f"2026-06-0{day}"
        ryan_instance_id = f"00000000-0000-0000-0000-00000000000{day}"
        ryan_instances_by_date[date_str] = ryan_instance_id
        
        status = "EventScheduled"
        if day == 1:
            status = "EventCompleted"
        elif day == 2:
            status = "EventActive"
            
        sql_statements.append(
            f"INSERT INTO user_event_instances (id, schedule_id, instance_start_date, instance_end_date, event_status, created_at) "
            f"VALUES ('{ryan_instance_id}', '{ryan_schedule_id}', '{date_str} 09:00:00-07:00', '{date_str} 17:00:00-07:00', "
            f"'{status}', NOW()) "
            f"ON CONFLICT DO NOTHING;"
        )
        instances_count += 1
        
    # Ryan's Active Instance ID for June 2nd
    ryan_active_instance_id = ryan_instances_by_date["2026-06-02"]
    
    # Ryan's Task 1: Register Opening (PENDING so he can test executing it)
    ryan_t1_exec_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, "ryan-task-register-2026-06-02"))
    checklist_state = make_pending_checklist(STEPS_REGISTER_OPEN)
    sql_statements.append(
        f"INSERT INTO task_executions (id, task_template_id, parent_execution_id, execution_type, "
        f"subject_execution_id, initiator_id, assignee_id, event_instance_id, description, status, priority, "
        f"due_at, prerequisite_execution_id, decision, completed_at, checklist_state, override_flags, locked_at, "
        f"locked_by, retry_count, max_retries, last_error, created_at, updated_at, version, started_at, paused_at, total_paused_seconds) "
        f"VALUES ('{ryan_t1_exec_id}', '{TEMPLATE_REGISTER_OPEN}', NULL, 'STANDARD', NULL, '{ryan_user_id}', "
        f"'{ryan_user_id}', '{ryan_active_instance_id}', 'Opening shift register setup and till cash drop checks.', "
        f"'PENDING', 1, '2026-06-02 10:00:00-07:00', NULL, NULL, NULL, '{checklist_state}', '{{}}', NULL, NULL, 0, 3, NULL, NOW(), NOW(), 1, "
        f"NULL, NULL, 0) "
        f"ON CONFLICT DO NOTHING;"
    )
    executions_count += 1
    
    # Ryan's Task 2: Showcase Refresh & Calibration (IN_PROGRESS)
    ryan_t2_exec_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, "ryan-task-showcase-2026-06-02"))
    checklist_state = make_inprogress_checklist(STEPS_SHOWROOM_REFRESH, 1, ["2026-06-02T10:15:00Z"])
    sql_statements.append(
        f"INSERT INTO task_executions (id, task_template_id, parent_execution_id, execution_type, "
        f"subject_execution_id, initiator_id, assignee_id, event_instance_id, description, status, priority, "
        f"due_at, prerequisite_execution_id, decision, completed_at, checklist_state, override_flags, locked_at, "
        f"locked_by, retry_count, max_retries, last_error, created_at, updated_at, version, started_at, paused_at, total_paused_seconds) "
        f"VALUES ('{ryan_t2_exec_id}', '{TEMPLATE_SHOWROOM_REFRESH}', NULL, 'STANDARD', NULL, '{ryan_user_id}', "
        f"'{ryan_user_id}', '{ryan_active_instance_id}', 'Calibrate interactive demo products and refresh showcase displays.', "
        f"'IN_PROGRESS', 2, '2026-06-02 13:00:00-07:00', NULL, NULL, NULL, "
        f"'{checklist_state}', '{{}}', "
        f"'2026-06-02 10:00:00-07:00', 'Ryan', 0, 3, NULL, NOW(), NOW(), 1, "
        f"'2026-06-02 10:00:00-07:00', NULL, 0) "
        f"ON CONFLICT DO NOTHING;"
    )
    executions_count += 1
    
    # Ryan's Task 3: Empty Stockout shelf replenishment (PENDING)
    ryan_t3_exec_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, "ryan-task-stockout-2026-06-02"))
    checklist_state = make_pending_checklist(STEPS_SHELF_REPLENISH)
    sql_statements.append(
        f"INSERT INTO task_executions (id, task_template_id, parent_execution_id, execution_type, "
        f"subject_execution_id, initiator_id, assignee_id, event_instance_id, description, status, priority, "
        f"due_at, prerequisite_execution_id, decision, completed_at, checklist_state, override_flags, locked_at, "
        f"locked_by, retry_count, max_retries, last_error, created_at, updated_at, version, started_at, paused_at, total_paused_seconds) "
        f"VALUES ('{ryan_t3_exec_id}', '{TEMPLATE_SHELF_REPLENISH}', NULL, 'STANDARD', NULL, '{ryan_user_id}', "
        f"'{ryan_user_id}', '{ryan_active_instance_id}', 'Stock out correction run for core inventory aisle shelves.', "
        f"'PENDING', 2, '2026-06-02 16:30:00-07:00', NULL, NULL, NULL, '{checklist_state}', '{{}}', NULL, NULL, 0, 3, NULL, NOW(), NOW(), 1, "
        f"NULL, NULL, 0) "
        f"ON CONFLICT DO NOTHING;"
    )
    executions_count += 1
    sql_statements.append("")
        
    sql_statements.append("COMMIT;")
    
    # 4. Save file
    with open(OUTPUT_SQL_PATH, "w", encoding="utf-8") as f:
        f.write("\n".join(sql_statements) + "\n")
        
    print(f"Successfully completed tasking and schedule seeds generation!")
    print(f"Generated and wrote to {OUTPUT_SQL_PATH}:")
    print(f"  - {events_count} events")
    print(f"  - {schedules_count} schedules")
    print(f"  - {instances_count} instances")
    print(f"  - {executions_count} task executions")
if __name__ == "__main__":
    main()
