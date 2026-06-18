// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

export interface TaskTemplate {
  ID: string;
  Name: string;
  Description: string;
  TaskType: string;
  Priority: number;
  ChecklistTemplate: any;
}

export interface TaskExecution {
  id: string;
  task_template_id: string;
  Task?: TaskTemplate;
  execution_type: string;
  status: string;
  priority: number;
  description: string;
  due_at?: string;
  checklist_state: string; // JSON String representing steps
  locked_by?: string;
  locked_at?: string;
  assignee_id?: string;
  Assignee?: {
    id: string;
    name: string;
    email: string;
  };
  started_at?: string;
  paused_at?: string;
  total_paused_seconds: number;
  completed_at?: string;
  retry_count: number;
  max_retries: number;
  last_error?: string;
  created_at: string;
}

export interface ChecklistStep {
  step: number;
  action: string;
  required: boolean;
  completed?: boolean;
  status?: string;
  started_at?: string;
  paused_at?: string;
  total_paused_seconds?: number;
  completed_at?: string;
  completed_by_id?: string;
  slo_seconds?: number;
  slo_delta_seconds?: number;
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  a2uiType?: 'TRADE' | 'VAULT_DROP' | 'WEATHER' | 'STATUS';
  a2uiData?: any;
}
