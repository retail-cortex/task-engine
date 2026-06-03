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

# Standard Infrastructure Providers Integration for Retail Workspace
# Maps global backend definitions, remote GCS lock paths, and multi-provider imports

terraform {
  required_version = ">= 1.5.0"

  # backend "gcs" {
  #   bucket = "cs-poc-gvosjaln9q6gcudiayjqdzq-tfstate"
  #   prefix = "env/prod/users"
  # }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }
}

# Dynamic credentials integration via active gcloud first-party session
data "external" "gcloud_token" {
  program = ["sh", "-c", "echo \"{\\\"token\\\": \\\"$(gcloud auth application-default print-access-token)\\\"}\""]
}

provider "google" {
  project               = var.project_id
  region                = var.region
  user_project_override = true
  billing_project       = var.project_id
  access_token          = data.external.gcloud_token.result.token
}

