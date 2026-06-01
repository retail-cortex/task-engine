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
"""Automated data generation engine for provisioning test store personnel in GCP.

This tool extracts active storefront locations from relational SQL schemas, groups stores into 
geographical regions, generates dynamic Workspace identity maps (including regional managers),
synchronizes dynamic credentials with pre-existing state files, and compiles db-migration seeds.
"""

import csv
import json
import os
import re
import secrets
import string
import uuid
from typing import Dict, List, Tuple

# Domain variables
DOMAIN: str = "rmcguinness.altostrat.com"
GCP_PROJECT_ID: str = "cs-poc-gvosjaln9q6gcudiayjqdzq"

# Static standard mapping definitions
ROLE_ADMIN: str = "77777777-7777-7777-7777-777777770001"
ROLE_REGION_MANAGER: str = "77777777-7777-7777-7777-777777770002"
ROLE_SITE_MANAGER: str = "77777777-7777-7777-7777-777777770003"
ROLE_SITE_ASSOCIATE: str = "77777777-7777-7777-7777-777777770004"
ROLE_SITE_3P: str = "77777777-7777-7777-7777-777777770005"

REGIONS: List[str] = ["west", "northeast", "southeast", "midwest"]

USER_PROFILES: List[Dict[str, str]] = [
    {
        "role_slug": "manager",
        "given_name": "Manager",
        "org_path": "Managers",
        "role_id": ROLE_SITE_MANAGER,
        "title": "Store Operations Manager",
    },
    {
        "role_slug": "lead",
        "given_name": "Lead",
        "org_path": "Leads",
        "role_id": ROLE_SITE_ASSOCIATE,
        "title": "Senior Operations Lead",
    },
    {
        "role_slug": "associate1",
        "given_name": "Associate One",
        "org_path": "Associates",
        "role_id": ROLE_SITE_ASSOCIATE,
        "title": "Customer Support Associate",
    },
    {
        "role_slug": "associate2",
        "given_name": "Associate Two",
        "org_path": "Associates",
        "role_id": ROLE_SITE_ASSOCIATE,
        "title": "Inventory Replenishment Associate",
    },
    {
        "role_slug": "vendor",
        "given_name": "Vendor Checker",
        "org_path": "Vendors",
        "role_id": ROLE_SITE_3P,
        "title": "Logistics Vendor Inspector",
    },
]


def slugify(name: str) -> str:
    """Generates an alphanumeric URL-safe unique representation of strings.

    Args:
        name: Raw display name of the storefront location.

    Returns:
        Cleaned lowercase alphanumeric string representation.
    """
    cleaned: str = name.lower()
    cleaned = cleaned.replace("&", "and")
    cleaned = re.sub(r"[^a-z0-9\s-]", "", cleaned)
    cleaned = re.sub(r"[\s-]+", "-", cleaned)
    return cleaned.strip("-")


def get_store_region(name: str) -> str:
    """Determines standard geographical regions based on city name markers.

    Args:
        name: Raw display name of the storefront location.

    Returns:
        Geographical region identifier (west, northeast, southeast, or midwest).
    """
    name_lower: str = name.lower()

    west_cities: List[str] = [
        "seattle", "san francisco", "los angeles", "denver", "phoenix",
        "san diego", "san jose", "portland", "las vegas", "albuquerque",
        "tucson", "fresno", "sacramento", "mesa", "colorado springs",
        "long beach", "oakland", "bakersfield"
    ]
    northeast_cities: List[str] = [
        "new york", "boston", "philadelphia", "washington", "baltimore"
    ]
    southeast_cities: List[str] = [
        "atlanta", "miami", "jacksonville", "charlotte", "nashville",
        "memphis", "tampa", "arlington", "new plains", "new orleans",
        "virginia beach", "raleigh"
    ]

    for city in west_cities:
        if city in name_lower:
            return "west"

    for city in northeast_cities:
        if city in name_lower:
            return "northeast"

    for city in southeast_cities:
        if city in name_lower:
            return "southeast"

    return "midwest"


def parse_stores_from_sql(sql_file_path: str) -> List[Tuple[str, str]]:
    """Reads a relational SQL schema file and extracts sites matching site_type STORE.

    Args:
        sql_file_path: Absolute path to standard SQL database seeds file.

    Returns:
        List of tuples representing parsed (store_uuid, store_name).
    """
    stores: List[Tuple[str, str]] = []
    if not os.path.exists(sql_file_path):
        raise FileNotFoundError(f"Source file not located: {sql_file_path}")

    with open(sql_file_path, "r", encoding="utf-8") as f:
        for line in f:
            line_str = line.strip()
            if not line_str.startswith("("):
                continue

            # Process insert value lines
            content: str = line_str
            if content.endswith(","):
                content = content[:-1]
            elif content.endswith(";"):
                content = content[:-1]

            if content.startswith("(") and content.endswith(")"):
                content = content[1:-1]
            else:
                continue

            # Parse tuple fields manually to ignore parenthesis inside single quotes
            fields: List[str] = []
            current_field: List[str] = []
            in_string: bool = False
            escape: bool = False

            for char in content:
                if in_string:
                    if char == "'" and not escape:
                        in_string = False
                    elif char == "\\" and not escape:
                        escape = True
                        current_field.append(char)
                    else:
                        escape = False
                        current_field.append(char)
                else:
                    if char == "'":
                        in_string = True
                    elif char == ",":
                        fields.append("".join(current_field).strip())
                        current_field = []
                    else:
                        current_field.append(char)
            fields.append("".join(current_field).strip())

            if len(fields) >= 4:
                site_id: str = fields[0]
                site_name: str = fields[2]
                site_type: str = fields[3]
                if site_type == "STORE":
                    stores.append((site_id, site_name))

    return stores


def load_existing_passwords(tfstate_path: str) -> Dict[str, str]:
    """Reads pre-existing terraform.tfstate and extracts generated passwords results.

    Locks dynamic credential data to previously applied outputs to protect database mapping syncs.

    Args:
        tfstate_path: Absolute path to standard local terraform.tfstate file.

    Returns:
        Map of unique user keys to password result strings.
    """
    passwords: Dict[str, str] = {}
    if not os.path.exists(tfstate_path):
        return passwords

    try:
        with open(tfstate_path, "r", encoding="utf-8") as f:
            state = json.load(f)

        resources = state.get("resources", [])
        for res in resources:
            if res.get("type") == "random_password":
                name = res.get("name")
                instances = res.get("instances", [])
                for inst in instances:
                    index_key = inst.get("index_key")
                    result = inst.get("attributes", {}).get("result")
                    if index_key and result:
                        if name == "user_passwords":
                            # E.g. "44444444-4444-4444-4444-444444440000-associate1"
                            passwords[index_key] = result
                        elif name == "regional_passwords":
                            # E.g. index_key "west" -> maps to key "regional-manager-west"
                            passwords[f"regional-manager-{index_key}"] = result
        if passwords:
            print(f"Synchronized {len(passwords)} passwords context from existing: {tfstate_path}")
    except Exception as e:
        print(f"[SyncWarning] Failed to extract credentials from state: {e}")

    return passwords


def generate_passwords(user_keys: List[str], existing_passwords: Dict[str, str]) -> Dict[str, str]:
    """Generates secure, temporary passwords, locking values to existing states.

    Args:
        user_keys: List of individual user key references.
        existing_passwords: Pre-extracted credentials mapped from state blocks.

    Returns:
        Map of user key strings to password strings.
    """
    alphabet: str = string.ascii_letters + string.digits + "!@#*+=-"
    passwords: Dict[str, str] = {}
    for key in user_keys:
        if key in existing_passwords:
            # Synchronize with active state record to preserve identity mappings!
            passwords[key] = existing_passwords[key]
        else:
            # Generate new 16-character secure, random password
            password: str = "".join(secrets.choice(alphabet) for _ in range(16))
            passwords[key] = password
    return passwords


def main() -> None:
    """Orchestrates dynamic data extraction and generates mapping output targets."""
    # Define paths relative to this script location
    base_dir: str = os.path.dirname(os.path.abspath(__file__))
    sql_file_path: str = os.path.join(base_dir, "dev_env.sql")
    tf_dir: str = os.path.join(base_dir, "terraform")
    
    os.makedirs(tf_dir, exist_ok=True)
    
    tfvars_path: str = os.path.join(tf_dir, "terraform.tfvars.json")
    sql_seed_path: str = os.path.join(tf_dir, "app_users_seed.sql")
    passwords_path: str = os.path.join(tf_dir, "passwords_registry.csv")
    tfstate_path: str = os.path.join(tf_dir, "terraform.tfstate")

    print(f"Reading relational store maps from: {sql_file_path}")
    stores: List[Tuple[str, str]] = parse_stores_from_sql(sql_file_path)
    print(f"Located {len(stores)} active teststores.")

    # Group stores by region to map regional managers correctly
    stores_by_region: Dict[str, List[str]] = {r: [] for r in REGIONS}

    # 1. Generate terraform.tfvars.json map targets
    tf_stores: Dict[str, Dict[str, str]] = {}
    for store_id, store_name in stores:
        region: str = get_store_region(store_name)
        stores_by_region[region].append(store_id)
        
        tf_stores[store_id] = {
            "name": store_name,
            "slug": slugify(store_name),
            "region": region,
        }
    
    # Generate dynamic regional managers configurations map
    tf_regional_managers: Dict[str, Dict[str, str]] = {}
    for reg in REGIONS:
        tf_regional_managers[reg] = {
            "email": f"regional-manager-{reg}@{DOMAIN}",
            "name": f"{reg.capitalize()} Regional Manager",
            "slug": reg,
        }
    
    tfvars_data = {
        "test_stores": tf_stores,
        "regional_managers": tf_regional_managers
    }

    with open(tfvars_path, "w", encoding="utf-8") as f:
        json.dump(tfvars_data, f, indent=2)
    print(f"Successfully generated Terraform parameters: {tfvars_path}")

    # 2. Extract state-locked passwords to prevent database mapping drift
    existing_passwords: Dict[str, str] = load_existing_passwords(tfstate_path)

    # 3. Generate passwords targets for both store users and regional managers
    user_keys: List[str] = []
    
    # Store personnel keys
    for store_id, store in tf_stores.items():
        for profile in USER_PROFILES:
            user_keys.append(f"{store_id}-{profile['role_slug']}")
            
    # Regional manager keys
    for reg in REGIONS:
        user_keys.append(f"regional-manager-{reg}")

    passwords_map: Dict[str, str] = generate_passwords(user_keys, existing_passwords)
    
    with open(passwords_path, "w", encoding="utf-8", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(["target_id", "target_name", "category", "email", "temporary_password"])
        
        # Write regional manager records first
        for reg in REGIONS:
            key: str = f"regional-manager-{reg}"
            email: str = f"regional-manager-{reg}@{DOMAIN}"
            writer.writerow([
                f"regional-{reg}",
                f"{reg.capitalize()} Region",
                "regional_manager",
                email,
                passwords_map[key]
            ])
            
        # Write individual storefront personnel records
        for store_id, store in tf_stores.items():
            for profile in USER_PROFILES:
                key: str = f"{store_id}-{profile['role_slug']}"
                email: str = f"{profile['role_slug']}-{store['slug']}@{DOMAIN}"
                writer.writerow([
                    store_id,
                    store["name"],
                    f"store_{profile['role_slug']}",
                    email,
                    passwords_map[key]
                ])
    print(f"Successfully compiled secure passwords registry (git-ignored): {passwords_path}")

    # 4. Compile database migration application user seed queries
    print("Generating GTE database matching user registration SQL script...")
    sql_lines: List[str] = [
        "-- GTE Application Seeding Script for Provisioned Cloud Identity Users",
        "-- Automatically maps Google Workspace profiles natively under internal tables",
        "\\connect gte_dev_db;",
        "BEGIN;",
        ""
    ]

    # A. SEED REGIONAL MANAGER PROFILES MATRIX FIRST
    sql_lines.append("-- ==============================================================================")
    sql_lines.append("-- SEEDING REGIONAL MANAGER PROFILES MATRIX")
    sql_lines.append("-- ==============================================================================")
    
    for reg in REGIONS:
        manager_uuid: str = str(uuid.uuid4())
        username: str = f"regional-manager-{reg}"
        email: str = f"{username}@{DOMAIN}"
        title: str = f"{reg.capitalize()} Regional Operations Director"
        meta_json: str = json.dumps({"title": title})
        meta_escaped: str = meta_json.replace("'", "''")

        sql_lines.append(f"-- REGIONAL MANAGER: {reg.upper()}")
        
        # A.1. Insert GORM User record
        sql_lines.append(
            f"INSERT INTO users (id, o_auth_provider, o_auth_id, email, name, metadata, created_at, updated_at, version) VALUES "
            f"('{manager_uuid}', 'google', 'user-oauth-id-{username}', '{email}', '{reg.capitalize()} Regional Manager', '{meta_escaped}', NOW(), NOW(), 1) "
            f"ON CONFLICT (o_auth_provider, o_auth_id) DO UPDATE SET email = EXCLUDED.email, name = EXCLUDED.name, metadata = EXCLUDED.metadata, updated_at = NOW();"
        )

        # A.2. Insert GORM User Role mapping (REGION_MANAGER)
        sql_lines.append(
            f"INSERT INTO user_roles (user_id, role_id) VALUES "
            f"('{manager_uuid}', '{ROLE_REGION_MANAGER}') "
            f"ON CONFLICT DO NOTHING;"
        )

        # A.3. Insert GORM User Site mappings for ALL sites under this manager's region!
        # The first mapped store becomes is_primary=TRUE, others FALSE.
        reg_sites: List[str] = stores_by_region[reg]
        for idx, store_id in enumerate(reg_sites):
            user_site_uuid: str = str(uuid.uuid4())
            is_primary: str = "TRUE" if idx == 0 else "FALSE"
            sql_lines.append(
                f"INSERT INTO user_sites (id, user_id, site_id, is_primary, metadata, created_at) VALUES "
                f"('{user_site_uuid}', '{manager_uuid}', '{store_id}', {is_primary}, '{{}}', NOW()) "
                f"ON CONFLICT DO NOTHING;"
            )
        sql_lines.append("")  # Spacing between regional managers

    # B. SEED INDIVIDUAL STORE PERSONNEL MATRIX SECOND
    sql_lines.append("-- ==============================================================================")
    sql_lines.append("-- SEEDING STOREFRONT ASSOCIATE IDENTITY MATRICES")
    sql_lines.append("-- ==============================================================================")
    
    for store_id, store in tf_stores.items():
        sql_lines.append(f"-- TARGET SITE: {store['name']} (Region: {store['region'].upper()})")
        
        for profile in USER_PROFILES:
            user_uuid: str = str(uuid.uuid4())
            user_site_uuid: str = str(uuid.uuid4())
            username: str = f"{profile['role_slug']}-{store['slug']}"
            email: str = f"{username}@{DOMAIN}"
            title: str = f"{profile['title']} - {store['name']}"
            meta_json: str = json.dumps({"title": title})
            meta_escaped: str = meta_json.replace("'", "''")
            
            # B.1. Insert GORM User record
            sql_lines.append(
                f"INSERT INTO users (id, o_auth_provider, o_auth_id, email, name, metadata, created_at, updated_at, version) VALUES "
                f"('{user_uuid}', 'google', 'user-oauth-id-{username}', '{email}', '{profile['given_name']}', '{meta_escaped}', NOW(), NOW(), 1) "
                f"ON CONFLICT (o_auth_provider, o_auth_id) DO UPDATE SET email = EXCLUDED.email, name = EXCLUDED.name, metadata = EXCLUDED.metadata, updated_at = NOW();"
            )

            # B.2. Insert GORM User Role mapping
            sql_lines.append(
                f"INSERT INTO user_roles (user_id, role_id) VALUES "
                f"('{user_uuid}', '{profile['role_id']}') "
                f"ON CONFLICT DO NOTHING;"
            )

            # B.3. Insert GORM User Site primary mapping
            sql_lines.append(
                f"INSERT INTO user_sites (id, user_id, site_id, is_primary, metadata, created_at) VALUES "
                f"('{user_site_uuid}', '{user_uuid}', '{store_id}', TRUE, '{{}}', NOW()) "
                f"ON CONFLICT DO NOTHING;"
            )
            
        sql_lines.append("")  # Spacing line between stores

    sql_lines.append("COMMIT;")
    
    with open(sql_seed_path, "w", encoding="utf-8") as f:
        f.write("\n".join(sql_lines))
    print(f"Successfully generated database seeding queries: {sql_seed_path}")
    
    total_users: int = len(stores) * len(USER_PROFILES) + len(REGIONS)
    print(f"Provisioning setup complete! {total_users} total users mapped across {len(stores)} stores.")


if __name__ == "__main__":
    main()
