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

# Copyright 2026 Google LLC

set -e

# Locate main.py inside the runfiles, and resolve its real physical path
# to escape the Bazel sandbox and find the real workspace root.
TARGET_FILE="cmd/task_agent/main.py"
if [ ! -f "$TARGET_FILE" ]; then
  # Fallback to finding it from BASH_SOURCE directory
  SCRIPT_DIR="$( cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
  TARGET_FILE="$SCRIPT_DIR/main.py"
fi

# Resolve symlink to the real file in the physical workspace
while [ -h "$TARGET_FILE" ]; do
  DIR="$( cd -P "$( dirname "$TARGET_FILE" )" && pwd )"
  TARGET_FILE="$(readlink "$TARGET_FILE")"
  [[ $TARGET_FILE != /* ]] && TARGET_FILE="$DIR/$TARGET_FILE"
done

# TARGET_FILE is now the absolute path to the real physical main.py:
# e.g., /Users/rmcguinness/Projects/internal/gemini_task_engine/cmd/task_agent/main.py
REAL_DIR="$( cd -P "$( dirname "$TARGET_FILE" )" && pwd )"
# Since main.py is in cmd/task_agent/, the workspace root is two levels up from REAL_DIR
WORKSPACE_ROOT="$( cd "${REAL_DIR}/../.." && pwd )"
cd "$WORKSPACE_ROOT"

# Ensure logs directory exists
mkdir -p logs

echo "➜ Python Agent starting... Logs redirected to logs/python_agent.log"

# Redirect all stdout & stderr to log file
exec > logs/python_agent.log 2>&1

export GOOGLE_GENAI_USE_VERTEXAI=1
export GOOGLE_CLOUD_PROJECT="$(gcloud config get-value project 2>/dev/null || echo "cs-poc-gvosjaln9q6gcudiayjqdzq")"
export GOOGLE_CLOUD_LOCATION=us-central1
export GRPC_ENABLE_FORK_SUPPORT=true

# Dump environment variables for analysis
env > logs/env.log

# Clean the Python environment to prevent Bazel from poisoning uv
# Prepend the adk-python source directory to PYTHONPATH to bypass nspkg.pth hijacking
export PYTHONPATH="/Users/rmcguinness/Projects/google/adk-python/src:$PYTHONPATH"
unset PYTHONHOME

# Explicitly activate the local virtual environment
export VIRTUAL_ENV="$WORKSPACE_ROOT/.venv"
export PATH="$VIRTUAL_ENV/bin:$PATH"

# Diagnostic prints using raw python
python -c "import sys; print('=== PYTHON DEBUG ==='); print('EXEC:', sys.executable); print('PATH:', sys.path)"
python -c "import os; print('=== OS PATH EXISTS ==='); print('/Users/rmcguinness/Projects/google/adk-python/src/google exists:', os.path.exists('/Users/rmcguinness/Projects/google/adk-python/src/google'))"
python -c "import importlib.machinery; print('=== PATH FINDER SPEC ==='); print(importlib.machinery.PathFinder.find_spec('google'))"
python -c "import google; print('=== GOOGLE PATHS ==='); print(google.__path__)"

# Execute using the virtual environment python directly to preserve PYTHONPATH (uv strips it!)
exec python -u cmd/task_agent/main.py

# Trigger reload: Refreshing credentials as ryan@. E2E success incoming.
