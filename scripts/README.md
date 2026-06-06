# Enterprise Task Engine - Workspace & Deployment Script Orchestration

This directory houses build-deployment targets and identity-provisioning automation scripts used to bootstrap development environments and sync Google Workspace directories.

---

## 1. Google Cloud IAM & Workspace Prerequisites

Before executing provisioning or deployment scripts, the following IAM permissions and directory configurations must be granted to the execution identities.

### A. Cloud Deployment Permissions
The Google account executing the deployment script (`deploy_dev.sh`) requires the following roles on the GCP project `cs-poc-gvosjaln9q6gcudiayjqdzq`:

* **Cloud Build Editor (`roles/cloudbuild.builds.editor`)** - Authorizes the execution of container builds.
* **Storage Admin (`roles/storage.admin`)** - Authorizes read/write access to staging build GCS buckets and Artifact Registry.
* **Cloud Run Developer (`roles/run.developer`)** - Authorizes deploying and modifying Cloud Run service instances.
* **Service Account User (`roles/iam.serviceAccountUser`)** - Authorizes binding the Compute Engine default service account to the deployed Cloud Run service.

**Grant command (replace `<YOUR_EMAIL>`):**
```bash
for role in \
  "roles/cloudbuild.builds.editor" \
  "roles/storage.admin" \
  "roles/run.developer" \
  "roles/iam.serviceAccountUser"; do \
    gcloud projects add-iam-policy-binding "cs-poc-gvosjaln9q6gcudiayjqdzq" \
      --member="user:<YOUR_EMAIL>" \
      --role="$role"; \
done
```

### B. Cloud Build Service Account Permissions
The Compute Engine default service account running the Cloud Build agent (`10781708810-compute@developer.gserviceaccount.com`) must have the following project roles:

* **Storage Object Viewer (`roles/storage.objectViewer`)** - Allows the build worker to retrieve source tarballs from the staging GCS bucket.
* **Logs Writer (`roles/logging.logWriter`)** - Allows streaming build logs to Google Cloud Logging.
* **Artifact Registry Writer (`roles/artifactregistry.writer`)** - Allows uploading container images to Artifact Registry.
* **AlloyDB Client (`roles/alloydb.client`)** - Allows the Cloud Run container to connect to the AlloyDB database using the native dialer.

**Grant command:**
```bash
for role in \
  "roles/storage.objectViewer" \
  "roles/logging.logWriter" \
  "roles/artifactregistry.writer" \
  "roles/alloydb.client"; do \
    gcloud projects add-iam-policy-binding "cs-poc-gvosjaln9q6gcudiayjqdzq" \
      --member="serviceAccount:10781708810-compute@developer.gserviceaccount.com" \
      --role="$role"; \
done
```

### C. Workspace Directory Delegation Permissions
The script automating Google Workspace setup (`provision_workspace_directory.py`) delegates authority using service account credentials.

1. **Workspace Admin Console Configuration:**
   * Navigate to Security > Access and data control > API controls > Manage Domain-wide Delegation in Google Workspace.
   * Add a client ID configuration for the service account `workspace-provisioner@cs-poc-gvosjaln9q6gcudiayjqdzq.iam.gserviceaccount.com`.
   * Add the following OAuth scopes:
     * `https://www.googleapis.com/auth/admin.directory.orgunit`
     * `https://www.googleapis.com/auth/admin.directory.user`

2. **Execution User GCP Permissions:**
   * The developer executing the python script requires the **Service Account Token Creator (`roles/iam.serviceAccountTokenCreator`)** role on the service account in order to sign the JWT delegation claim.

---

## 2. Scripts Specification

### A. Deploy Development Environment (`deploy_dev.sh`)
* **Purpose:** Automates the local build and Google Cloud Run deployment pipeline.
* **Execution Flow:**
  1. Compiles the Go server binary for `linux/amd64` using Bazel.
  2. Stages a temporary copy of the static binary (`server_bin`) and `Dockerfile.dev` in the root context (bypassing symlink limits and standard `.gitignore` rules).
  3. Triggers Cloud Build to package the distroless Debian container and pushes it to GCR.
  4. Deploys the container to Cloud Run with `MODENV_RUNTIME=dev`.
  5. Cleans up staging files automatically on exit.
* **Execution:**
  ```bash
  ./scripts/deploy_dev.sh
  ```
  *Note: To override target projects or regions, prefix the execution: `GOOGLE_CLOUD_PROJECT=my-project GOOGLE_CLOUD_REGION=us-east1 ./scripts/deploy_dev.sh`*
  *Note: The script configures Direct VPC Egress on the VPC network `rmcguinness-central` and subnetwork `rmcguinness-snet-01` to enable private network path access to the AlloyDB instance.*


### B. Generate Provisioning Data (`generate_provisioning_data.py`)
* **Purpose:** Systematically extracts store information from relational baseline schemas and generates user directory map records.
* **Execution Flow:**
  1. Parses `scripts/dev_env.sql` to identify active storefront locations and maps them into 6 geographical divisions.
  2. Groups users (admins, managers, cashiers, associates, vault custodians) under store organizational trees.
  3. Mapped managers/admins of the first store in every five-store block to support the second store as a secondary site assignment (multi-store coverage).
  4. Resolves credentials against previous local state (`terraform/terraform.tfstate`) to lock temporary passwords across pipeline runs.
  5. Outputs three generated artifacts:
     * `scripts/terraform/terraform.tfvars.json` (Cloud Identity Group bindings)
     * `scripts/terraform/passwords_registry.csv` (Secure credentials registry)
     * `scripts/terraform/app_users_seed.sql` (Relational user profile database seeds)
* **Execution:**
  ```bash
  uv run scripts/generate_provisioning_data.py
  ```

### C. Workspace Directory Provisioner (`provision_workspace_directory.py`)
* **Purpose:** Creates Google Workspace organizational units (OUs) and user directory profiles.
* **Execution Flow:**
  1. Generates and signs a Workspace domain-wide delegation JWT token via `gcloud sign-jwt`.
  2. Calls the Google Admin Directory SDK API.
  3. Idempotently creates the nested OU structure (`/Stores/<StoreName>/{Admins,Managers,Cashiers,Associates,Vault}`).
  4. Sequentially provisions individual domain user directory logins read from `passwords_registry.csv`.
* **Execution:**
  ```bash
  uv run scripts/provision_workspace_directory.py
  ```

### D. Generate Tasking Seeds (`generate_tasking_seeds.py`)
* **Purpose:** Generates SQL database seeds representing task executions, shifts, calendar schedules, and checklists for all stores and associates.
* **Execution Flow:**
  1. Mapped store details and active personnel by parsing `app_users_seed.sql` and `terraform.tfvars.json`.
  2. Schedules daily shifts and weekly recurring calendars mapped to regional time zones.
  3. Generates task executions (including cash drawer drops, fresh chiller checks, and shelf inventory updates) across active shift days.
  4. Generates specific tasking instances for supervisor Ryan (`ryan@rmcguinness.altostrat.com`) to allow testing of Maker/Checker scenarios in the admin cockpit.
  5. Outputs `scripts/seed_store_tasks.sql`.
* **Execution:**
  ```bash
  uv run scripts/generate_tasking_seeds.py
  ```

---

## 3. Database SQL Scripts Mappings

* **`dev_env.sql`** - Baseline database schema definitions including spatial location coordinates.
* **`dev_events.sql`** - Initial scheduling schema mappings.
* **`icao_codes.sql`** - Standard airport operational weather codes reference data.
* **`time_zones.sql`** - Time zone offsets seed entries.
* **`seed_store_tasks.sql`** - Materialized task execution logs and checklists (output of `generate_tasking_seeds.py`).
