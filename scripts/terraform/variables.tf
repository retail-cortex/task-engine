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

# Terraform Variables Specifications for Retail Users Provisioning Stack

variable "project_id" {
  type        = string
  description = "The target Google Cloud Platform project identifier."
  default     = "cs-poc-gvosjaln9q6gcudiayjqdzq"
}

variable "region" {
  type        = string
  description = "The target default regional location zone for resource deployments."
  default     = "us-central1"
}

variable "domain" {
  type        = string
  description = "The primary Google Workspace / Cloud Identity domain context where users will be created."
  default     = "rmcguinness.altostrat.com"
}

variable "customer_id" {
  type        = string
  description = "The Workspace corporate Customer ID."
  default     = "C012qshn5"
}

variable "gcp_project_roles" {
  type        = list(string)
  description = "The standard least-privilege enterprise IAM roles assigned to the store-specific operations groups at project scope."
  default = [
    "roles/run.invoker"
  ]
}

variable "env" {
  type        = string
  description = "The target deployment lifecycle context (e.g. prod, dev)."
  default     = "prod"
}

variable "test_stores" {
  type = map(object({
    name   = string
    slug   = string
    region = string
  }))
  description = "Active retail teststore site locations. Automatically extracted and generated from application databases."
}

variable "regional_managers" {
  type = map(object({
    email = string
    name  = string
    slug  = string
  }))
  description = "Dynamic regional manager identity profile targets mapped across stores regional groups."
}

variable "store_memberships" {
  type = map(object({
    user_email = string
    group_slug = string
  }))
  description = "Static and dynamic store role membership associations generated out-of-band by provisioning engine."
}

