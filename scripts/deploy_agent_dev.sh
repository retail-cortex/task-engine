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
SERVICE_NAME="gemini-task-agent-dev"
BACKEND_SERVICE_NAME="gemini-task-engine-dev"

if [ -z "${GOOGLE_CLOUD_PROJECT}" ]; then
  echo "Error: GOOGLE_CLOUD_PROJECT is not set and could not be resolved from active gcloud config." >&2
  exit 1
fi

echo "========================================="
echo "Resolving Backend MCP Server URL..."
echo "========================================="
BACKEND_URL=$(gcloud run services describe "${BACKEND_SERVICE_NAME}" \
  --project="${GOOGLE_CLOUD_PROJECT}" \
  --region="${GOOGLE_CLOUD_REGION}" \
  --format="value(status.url)" 2>/dev/null || true)

if [ -z "${BACKEND_URL}" ]; then
  echo "Warning: Could not resolve URL for backend service ${BACKEND_SERVICE_NAME}."
  echo "Defaulting to fallback local URL context."
  MCP_URL="http://localhost:8080/api/v1/mcp"
else
  MCP_URL="${BACKEND_URL}/api/v1/mcp"
fi
echo "Resolved MCP Server URL: ${MCP_URL}"

IMAGE_TAG="${GOOGLE_CLOUD_REGION}-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/gte-repo/gemini-task-agent:dev"

echo "========================================="
echo "Ensuring Artifact Registry Repository Exists..."
echo "========================================="
gcloud artifacts repositories create gte-repo \
  --repository-format=docker \
  --location="${GOOGLE_CLOUD_REGION}" \
  --project="${GOOGLE_CLOUD_PROJECT}" \
  --quiet 2>/dev/null || true

echo "========================================="
echo "Building Agent Container via Cloud Build..."
echo "========================================="
rm -f Dockerfile
cp Dockerfile.agent Dockerfile
trap "rm -f Dockerfile" EXIT

gcloud builds submit \
  --project="${GOOGLE_CLOUD_PROJECT}" \
  --tag="${IMAGE_TAG}" \
  .

echo "========================================="
echo "Deploying Agent to Google Cloud Run..."
echo "========================================="
PUBLIC_URL=$(gcloud run services describe "${SERVICE_NAME}" \
  --project="${GOOGLE_CLOUD_PROJECT}" \
  --region="${GOOGLE_CLOUD_REGION}" \
  --format="value(status.url)" 2>/dev/null || true)

gcloud run deploy "${SERVICE_NAME}" \
  --project="${GOOGLE_CLOUD_PROJECT}" \
  --image="${IMAGE_TAG}" \
  --region="${GOOGLE_CLOUD_REGION}" \
  --platform=managed \
  --network="rmcguinness-central" \
  --subnet="rmcguinness-snet-01" \
  --set-env-vars="MODENV_RUNTIME=dev,MCP_SERVER_URL=${MCP_URL},PUBLIC_URL=${PUBLIC_URL},GOOGLE_GENAI_USE_VERTEXAI=1,GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT},GOOGLE_CLOUD_LOCATION=${GOOGLE_CLOUD_REGION}" \
  --allow-unauthenticated
