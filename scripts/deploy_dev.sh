#!/usr/bin/env bash
set -eo pipefail

# Configuration
GCP_PROJECT="rmcguinness"
GCP_REGION="us-central1"
SERVICE_NAME="gemini-task-engine-dev"
IMAGE_TAG="gcr.io/${GCP_PROJECT}/gemini-task-engine-backend:dev"

echo "========================================="
echo "Building Go API Service for Linux AMD64..."
echo "========================================="
bazel build --platforms=@rules_go//go/toolchain:linux_amd64 //cmd/server:server

echo "========================================="
echo "Building Container via Cloud Build..."
echo "========================================="
# Builds the container image remotely using Google Cloud Build
gcloud builds submit \
  --project="${GCP_PROJECT}" \
  --tag="${IMAGE_TAG}" \
  --file=Dockerfile.dev \
  .

echo "========================================="
echo "Deploying to Google Cloud Run..."
echo "========================================="
# Deploy with MODENV_RUNTIME=dev so it merges .env.dev.toml
gcloud run deploy "${SERVICE_NAME}" \
  --project="${GCP_PROJECT}" \
  --image="${IMAGE_TAG}" \
  --region="${GCP_REGION}" \
  --platform=managed \
  --set-env-vars="MODENV_RUNTIME=dev" \
  --allow-unauthenticated
