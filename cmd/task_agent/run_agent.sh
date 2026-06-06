#!/usr/bin/env bash
# Copyright 2026 Google LLC

set -e

# When running via 'bazel run', Bazel sets BUILD_WORKSPACE_DIRECTORY
# to the root of the workspace. Change directory there to resolve uv and relative paths.
if [ -n "$BUILD_WORKSPACE_DIRECTORY" ]; then
  cd "$BUILD_WORKSPACE_DIRECTORY"
fi

export GOOGLE_GENAI_USE_VERTEXAI=1
export GOOGLE_CLOUD_PROJECT=cs-poc-gvosjaln9q6gcudiayjqdzq
export GOOGLE_CLOUD_LOCATION=us-central1

exec uv run python -u cmd/task_agent/main.py
