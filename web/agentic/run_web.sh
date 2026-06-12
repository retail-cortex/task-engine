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

set -e

# Locate package.json inside the runfiles, and resolve its real physical path
# to escape the Bazel sandbox and find the real workspace root.
TARGET_FILE="web/agentic/package.json"
if [ ! -f "$TARGET_FILE" ]; then
  # Fallback to finding it from BASH_SOURCE directory
  SCRIPT_DIR="$( cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
  TARGET_FILE="$SCRIPT_DIR/package.json"
fi

# Resolve symlink to the real file in the physical workspace
while [ -h "$TARGET_FILE" ]; do
  DIR="$( cd -P "$( dirname "$TARGET_FILE" )" && pwd )"
  TARGET_FILE="$(readlink "$TARGET_FILE")"
  [[ $TARGET_FILE != /* ]] && TARGET_FILE="$DIR/$TARGET_FILE"
done

# TARGET_FILE is now the absolute path to the real physical package.json
REAL_DIR="$( cd -P "$( dirname "$TARGET_FILE" )" && pwd )"
# Since package.json is in web/agentic/, the workspace root is two levels up from REAL_DIR
WORKSPACE_ROOT="$( cd "${REAL_DIR}/../.." && pwd )"
cd "$WORKSPACE_ROOT"

# Ensure logs directory exists
mkdir -p logs

echo "➜ React Portal starting... Logs redirected to logs/react_portal.log"

# Redirect all stdout & stderr to log file
exec > logs/react_portal.log 2>&1

# Find pnpm in common installation locations to survive a stripped PATH
PNPM_PATH="pnpm"
if ! command -v pnpm &> /dev/null; then
  for path in "/opt/homebrew/bin/pnpm" "/usr/local/bin/pnpm" "/Users/rmcguinness/Library/pnpm/pnpm" "$HOME/.local/bin/pnpm"; do
    if [ -f "$path" ]; then
      PNPM_PATH="$path"
      break
    fi
  done
fi

# Run the Vite dev server inside the web/agentic directory
exec "$PNPM_PATH" --dir "$WORKSPACE_ROOT/web/agentic" dev
