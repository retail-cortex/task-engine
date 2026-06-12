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

import contextvars

# Thread-local context variable to propagate the authenticated GORM User UUID
# from the incoming A2A request headers down to the outgoing MCP client headers.
active_user_id_var = contextvars.ContextVar("active_user_id", default="")

# Global mapping from context_id (session ID) to user_id (email address)
# to survive async ContextVar loss across ADK thread/task boundaries.
session_user_map = {}

