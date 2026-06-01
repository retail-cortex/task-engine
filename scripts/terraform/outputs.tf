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

# Terraform Outputs Definitions for Retail Personnel Provisioning

output "parent_org_unit" {
  value       = googleworkspace_org_unit.root_stores.org_unit_path
  description = "The parent organization unit branch path target."
}

output "total_stores_processed" {
  value       = length(var.test_stores)
  description = "The absolute count of physical retail storefront branches processed."
}

output "total_users_provisioned" {
  value       = length(googleworkspace_user.users)
  description = "The absolute count of active domain user login profiles provisioned."
}

output "store_groups" {
  value = {
    for store_id, group in google_cloud_identity_group.store_groups :
    store_id => {
      name        = group.display_name
      email       = group.group_key[0].id
      description = group.description
    }
  }
  description = "Operational Cloud Identity Groups mapped to storefronts, used to link GCP IAM access dynamically."
}
