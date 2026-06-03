# Google Workspace & GCP Enterprise Access Provisioning Stack

This directory contains the Terraform configurations and automated provisioning engines to orchestrate Google Workspace directories, Cloud Identity groups, and Google Cloud Platform (GCP) IAM access controls at scale for retail operations.

Due to Google Workspace Directory API security boundaries, provisioning base structures requires Domain-Wide Delegation (DWD) permissions and active Google Cloud CLI authorization. This document details the mandatory manual setup steps and execution procedures.

---

## 1. Google Workspace Domain-Wide Delegation (DWD)

Google Workspace requires explicit Domain-Wide Delegation to allow the GCP Service Account programmatically to provision Organizational Units (OUs) and User profiles on behalf of a delegated Administrator account.

### Authorization Steps
1. Log in to the [Google Workspace Admin Console](https://admin.google.com) as a Super Administrator.
2. Navigate to **Security** > **Access and data control** > **API controls**.
3. Under the **Domain-wide delegation** card, click **Manage Domain-wide Delegation**.
4. Click **Add new**.
5. Configure the API client record with these exact specifications:
   * **Client ID:** `104271089685538622756`
   * **OAuth Scopes (comma-separated):**
     ```
     https://www.googleapis.com/auth/admin.directory.orgunit, https://www.googleapis.com/auth/admin.directory.user
     ```
6. Click **Authorize**.

---

## 2. Active Regional Managers Mapping Matrix

To divide the retail footprint into **6 distinct geographic footprints** while using the **4 pre-existing regional manager profiles** provisioned inside your active Google Workspace domain, the Terraform stack and provisioning engine enforce the following mapping:

| Retail Footprint Region | Mapped Regional Manager Account |
| :--- | :--- |
| `northeast` | `regional-manager-northeast@rmcguinness.altostrat.com` |
| `northwest` | `regional-manager-west@rmcguinness.altostrat.com` |
| `southeast` | `regional-manager-southeast@rmcguinness.altostrat.com` |
| `southwest` | `regional-manager-west@rmcguinness.altostrat.com` |
| `northcentral` | `regional-manager-midwest@rmcguinness.altostrat.com` |
| `southcentral` | `regional-manager-midwest@rmcguinness.altostrat.com` |

This mapping is automatically resolved inside both GORM database seeds and Cloud Identity group memberships to ensure consistent RBAC rules.

---

## 3. Execution & Provisioning Procedures

Follow this exact sequence to apply the Terraform configurations and provision the Workspace directory.

### Step 3.1: Refresh local Google Cloud SDK credentials
To sign the JWT token assertions programmatically, the active gcloud session must have fresh credentials:
```bash
gcloud auth login
```

### Step 3.2: Run Terraform Apply
Orchestrate the 550 store role groups, regional groupings, GCS SOP buckets, and parent nested security groups:
```bash
cd scripts/terraform
terraform init
terraform apply
```

### Step 3.3: Set up the Python Virtual Environment
Create a virtual environment and install the directory provisioning client libraries bypassing custom indices:
```bash
cd ../..
uv venv .venv
uv pip install --index-url https://pypi.org/simple --python .venv requests google-api-python-client
```

### Step 3.4: Run the Workspace Directory Provisioner
Run the python script to provision all organizational branches and active store user accounts in your Workspace console:
```bash
./.venv/bin/python scripts/provision_workspace_directory.py
```

> [!NOTE]
> The provisioner script is completely **idempotent** and self-healing. It queries active OUs at start to dynamically skip already provisioned directories. It also features **exponential backoff rate-limiting handlers** (with a `0.15s` delay) to automatically survive Google Directory API quota spikes.

---

## 4. Relational Database Seeding

To synchronize GTE application users, primary stores, and secondary overlapping store assignments with your live Google Workspace directories, seed the local Postgres instance via the project's native `db_loader` utility:

```bash
go run cmd/db_loader/main.go -file=scripts/terraform/app_users_seed.sql
```
