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

import type { TaskExecution, TaskTemplate, ChecklistStep } from '../components/types';

// Resolves and deserialises Checklist JSON States recursively (supports Raw Arrays and Base64 strings)
export const resolveChecklist = (state: any): ChecklistStep[] => {
  if (!state) return [];
  if (Array.isArray(state)) return state;
  if (typeof state === 'string') {
    try {
      const parsed = JSON.parse(state);
      if (Array.isArray(parsed)) return parsed;
    } catch {
      try {
        const base64 = atob(state);
        const parsed = JSON.parse(base64);
        if (Array.isArray(parsed)) return parsed;
      } catch {}
    }
  }
  return [];
};

// Normalises dynamic database objects returned by backend GORM schemas into strictly typed TaskExecutions
export const normalizeTask = (t: any): TaskExecution => {
  const mappedTemplate = t.Task || t.task || undefined;
  let normalizedTemplate: TaskTemplate | undefined = undefined;
  if (mappedTemplate) {
    normalizedTemplate = {
      ID: mappedTemplate.id || mappedTemplate.ID || '',
      Name: mappedTemplate.name || mappedTemplate.Name || '',
      Description: mappedTemplate.description || mappedTemplate.Description || '',
      TaskType: mappedTemplate.task_type || mappedTemplate.TaskType || 'STANDARD',
      Priority: mappedTemplate.priority !== undefined ? mappedTemplate.priority : (mappedTemplate.Priority !== undefined ? mappedTemplate.Priority : 3),
      ChecklistTemplate: mappedTemplate.checklist_template || mappedTemplate.ChecklistTemplate
    };
  }
  
  let rawChecklist = t.checklist_state || t.ChecklistState || '';
  let parsedChecklist: any = null;
  if (rawChecklist) {
    try {
      parsedChecklist = JSON.parse(rawChecklist);
    } catch {
      try {
        parsedChecklist = JSON.parse(atob(rawChecklist));
      } catch {}
    }
  }

  const isChecklistEmpty = !parsedChecklist || !Array.isArray(parsedChecklist) || parsedChecklist.length === 0;

  if (isChecklistEmpty && normalizedTemplate) {
    const templateChecklist = normalizedTemplate.ChecklistTemplate || '';
    rawChecklist = typeof templateChecklist === 'string' ? templateChecklist : JSON.stringify(templateChecklist);
  }
  if (!rawChecklist) {
    rawChecklist = '[]';
  }

  return {
    id: t.id || t.ID || '',
    task_template_id: t.task_template_id || t.TaskTemplateID || '',
    Task: normalizedTemplate,
    execution_type: t.execution_type || t.ExecutionType || 'STANDARD',
    status: t.status || t.Status || 'PENDING',
    priority: t.priority !== undefined ? t.priority : (t.Priority !== undefined ? t.Priority : 3),
    description: t.description || t.Description || (normalizedTemplate ? normalizedTemplate.Description : ''),
    due_at: t.due_at || t.DueAt,
    checklist_state: rawChecklist,
    locked_by: t.locked_by || t.LockedBy,
    locked_at: t.locked_at || t.LockedAt,
    assignee_id: t.assignee_id || t.AssigneeID || '',
    retry_count: t.retry_count !== undefined ? t.retry_count : (t.RetryCount !== undefined ? t.RetryCount : 0),
    max_retries: t.max_retries !== undefined ? t.max_retries : (t.MaxRetries !== undefined ? t.MaxRetries : 3),
    last_error: t.last_error || t.LastError,
    created_at: t.created_at || t.CreatedAt || new Date().toISOString()
  };
};

// Decodes cryptographically signed Google OAuth ID tokens payloads natively inside the browser context
export const decodeOAuthTokenClaims = (idToken: string): { name: string; email: string; picture: string | null } | null => {
  try {
    const base64Url = idToken.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
    
    const claims = JSON.parse(jsonPayload);
    return {
      name: claims.name || 'Associate Hanna',
      email: claims.email || 'hanna@rmcguinness.altostrat.com',
      picture: claims.picture || null
    };
  } catch (e) {
    console.warn("[OAuth API] Claims extraction error: ", e);
    return null;
  }
};
