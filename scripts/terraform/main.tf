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
  # Standard personnel role schemas matching Enterprise Task Engine roles
  user_profiles = [
    {
      role_slug  = "manager"
      given_name = "Manager"
      org_path   = "Managers"
      gte_role   = "SITE_MANAGER"
      title      = "Store Operations Manager"
    },
    {
      role_slug  = "lead"
      given_name = "Lead"
      org_path   = "Leads"
      gte_role   = "SITE_ASSOCIATE"
      title      = "Senior Operations Lead"
    },
    {
      role_slug  = "associate1"
      given_name = "Associate One"
      org_path   = "Associates"
      gte_role   = "SITE_ASSOCIATE"
      title      = "Customer Support Associate"
    },
    {
      role_slug  = "associate2"
      given_name = "Associate Two"
      org_path   = "Associates"
      gte_role   = "SITE_ASSOCIATE"
      title      = "Inventory Replenishment Associate"
    },
    {
      role_slug  = "vendor"
      given_name = "Vendor Checker"
      org_path   = "Vendors"
      gte_role   = "SITE_3P"
      title      = "Logistics Vendor Inspector"
    }
  ]

  # List of standard nested Organizational Units required for each store location
  sub_unit_paths = ["Managers", "Leads", "Associates", "Vendors"]

  # Nested comprehension to flatten [110 stores * 4 sub-OUs] -> 440 target Org Units mapping
  flat_sub_units = merge([
    for store_id, store in var.test_stores : {
      for sub_path in local.sub_unit_paths :
      "${store_id}-${sub_path}" => {
        store_id   = store_id
        store_name = store.name
        sub_path   = sub_path
      }
    }
  ]...)

  # Nested comprehension to flatten [110 stores * 5 roles] -> 550 total individual domain users mapping
  flat_users = merge([
    for store_id, store in var.test_stores : {
      for profile in local.user_profiles :
      "${store_id}-${profile.role_slug}" => {
        store_id    = store_id
        store_name  = store.name
        store_slug  = store.slug
        role_slug   = profile.role_slug
        email       = "${profile.role_slug}-${store.slug}@${var.domain}"
        given_name  = profile.given_name
        family_name = store.name
        org_path    = "/Stores/${store.name}/${profile.org_path}"
        gte_role    = profile.gte_role
        title       = profile.title
      }
    }
  ]...)

  # Nested comprehension to flatten [110 stores * 2 GCP roles] -> 220 IAM member targets mapping
  store_group_iam_roles = merge([
    for store_id, store in var.test_stores : {
      for role in var.gcp_project_roles :
      "${store_id}-${role}" => {
        store_id   = store_id
        store_slug = store.slug
        role       = role
      }
    }
  ]...)

  # Dynamic association mapping binding Regional Managers directly into the 110 storefront operational groups
  regional_manager_memberships = {
    for store_id, store in var.test_stores :
    "${store_id}-regional-manager" => {
      store_id     = store_id
      store_name   = store.name
      manager_slug = store.region
    }
  }
}

# ==============================================================================
# 2. Google Workspace Organizational Units (OUs) Orchestration
# ==============================================================================

# Root directory holding all operational storefronts
resource "googleworkspace_org_unit" "root_stores" {
  name                 = "Stores"
  parent_org_unit_path = "/"
  description          = "Parent directory holding all retail store locations personnel."
}

# Per-location corporate directory parent node (e.g. /Stores/Volt & Vine - Seattle)
resource "googleworkspace_org_unit" "store_units" {
  for_each             = var.test_stores
  name                 = each.value.name
  parent_org_unit_path = googleworkspace_org_unit.root_stores.org_unit_path
  description          = "Directory branch node for ${each.value.name} location staff profiles."
}

# Nested departmental sub-directories (Managers, Leads, Associates, Vendors)
# Linked directly to ensure proper sequential tree generation in the API graph
resource "googleworkspace_org_unit" "store_sub_units" {
  for_each             = local.flat_sub_units
  name                 = each.value.sub_path
  parent_org_unit_path = googleworkspace_org_unit.store_units[each.value.store_id].org_unit_path
  description          = "${each.value.sub_path} personnel records branch target under ${each.value.store_name}."
}

# Dedicated directory node branch holding Regional Management staff profiles
resource "googleworkspace_org_unit" "regional_managers" {
  name                 = "Regional Managers"
  parent_org_unit_path = googleworkspace_org_unit.root_stores.org_unit_path
  description          = "Directory branch node for regional general operations director profiles."
}

# ==============================================================================
# 3. Google Workspace Identity Provisioning (Store Staff & Regional Managers)
# ==============================================================================

# Generate secure dynamic passwords for regional managers
resource "random_password" "regional_passwords" {
  for_each = var.regional_managers
  length   = 16
  special  = true
  numeric  = true
  lower    = true
  upper    = true
}

# Provision Regional Managers in their standard Workspace Org Unit
resource "googleworkspace_user" "regional_managers" {
  for_each      = var.regional_managers
  primary_email = each.value.email
  password      = random_password.regional_passwords[each.key].result

  name {
    given_name  = "${each.value.name}"
    family_name = "Region"
  }

  org_unit_path = googleworkspace_org_unit.regional_managers.org_unit_path
  change_password_at_next_login = true
}

# Generate distinct cryptographically secure initial passwords dynamically for storefront staff
resource "random_password" "user_passwords" {
  for_each = local.flat_users
  length   = 16
  special  = true
  numeric  = true
  lower    = true
  upper    = true
}

# Core user directories orchestration block
resource "googleworkspace_user" "users" {
  for_each      = local.flat_users
  primary_email = each.value.email
  password      = random_password.user_passwords[each.key].result

  name {
    given_name  = each.value.given_name
    family_name = each.value.family_name
  }

  # Enforce dynamic mapping dependency constraints to secure parent OU lifecycle order
  org_unit_path = each.value.org_path
  change_password_at_next_login = true

  depends_on = [
    googleworkspace_org_unit.store_sub_units
  ]
}

# ==============================================================================
# 4. Operations Access Control Groups (GCP Native Cloud Identity)
# ==============================================================================

# Central Google Cloud Identity Operations Group per retail store location
resource "google_cloud_identity_group" "store_groups" {
  for_each     = var.test_stores
  display_name = "${each.value.name} Operations Group"
  description  = "Central Cloud Identity security group mapping operational personnel access for ${each.value.name}."
  
  parent = "customers/${var.customer_id}"
  
  group_key {
    id = "store-${each.value.slug}-group@${var.domain}"
  }
  
  labels = {
    "cloudidentity.googleapis.com/groups.discussion_forum" = ""
  }
}

# Enforce individual membership records connecting staff directly into their store group
resource "google_cloud_identity_group_membership" "members" {
  for_each = local.flat_users
  group    = google_cloud_identity_group.store_groups[each.value.store_id].id
  
  preferred_member_key {
    id = googleworkspace_user.users[each.key].primary_email
  }
  
  roles {
    name = "MEMBER"
  }
}

# Add Regional Managers dynamically as active members of all storefront operational groups under their region
resource "google_cloud_identity_group_membership" "regional_managers" {
  for_each = local.regional_manager_memberships
  group    = google_cloud_identity_group.store_groups[each.value.store_id].id
  
  preferred_member_key {
    id = googleworkspace_user.regional_managers[each.value.manager_slug].primary_email
  }
  
  roles {
    name = "MEMBER"
  }
}

# ==============================================================================
# 5. Google Cloud Platform Integration (IAM Least Privilege Roles)
# ==============================================================================

# Assign project-level least privilege invoker roles dynamically to groups
resource "google_project_iam_member" "store_group_roles" {
  for_each = local.store_group_iam_roles
  project  = var.project_id
  role     = each.value.role
  member   = "group:${google_cloud_identity_group.store_groups[each.value.store_id].group_key[0].id}"
}

# Enforce bucket-level least privilege read access on the GCS SOP documents bucket
# Isolates object operations to prevent project-level global storage view access
resource "google_storage_bucket_iam_member" "store_group_sop_viewer" {
  for_each = var.test_stores
  bucket   = "retail-tasking-sops-${var.env}"
  role     = "roles/storage.objectViewer"
  member   = "group:${google_cloud_identity_group.store_groups[each.key].group_key[0].id}"
}
