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

# Divided footprint into 6 parts
REGIONS: List[str] = [
    "northeast",
    "northwest",
    "southeast",
    "southwest",
    "northcentral",
    "southcentral"
]

USER_PROFILES: List[Dict[str, str]] = [
    {
        "role_slug": "admin",
        "given_name": "Admin",
        "org_path": "Admins",
        "role_id": ROLE_SITE_MANAGER,
        "title": "Store Systems Administrator",
    },
    {
        "role_slug": "manager",
        "given_name": "Manager",
        "org_path": "Managers",
        "role_id": ROLE_SITE_MANAGER,
        "title": "Store Operations Manager",
    },
    {
        "role_slug": "cashier",
        "given_name": "Cashier",
        "org_path": "Cashiers",
        "role_id": ROLE_SITE_ASSOCIATE,
        "title": "Store Cashier",
    },
    {
        "role_slug": "associate",
        "given_name": "Associate",
        "org_path": "Associates",
        "role_id": ROLE_SITE_ASSOCIATE,
        "title": "Customer Support Associate",
    },
    {
        "role_slug": "vault",
        "given_name": "Vault",
        "org_path": "Vault",
        "role_id": ROLE_SITE_ASSOCIATE,
        "title": "Vault Cash Custodian",
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
        Geographical region identifier (northeast, northwest, southeast, southwest, northcentral, or southcentral).
    """
    name_lower: str = name.lower()

    northeast_cities: List[str] = [
        "new york", "boston", "philadelphia", "washington", "baltimore", "newark", "jersey city", "buffalo"
    ]
    northwest_cities: List[str] = [
        "seattle", "portland", "boise", "spokane", "tacoma", "anchorage"
    ]
    southeast_cities: List[str] = [
        "atlanta", "miami", "jacksonville", "charlotte", "nashville", "memphis", "tampa", 
        "arlington", "new plains", "new orleans", "virginia beach", "raleigh", "greensboro",
        "orlando", "st. petersburg", "chesapeake", "norfolk", "durham", "hialeah", "richmond"
    ]
    southwest_cities: List[str] = [
        "phoenix", "las vegas", "albuquerque", "tucson", "mesa", "chandler", "scottsdale",
        "reno", "gilbert", "glendale", "north las vegas", "henderson",
        "los angeles", "san francisco", "san diego", "san jose", "fresno", "sacramento",
        "long beach", "oakland", "bakersfield", "anaheim", "santa ana", "riverside",
        "chula vista", "irvine", "fremont", "modesto"
    ]
    northcentral_cities: List[str] = [
        "chicago", "columbus", "indianapolis", "detroit", "milwaukee", "minneapolis",
        "cleveland", "aurora", "stockton", "saint paul", "cincinnati", "lincoln",
        "des moines", "fort wayne", "madison", "toledo"
    ]

    for city in northeast_cities:
        if city in name_lower:
            return "northeast"

    for city in northwest_cities:
        if city in name_lower:
            return "northwest"

    for city in southeast_cities:
        if city in name_lower:
            return "southeast"

    for city in southwest_cities:
        if city in name_lower:
            return "southwest"

    for city in northcentral_cities:
        if city in name_lower:
            return "northcentral"

    return "southcentral"


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
                            passwords[index_key] = result
                        elif name == "regional_passwords":
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
            passwords[key] = existing_passwords[key]
        else:
            password: str = "".join(secrets.choice(alphabet) for _ in range(16))
            passwords[key] = password
    return passwords


def main() -> None:
    """Orchestrates dynamic data extraction and generates mapping output targets."""
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

    # Group stores by region
    stores_by_region: Dict[str, List[str]] = {r: [] for r in REGIONS}
    tf_stores: Dict[str, Dict[str, str]] = {}
    
    for store_id, store_name in stores:
        region: str = get_store_region(store_name)
        stores_by_region[region].append(store_id)
        
        tf_stores[store_id] = {
            "name": store_name,
            "slug": slugify(store_name),
            "region": region,
        }
    
    # Define 4 active regional managers
    ACTIVE_MANAGERS: List[str] = ["west", "northeast", "southeast", "midwest"]
    tf_regional_managers: Dict[str, Dict[str, str]] = {}
    for reg in ACTIVE_MANAGERS:
        tf_regional_managers[reg] = {
            "email": f"regional-manager-{reg}@{DOMAIN}",
            "name": f"{reg.capitalize()} Regional Manager",
            "slug": reg,
        }

    # Define exact Group-Membership mappings in terraform tfvars
    # Standard store role groups memberships
    tf_memberships: Dict[str, Dict[str, str]] = {}
    
    # Define User accounts database mapping tracking (avoiding duplicates)
    user_seed_mappings: Dict[str, Dict] = {}
    
    # Keep track of primary and secondary store mappings
    # We will dynamically define overlapping stores for managers/admins
    # Rule: For every 5 stores in a region, the manager and admin of the 1st store
    # also supports the 2nd store as a secondary site assignment!
    overlap_relations: Dict[str, str] = {} # primary_store_id -> secondary_store_id
    for reg in REGIONS:
        reg_sites = stores_by_region[reg]
        for idx in range(0, len(reg_sites), 5):
            if idx + 1 < len(reg_sites):
                primary_id = reg_sites[idx]
                secondary_id = reg_sites[idx + 1]
                overlap_relations[primary_id] = secondary_id

    # Build the exact memberships for Cloud Identity
    for store_id, store in tf_stores.items():
        for profile in USER_PROFILES:
            role_slug = profile["role_slug"]
            user_email = f"{role_slug}-{store['slug']}@{DOMAIN}"
            primary_group_slug = f"{store['slug']}-{role_slug}"
            
            # Primary membership
            primary_mem_key = f"{primary_group_slug}_{user_email}"
            tf_memberships[primary_mem_key] = {
                "user_email": user_email,
                "group_slug": primary_group_slug
            }
            
            # Register the user's master record
            user_key = f"{store_id}-{role_slug}"
            user_seed_mappings[user_key] = {
                "user_uuid": str(uuid.uuid4()),
                "username": f"{role_slug}-{store['slug']}",
                "email": user_email,
                "name": f"{profile['given_name']} {store['name']}",
                "title": f"{profile['title']} - {store['name']}",
                "role_id": profile["role_id"],
                "sites": [(store_id, True)] # List of (store_id, is_primary)
            }

    # Apply overlap memberships
    for primary_id, secondary_id in overlap_relations.items():
        primary_store = tf_stores[primary_id]
        secondary_store = tf_stores[secondary_id]
        
        # Apply to 'manager' and 'admin' profiles
        for role_slug in ["manager", "admin"]:
            # The user is the primary user from the primary store
            user_email = f"{role_slug}-{primary_store['slug']}@{DOMAIN}"
            secondary_group_slug = f"{secondary_store['slug']}-{role_slug}"
            
            # Add secondary membership in Cloud Identity!
            overlap_mem_key = f"{secondary_group_slug}_{user_email}"
            tf_memberships[overlap_mem_key] = {
                "user_email": user_email,
                "group_slug": secondary_group_slug
            }
            
            # Update the seed record to include the secondary site mapping
            primary_user_key = f"{primary_id}-{role_slug}"
            if primary_user_key in user_seed_mappings:
                user_seed_mappings[primary_user_key]["sites"].append((secondary_id, False))

    tfvars_data = {
        "test_stores": tf_stores,
        "regional_managers": tf_regional_managers,
        "store_memberships": tf_memberships
    }

    with open(tfvars_path, "w", encoding="utf-8") as f:
        json.dump(tfvars_data, f, indent=2)
    print(f"Successfully generated Terraform parameters with {len(tf_memberships)} memberships: {tfvars_path}")

    # Get state-locked passwords context
    existing_passwords: Dict[str, str] = load_existing_passwords(tfstate_path)

    # Generate passwords for all unique users
    user_keys: List[str] = list(user_seed_mappings.keys())
    for reg in ACTIVE_MANAGERS:
        user_keys.append(f"regional-manager-{reg}")

    passwords_map: Dict[str, str] = generate_passwords(user_keys, existing_passwords)
    
    with open(passwords_path, "w", encoding="utf-8", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(["target_id", "target_name", "category", "email", "temporary_password"])
        
        # Write regional managers first
        for reg in ACTIVE_MANAGERS:
            key: str = f"regional-manager-{reg}"
            email: str = f"regional-manager-{reg}@{DOMAIN}"
            writer.writerow([
                f"regional-{reg}",
                f"{reg.capitalize()} Region",
                "regional_manager",
                email,
                passwords_map[key]
            ])
            
        # Write store associates
        for key, user_rec in user_seed_mappings.items():
            store_id = user_rec["sites"][0][0]
            store_name = tf_stores[store_id]["name"]
            writer.writerow([
                store_id,
                store_name,
                f"store_{key.split('-')[-1]}",
                user_rec["email"],
                passwords_map[key]
            ])
    print(f"Successfully compiled secure passwords registry (git-ignored): {passwords_path}")

    # Compile GTE database seeding script
    print("Generating GTE database matching user registration SQL script...")
    sql_lines: List[str] = [
        "-- GTE Application Seeding Script for Provisioned Cloud Identity Users",
        "-- Automatically maps Google Workspace profiles natively under internal tables",
        "BEGIN;",
        ""
    ]

    # Helper mapping function
    def get_manager_slug_for_region(region: str) -> str:
        if region in ["northwest", "southwest"]:
            return "west"
        if region in ["northcentral", "southcentral"]:
            return "midwest"
        return region

    # Seed Regional Managers
    sql_lines.append("-- ==============================================================================")
    sql_lines.append("-- SEEDING REGIONAL MANAGER PROFILES MATRIX")
    sql_lines.append("-- ==============================================================================")
    
    for reg in ACTIVE_MANAGERS:
        manager_uuid: str = str(uuid.uuid4())
        username: str = f"regional-manager-{reg}"
        email: str = f"{username}@{DOMAIN}"
        title: str = f"{reg.capitalize()} Regional Operations Director"
        meta_json: str = json.dumps({"title": title})
        meta_escaped: str = meta_json.replace("'", "''")

        sql_lines.append(f"-- REGIONAL MANAGER: {reg.upper()}")
        sql_lines.append(
            f"INSERT INTO users (id, o_auth_provider, o_auth_id, email, name, metadata, created_at, updated_at, version) VALUES "
            f"('{manager_uuid}', 'google', 'user-oauth-id-{username}', '{email}', '{reg.capitalize()} Regional Manager', '{meta_escaped}', NOW(), NOW(), 1) "
            f"ON CONFLICT (o_auth_provider, o_auth_id) DO UPDATE SET email = EXCLUDED.email, name = EXCLUDED.name, metadata = EXCLUDED.metadata, updated_at = NOW();"
        )
        sql_lines.append(
            f"INSERT INTO user_roles (user_id, role_id) VALUES "
            f"('{manager_uuid}', '{ROLE_REGION_MANAGER}') "
            f"ON CONFLICT DO NOTHING;"
        )
        
        # Find all stores belonging to this manager's responsibility
        reg_sites: List[str] = []
        for store_id, store in tf_stores.items():
            if get_manager_slug_for_region(store["region"]) == reg:
                reg_sites.append(store_id)
                
        for idx, store_id in enumerate(reg_sites):
            user_site_uuid: str = str(uuid.uuid4())
            is_primary = "TRUE" if idx == 0 else "FALSE"
            sql_lines.append(
                f"INSERT INTO user_sites (id, user_id, site_id, is_primary, metadata, created_at) VALUES "
                f"('{user_site_uuid}', '{manager_uuid}', '{store_id}', {is_primary}, '{{}}', NOW()) "
                f"ON CONFLICT DO NOTHING;"
            )
        sql_lines.append("")

    # Seed Storefront Associates with Multi-Store assignments
    sql_lines.append("-- ==============================================================================")
    sql_lines.append("-- SEEDING STOREFRONT ASSOCIATE IDENTITY MATRICES")
    sql_lines.append("-- ==============================================================================")
    
    for key, user_rec in user_seed_mappings.items():
        meta_json: str = json.dumps({"title": user_rec["title"]})
        meta_escaped: str = meta_json.replace("'", "''")
        
        sql_lines.append(f"-- USER: {user_rec['email']}")
        sql_lines.append(
            f"INSERT INTO users (id, o_auth_provider, o_auth_id, email, name, metadata, created_at, updated_at, version) VALUES "
            f"('{user_rec['user_uuid']}', 'google', 'user-oauth-id-{user_rec['username']}', '{user_rec['email']}', '{user_rec['name']}', '{meta_escaped}', NOW(), NOW(), 1) "
            f"ON CONFLICT (o_auth_provider, o_auth_id) DO UPDATE SET email = EXCLUDED.email, name = EXCLUDED.name, metadata = EXCLUDED.metadata, updated_at = NOW();"
        )
        sql_lines.append(
            f"INSERT INTO user_roles (user_id, role_id) VALUES "
            f"('{user_rec['user_uuid']}', '{user_rec['role_id']}') "
            f"ON CONFLICT DO NOTHING;"
        )
        
        # Insert one or more sites depending on overlaps
        for store_id, is_primary in user_rec["sites"]:
            user_site_uuid: str = str(uuid.uuid4())
            primary_sql: str = "TRUE" if is_primary else "FALSE"
            sql_lines.append(
                f"INSERT INTO user_sites (id, user_id, site_id, is_primary, metadata, created_at) VALUES "
                f"('{user_site_uuid}', '{user_rec['user_uuid']}', '{store_id}', {primary_sql}, '{{}}', NOW()) "
                f"ON CONFLICT DO NOTHING;"
            )
        sql_lines.append("")

    sql_lines.append("COMMIT;")
    
    with open(sql_seed_path, "w", encoding="utf-8") as f:
        f.write("\n".join(sql_lines))
    print(f"Successfully generated database seeding queries: {sql_seed_path}")
    print(f"Provisioning setup complete! Total unique users seed: {len(user_seed_mappings) + len(ACTIVE_MANAGERS)}")


if __name__ == "__main__":
    main()
