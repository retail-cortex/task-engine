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

TRUNCATE TABLE time_zones CASCADE;

INSERT INTO time_zones (id, name, timezone_offset, created_at, updated_at) VALUES
	('11111111-2222-3333-4444-555555550001', 'America/New_York', 'UTC-05:00', NOW(), NOW()),
	('11111111-2222-3333-4444-555555550002', 'America/Chicago', 'UTC-06:00', NOW(), NOW());
