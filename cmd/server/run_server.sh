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

# Locate main.go inside the runfiles, and resolve its real physical path
# to escape the Bazel sandbox and find the real workspace root.
TARGET_FILE="cmd/server/main.go"
if [ ! -f "$TARGET_FILE" ]; then
  # Fallback to finding it from BASH_SOURCE directory
  SCRIPT_DIR="$( cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
  TARGET_FILE="$SCRIPT_DIR/main.go"
fi

# Resolve symlink to the real file in the physical workspace
while [ -h "$TARGET_FILE" ]; do
  DIR="$( cd -P "$( dirname "$TARGET_FILE" )" && pwd )"
  TARGET_FILE="$(readlink "$TARGET_FILE")"
  [[ $TARGET_FILE != /* ]] && TARGET_FILE="$DIR/$TARGET_FILE"
done

# TARGET_FILE is now the absolute path to the real physical main.go
REAL_DIR="$( cd -P "$( dirname "$TARGET_FILE" )" && pwd )"
# Since main.go is in cmd/server/, the workspace root is two levels up from REAL_DIR
WORKSPACE_ROOT="$( cd "${REAL_DIR}/../.." && pwd )"
cd "$WORKSPACE_ROOT"

# Ensure logs directory exists
mkdir -p logs

echo "➜ Go Server starting... Logs redirected to logs/go_server.log"

# Redirect all stdout & stderr to log file
exec > logs/go_server.log 2>&1

# Run the compiled Go server binary from the bazel-bin directory
exec "$WORKSPACE_ROOT/bazel-bin/cmd/server/server_/server"
