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

# Dynamic Google Cloud Identity & GCP IAM Enterprise Access Orchestrator
# Dynamically processes dynamic store configurations to scale 550 personnel identities cleanly

# ==============================================================================
# 1. Nesting Hierarchies & Local Flattening Computations
# ==============================================================================

locals {
  # Standard role slugs matching store roles
  roles = ["admin", "manager", "cashier", "associate", "vault"]

  # Standard 6 regions
  regions = ["northeast", "northwest", "southeast", "southwest", "northcentral", "southcentral"]

  # Cartesian product flattening: [110 stores * 5 roles] -> 550 store-role groups
  store_role_groups = merge([
    for store_id, store in var.test_stores : {
      for role in local.roles :
      "${store.slug}-${role}" => {
        store_id   = store_id
        store_name = store.name
        store_slug = store.slug
        role       = role
        group_id   = "${store.slug}-${role}@${var.domain}"
      }
    }
  ]...)

  # Dynamic association mapping binding Regional Managers directly into the manager group of each store in their region
  regional_manager_memberships = {
    for store_id, store in var.test_stores :
    "${store_id}-regional-manager" => {
      store_id     = store_id
      store_name   = store.name
      manager_slug = (
        store.region == "northwest" || store.region == "southwest" ? "west" :
        store.region == "northcentral" || store.region == "southcentral" ? "midwest" :
        store.region
      )
      group_key    = "${store.slug}-manager"
    }
  }
}

# ==============================================================================
# 2. Google Workspace Organizational Units (OUs) Orchestration
# ==============================================================================

# 
# ==============================================================================
# 4. Operations Access Control Groups (GCP Native Cloud Identity)
# ==============================================================================

# Dynamic Google Cloud Identity role-based groups per retail store location (5 roles * 110 stores = 550 groups)
resource "google_cloud_identity_group" "store_role_groups" {
  for_each     = local.store_role_groups
  display_name = "${each.value.store_name} ${title(each.value.role)} Group"
  description  = "Cloud Identity security role-based group mapping operational personnel with role ${each.value.role} for ${each.value.store_name}."

  parent               = "customerId/${var.customer_id}"
  initial_group_config = "WITH_INITIAL_OWNER"

  group_key {
    id = each.value.group_id
  }

  labels = {
    "cloudidentity.googleapis.com/groups.discussion_forum" = ""
  }
}

# 6 Regional Security Groups representing US retail footprints
resource "google_cloud_identity_group" "regional_groups" {
  for_each     = toset(local.regions)
  display_name = "OmniMart Region - ${title(each.key)} Group"
  description  = "Regional security group representing standard retail footprint for the ${title(each.key)} region."

  parent               = "customerId/${var.customer_id}"
  initial_group_config = "WITH_INITIAL_OWNER"

  group_key {
    id = "omnimart-region-${each.key}@${var.domain}"
  }

  labels = {
    "cloudidentity.googleapis.com/groups.discussion_forum" = ""
  }
}

# Nest the store-specific role groups directly inside their geographical region groups
resource "google_cloud_identity_group_membership" "region_store_nests" {
  for_each = local.store_role_groups
  group    = google_cloud_identity_group.regional_groups[var.test_stores[each.value.store_id].region].id

  preferred_member_key {
    id = each.value.group_id
  }

  roles {
    name = "MEMBER"
  }

  depends_on = [
    google_cloud_identity_group.store_role_groups,
    google_cloud_identity_group.regional_groups
  ]
}

# Enforce individual membership records mapping personnel to role-based store groups
resource "google_cloud_identity_group_membership" "members" {
  for_each = var.store_memberships
  group    = google_cloud_identity_group.store_role_groups[each.value.group_slug].id

  preferred_member_key {
    id = each.value.user_email
  }

  roles {
    name = "MEMBER"
  }

  depends_on = [
    google_cloud_identity_group.store_role_groups
  ]
}

# Add Regional Managers dynamically as active members of the store manager groups in their region
resource "google_cloud_identity_group_membership" "regional_managers" {
  for_each = local.regional_manager_memberships
  group    = google_cloud_identity_group.store_role_groups[each.value.group_key].id

  preferred_member_key {
    id = var.regional_managers[each.value.manager_slug].email
  }

  roles {
    name = "MEMBER"
  }

  depends_on = [
    google_cloud_identity_group.store_role_groups
  ]
}

# Parent Cloud Identity Group nesting all individual store groups to scale project IAM limits
resource "google_cloud_identity_group" "all_stores_group" {
  display_name         = "All Stores Operational Parent Group"
  description          = "Parent Cloud Identity group nested above all individual store groups to scale project-level IAM bindings."
  parent               = "customerId/${var.customer_id}"
  initial_group_config = "WITH_INITIAL_OWNER"

  group_key {
    id = "store-all-stores-parent-group@${var.domain}"
  }

  labels = {
    "cloudidentity.googleapis.com/groups.discussion_forum" = ""
  }
}

# Dynamically nest all individual store role groups as members of the parent group
resource "google_cloud_identity_group_membership" "parent_group_members" {
  for_each = local.store_role_groups
  group    = google_cloud_identity_group.all_stores_group.id

  preferred_member_key {
    id = each.value.group_id
  }

  roles {
    name = "MEMBER"
  }

  depends_on = [
    google_cloud_identity_group.store_role_groups
  ]
}


# ==============================================================================
# 5. Google Cloud Platform Integration (IAM Least Privilege Roles)
# ==============================================================================

# Assign project-level least privilege invoker roles dynamically to the parent operations group
resource "google_project_iam_member" "store_group_roles" {
  for_each = toset(var.gcp_project_roles)
  project  = var.project_id
  role     = each.value
  member   = "group:${google_cloud_identity_group.all_stores_group.group_key[0].id}"
}

# Enforce bucket-level least privilege read access on the GCS SOP documents bucket
# Isolates object operations to prevent project-level global storage view access
resource "google_storage_bucket" "sop_bucket" {
  name          = "retail-tasking-sops-${var.env}"
  location      = "US"
  force_destroy = true

  public_access_prevention    = "enforced"
  uniform_bucket_level_access = true
}

resource "google_storage_bucket_iam_member" "store_group_sop_viewer" {
  bucket = google_storage_bucket.sop_bucket.name
  role   = "roles/storage.objectViewer"
  member = "group:${google_cloud_identity_group.all_stores_group.group_key[0].id}"
}


