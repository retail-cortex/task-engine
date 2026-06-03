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
"""Automated Python orchestrator to provision Workspace OUs and Users.

Bypasses the quota project header limitation of the Terraform Workspace provider by
leveraging the official Google APIs Client Library which natively supports user quota override.
"""

import csv
import os
import sys
import time
import json
import subprocess
import requests
from googleapiclient.discovery import build
from google.oauth2.credentials import Credentials
from googleapiclient.errors import HttpError

DOMAIN = "rmcguinness.altostrat.com"
CUSTOMER_ID = "my_customer"
PROJECT_ID = "cs-poc-gvosjaln9q6gcudiayjqdzq"
CSV_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), "terraform", "passwords_registry.csv")

def get_workspace_token():
    sa_email = "workspace-provisioner@cs-poc-gvosjaln9q6gcudiayjqdzq.iam.gserviceaccount.com"
    target_user = "admin@rmcguinness.altostrat.com"
    scopes = "https://www.googleapis.com/auth/admin.directory.orgunit https://www.googleapis.com/auth/admin.directory.user"
    
    print("Building JWT claim for Workspace Domain-Wide Delegation...")
    input_dir = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "scratch")
    os.makedirs(input_dir, exist_ok=True)
    input_path = os.path.join(input_dir, "jwt_input.json")
    output_path = os.path.join(input_dir, "jwt_output.jwt")
    
    max_attempts = 40  # Poll for up to 20 minutes
    for attempt in range(1, max_attempts + 1):
        now = int(time.time())
        payload = {
            "iss": sa_email,
            "sub": target_user,
            "aud": "https://oauth2.googleapis.com/token",
            "iat": now,
            "exp": now + 3600,
            "scope": scopes
        }
        
        with open(input_path, "w") as f:
            json.dump(payload, f)
            
        try:
            subprocess.run([
                "gcloud", "iam", "service-accounts", "sign-jwt",
                input_path, output_path,
                f"--iam-account={sa_email}",
                f"--project={PROJECT_ID}"
            ], check=True, capture_output=True)
        except subprocess.CalledProcessError as e:
            print(f"gcloud sign-jwt failed: {e.stderr.decode().strip() if e.stderr else e}")
            sys.exit(1)
            
        with open(output_path, "r") as f:
            signed_jwt = f.read().strip()
            
        r_token = requests.post("https://oauth2.googleapis.com/token", data={
            "grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer",
            "assertion": signed_jwt
        })
        
        if r_token.status_code == 200:
            print("SUCCESS: Workspace Domain-Wide Delegation token retrieved successfully!")
            return r_token.json()["access_token"]
        
        err_json = r_token.json()
        if err_json.get("error") == "unauthorized_client":
            print(f"[{attempt}/{max_attempts}] Google Workspace Domain-Wide Delegation propagating... Retrying in 30s...")
            time.sleep(30)
        else:
            print("Failed to retrieve Workspace access token:", err_json)
            sys.exit(1)
            
    print("Error: Domain-Wide Delegation propagation timed out after 20 minutes.")
    sys.exit(1)

def main():
    print("Initializing Google Admin SDK Workspace Client...")
    token = get_workspace_token()
    credentials = Credentials(token=token)
    credentials.refresh = lambda request: None
    service = build('admin', 'directory_v1', credentials=credentials)
    
    # 1. Process CSV records to create Store OUs and Users
    print(f"Parsing user records from secure CSV registry: {CSV_PATH}")
    if not os.path.exists(CSV_PATH):
        print(f"Error: CSV file not found at {CSV_PATH}")
        sys.exit(1)
        
    with open(CSV_PATH, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        records = list(reader)
        
    print(f"Found {len(records)} identity profiles to reconcile.")
    
    # Set of already reconciled OUs to optimize speed, initialized with root
    created_ous = {"/"}
    try:
        print("Listing existing Workspace OUs to ensure idempotency...")
        list_req = service.orgunits().list(customerId=CUSTOMER_ID, orgUnitPath="/", type="all")
        list_res = execute_with_backoff(list_req)
        for ou in list_res.get("organizationUnits", []):
            created_ous.add(ou["orgUnitPath"])
        print(f"Loaded {len(created_ous)} existing OUs from directory.")
    except Exception as e:
        print("Could not list existing OUs (will fall back to standard insertion):", e)
        
    # Ensure base OUs exist
    if "/Stores" not in created_ous:
        create_org_unit(service, "Stores", "/")
        created_ous.add("/Stores")
        
    if "/Stores/Regional Managers" not in created_ous:
        create_org_unit(service, "Regional Managers", "/Stores")
        created_ous.add("/Stores/Regional Managers")
    
    # Gather unique stores we need to create OUs for
    stores_to_create = {}
    for row in records:
        category = row["category"]
        if category.startswith("store_"):
            store_id = row["target_id"]
            store_name = row["target_name"]
            stores_to_create[store_id] = store_name
            
    print(f"Provisioning directory branches for {len(stores_to_create)} retail locations...")
    
    for store_id, store_name in stores_to_create.items():
        store_ou_path = f"/Stores/{store_name}"
        if store_ou_path not in created_ous:
            create_org_unit(service, store_name, "/Stores")
            created_ous.add(store_ou_path)
            
        for sub in ["Admins", "Managers", "Cashiers", "Associates", "Vault"]:
            sub_path = f"{store_ou_path}/{sub}"
            if sub_path not in created_ous:
                create_org_unit(service, sub, store_ou_path)
                created_ous.add(sub_path)
                
    print("Workspace directory structure completed successfully.")
    
    # Second pass: provision all individual domain user login profiles
    print("Reconciling personnel identities...")
    for idx, row in enumerate(records):
        category = row["category"]
        email = row["email"]
        password = row["temporary_password"]
        target_name = row["target_name"]
        
        if category == "custom_user":
            given_name = target_name.split(" ")[0]
            family_name = target_name.split(" ")[1] if len(target_name.split(" ")) > 1 else "Altostrat"
            org_path = row["target_id"]
        elif category == "regional_manager":
            reg_slug = email.split("@")[0].replace("regional-manager-", "")
            given_name = reg_slug.capitalize()
            family_name = "Region"
            org_path = "/Stores/Regional Managers"
        elif category.startswith("store_"):
            role_slug = category.replace("store_", "")
            family_name = target_name
            if role_slug == "admin":
                given_name = "Admin"
                org_path = f"/Stores/{target_name}/Admins"
            elif role_slug == "manager":
                given_name = "Manager"
                org_path = f"/Stores/{target_name}/Managers"
            elif role_slug == "cashier":
                given_name = "Cashier"
                org_path = f"/Stores/{target_name}/Cashiers"
            elif role_slug == "associate":
                given_name = "Associate"
                org_path = f"/Stores/{target_name}/Associates"
            elif role_slug == "vault":
                given_name = "Vault"
                org_path = f"/Stores/{target_name}/Vault"
            else:
                print(f"Warning: Unknown role slug: {role_slug}")
                continue
        else:
            print(f"Warning: Unknown category: {category}")
            continue
            
        print(f"[{idx+1}/{len(records)}] Syncing: {email} under {org_path}")
        try:
            create_user_profile(service, email, given_name, family_name, password, org_path)
        except Exception as err:
            print(f"WARNING: Sync failed for user {email}: {err}. Proceeding to next profile...")

def execute_with_backoff(request, max_retries=5):
    delay = 1.0
    for attempt in range(max_retries):
        try:
            res = request.execute()
            time.sleep(0.15)  # Proactive short throttle to avoid aggressive burst limits
            return res
        except HttpError as err:
            if err.resp.status in [403, 429] and ("quotaExceeded" in str(err) or "rateLimitExceeded" in str(err) or err.resp.status == 429):
                print(f"Rate limit or Quota exceeded. Retrying in {delay:.1f}s (attempt {attempt + 1}/{max_retries})...")
                time.sleep(delay)
                delay *= 2.0
            else:
                raise err
    raise Exception("Max retries exceeded due to API rate limits.")

def create_org_unit(service, name, parent_path):
    path = f"{parent_path}/{name}" if parent_path != "/" else f"/{name}"
    body = {
        "name": name,
        "parentOrgUnitPath": parent_path,
        "description": f"Directory branch for {name}"
    }
    try:
        req = service.orgunits().insert(customerId=CUSTOMER_ID, body=body)
        result = execute_with_backoff(req)
        print(f"Created OU: {path}")
        return result
    except HttpError as err:
        if err.resp.status == 409:
            print(f"OU exists: {path}")
            return None
        else:
            print(f"Failed to create OU {path}: {err}")
            raise err

def create_user_profile(service, email, given_name, family_name, password, org_path):
    body = {
        "primaryEmail": email,
        "name": {
            "givenName": given_name,
            "familyName": family_name
        },
        "password": password,
        "orgUnitPath": org_path,
        "changePasswordAtNextLogin": True
    }
    try:
        req = service.users().insert(body=body)
        result = execute_with_backoff(req)
        print(f"Created User: {email}")
        return result
    except HttpError as err:
        if err.resp.status == 409:
            print(f"User already exists: {email}")
            return None
        else:
            print(f"Failed to create User {email}: {err}")
            raise err

if __name__ == "__main__":
    main()
