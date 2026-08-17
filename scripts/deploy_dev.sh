#!/usr/bin/env bash

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

set -eo pipefail

# Configuration
GOOGLE_CLOUD_PROJECT="${GOOGLE_CLOUD_PROJECT:-$(gcloud config get-value project 2>/dev/null)}"
GOOGLE_CLOUD_REGION="${GOOGLE_CLOUD_REGION:-$(gcloud config get-value compute/region 2>/dev/null)}"
GOOGLE_CLOUD_REGION="${GOOGLE_CLOUD_REGION:-us-central1}"
SERVICE_NAME="gemini-task-engine-dev"

if [ -z "${GOOGLE_CLOUD_PROJECT}" ]; then
  echo "Error: GOOGLE_CLOUD_PROJECT is not set and could not be resolved from active gcloud config." >&2
  exit 1
fi

IMAGE_TAG="${GOOGLE_CLOUD_REGION}-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/gte-repo/gemini-task-engine-backend:dev"

echo "========================================="
echo "Ensuring Artifact Registry Repository Exists..."
echo "========================================="
gcloud artifacts repositories create gte-repo \
  --repository-format=docker \
  --location="${GOOGLE_CLOUD_REGION}" \
  --project="${GOOGLE_CLOUD_PROJECT}" \
  --quiet 2>/dev/null || true

echo "========================================="
echo "Building Go API Service for Linux AMD64..."
echo "========================================="
bazel --output_base=.bazel-deploy-cache build --platforms=@rules_go//go/toolchain:linux_amd64 //cmd/server:server

echo "========================================="
echo "Building Container via Cloud Build..."
echo "========================================="
# Copy Dockerfile.dev and target binary to root context to support all gcloud CLI versions and bypass symlink exclusions
rm -f Dockerfile server_bin
cp Dockerfile.dev Dockerfile
cp bazel-bin/cmd/server/server_/server server_bin
trap "rm -f Dockerfile server_bin" EXIT

gcloud builds submit \
  --project="${GOOGLE_CLOUD_PROJECT}" \
  --tag="${IMAGE_TAG}" \
  .

echo "========================================="
echo "Deploying to Google Cloud Run..."
echo "========================================="
# Construct AlloyDB connection string dynamically using active project and region
DB_CONN="projects/${GOOGLE_CLOUD_PROJECT}/locations/${GOOGLE_CLOUD_REGION}/clusters/gemini-task-manager-dev/instances/gemini-task-manager-dev-primary"

gcloud run deploy "${SERVICE_NAME}" \
  --project="${GOOGLE_CLOUD_PROJECT}" \
  --image="${IMAGE_TAG}" \
  --region="${GOOGLE_CLOUD_REGION}" \
  --platform=managed \
  --network="rmcguinness-central" \
  --subnet="rmcguinness-snet-01" \
  --set-env-vars="MODENV_RUNTIME=dev,DB_CONNECTION_STRING=${DB_CONN}" \
  --allow-unauthenticated
