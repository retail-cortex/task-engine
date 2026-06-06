#!/bin/bash
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

# Resolve the absolute path of this script, following any symlinks (Bazel runfiles)
SCRIPT_PATH="${BASH_SOURCE[0]}"
if [ -z "$SCRIPT_PATH" ]; then
  SCRIPT_PATH="$0"
fi

# Resolve symlinks to find the real file in the host workspace
while [ -h "$SCRIPT_PATH" ]; do
  DIR="$(cd -P "$(dirname "$SCRIPT_PATH")" && pwd)"
  SYM="$(readlink "$SCRIPT_PATH")"
  if [[ "$SYM" = /* ]]; then
    SCRIPT_PATH="$SYM"
  else
    SCRIPT_PATH="$DIR/$SYM"
  fi
done

# Get the directory of the resolved script (scripts/) and then its parent (workspace root)
REAL_SCRIPTS_DIR="$(cd -P "$(dirname "$SCRIPT_PATH")" && pwd)"
WORKSPACE_ROOT="$(dirname "$REAL_SCRIPTS_DIR")"

if [ -n "$WORKSPACE_ROOT" ] && [ -d "$WORKSPACE_ROOT/.venv" ]; then
  echo "🚀 E2E Test Runner: Successfully resolved host workspace root at: $WORKSPACE_ROOT"
  cd "$WORKSPACE_ROOT"
else
  echo "⚠️ E2E Test Runner: Warning, could not resolve host workspace root. Running in place."
fi

# Execute the failsafe Python E2E test runner directly via the virtualenv Python
exec .venv/bin/python scripts/e2e_test.py
