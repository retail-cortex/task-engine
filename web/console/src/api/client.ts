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

import type { TaskExecution, ChecklistStep, ChatMessage } from '../components/types';

// Custom Response Error mapping the HTTP status code (enables dynamic client auth interceptors!)
export class ResponseError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
    this.name = 'ResponseError';
  }
}

// Centralised Template API Enpoints maps (prevents URL hardcoding across code files!)
const ENDPOINTS = {
  HEALTH_READINESS: '/health/readiness',
  ME: (orgId: string) => `/api/v1/organizations/${orgId}/me`,
  SITES: (orgId: string) => `/api/v1/organizations/${orgId}/sites`,
  TASKS: (orgId: string, siteId: string) => `/api/v1/organizations/${orgId}/sites/${siteId}/tasks`,
  USER_TASKS: (orgId: string, siteId: string, userId: string) => `/api/v1/organizations/${orgId}/sites/${siteId}/users/${userId}/tasks`,
  TASK_STATUS: (orgId: string, siteId: string, id: string) => `/api/v1/organizations/${orgId}/sites/${siteId}/tasks/${id}/status`,
  TASK_OVERRIDE: (orgId: string, siteId: string, id: string) => `/api/v1/organizations/${orgId}/sites/${siteId}/tasks/${id}/override`,
  TRADES: (orgId: string, siteId: string) => `/api/v1/organizations/${orgId}/sites/${siteId}/trades`,
  CHAT_MESSAGE: (orgId: string, siteId: string, userId: string, shiftId: string) => `/api/v1/organizations/${orgId}/sites/${siteId}/users/${userId}/sessions/shift/${shiftId}/message`,
  SCHEDULER_STATUS: '/api/v1/admin/scheduler/status',
  SCHEDULER_TRIGGER: '/api/v1/admin/scheduler/trigger',
  ALERTS: (orgId: string, siteId: string) => `/api/v1/organizations/${orgId}/sites/${siteId}/alerts`,
};

// Static UUID boundaries mappings for Dallas Store #1000 operational scopes
export const ORG_ID = '33333333-3333-3333-3333-333333333333';
export const SITE_ID = '55555555-5555-5555-5555-555555550000';
export const BYPASS_USER_ID = '00000000-0000-0000-0000-000000000000';
export const SHIFT_SESSION_ID = '11111111-1111-1111-1111-111111111111';

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
  
  let rawChecklist = t.checklist_state || t.ChecklistState || '';
  if (!rawChecklist && mappedTemplate) {
    const templateChecklist = mappedTemplate.checklist_template || mappedTemplate.ChecklistTemplate || '';
    rawChecklist = typeof templateChecklist === 'string' ? templateChecklist : JSON.stringify(templateChecklist);
  }
  if (!rawChecklist) {
    rawChecklist = '[]';
  }

  return {
    id: t.id || t.ID || '',
    task_template_id: t.task_template_id || t.TaskTemplateID || '',
    Task: mappedTemplate,
    execution_type: t.execution_type || t.ExecutionType || 'STANDARD',
    status: t.status || t.Status || 'PENDING',
    priority: t.priority !== undefined ? t.priority : (t.Priority !== undefined ? t.Priority : 3),
    description: t.description || t.Description || (mappedTemplate ? mappedTemplate.Description : ''),
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

// Stateless HTTP Client orchestrating all dynamic queries, network headers, and normalisation maps
export const ApiClient = {
  getAuthHeaders(token: string | null): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json'
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    return headers;
  },

  async probeReadiness(): Promise<{ client_id: string }> {
    const res = await fetch(ENDPOINTS.HEALTH_READINESS, {
      headers: { 'Content-Type': 'application/json' }
    });
    if (!res.ok) throw new ResponseError("Readiness checks failed", res.status);
    return res.json();
  },

  async fetchTasks(token: string | null, siteId: string = SITE_ID): Promise<TaskExecution[]> {
    const res = await fetch(ENDPOINTS.TASKS(ORG_ID, siteId), { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch active tasks failed", res.status);
    const data = await res.json();
    return Array.isArray(data) ? data.map(normalizeTask) : [];
  },

  async fetchUserProfile(token: string | null): Promise<any> {
    const res = await fetch(ENDPOINTS.ME(ORG_ID), { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch user profile failed", res.status);
    return res.json();
  },

  async fetchSites(token: string | null): Promise<any[]> {
    const res = await fetch(ENDPOINTS.SITES(ORG_ID), { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch active sites failed", res.status);
    return res.json();
  },

  async fetchUserTasks(token: string | null, siteId: string, userId: string): Promise<TaskExecution[]> {
    const res = await fetch(ENDPOINTS.USER_TASKS(ORG_ID, siteId, userId), { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch active user tasks failed", res.status);
    const data = await res.json();
    return Array.isArray(data) ? data.map(normalizeTask) : [];
  },

  async fetchUsers(token: string | null): Promise<any[]> {
    const res = await fetch(`/api/v1/admin/users`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch system users failed", res.status);
    return res.json();
  },

  async updateTaskStatus(token: string | null, taskId: string, status: string, checklistState: string): Promise<void> {
    const res = await fetch(ENDPOINTS.TASK_STATUS(ORG_ID, SITE_ID, taskId), {
      method: 'PATCH',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify({ status, checklist_state: checklistState })
    });
    if (!res.ok) throw new ResponseError("Update task GORM status failed", res.status);
  },

  async overrideAsset(token: string | null, taskId: string, assetId: string, justification: string): Promise<void> {
    const res = await fetch(ENDPOINTS.TASK_OVERRIDE(ORG_ID, SITE_ID, taskId), {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify({ asset_id: assetId, justification })
    });
    if (!res.ok) throw new ResponseError("Compliance override transaction failed", res.status);
  },

  async proposeTrade(token: string | null, taskId: string, proposedAssigneeId: string): Promise<void> {
    const res = await fetch(ENDPOINTS.TRADES(ORG_ID, SITE_ID), {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify({ task_execution_id: taskId, proposed_assignee_id: proposedAssigneeId })
    });
    if (!res.ok) throw new ResponseError("GORM trades ledger proposal failed", res.status);
  },

  async acceptTrade(token: string | null, tradeId: string): Promise<void> {
    const res = await fetch(`${ENDPOINTS.TRADES(ORG_ID, SITE_ID)}/${tradeId}/accept`, {
      method: 'POST',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("GORM trade acceptance failed", res.status);
  },

  async rejectTrade(token: string | null, tradeId: string): Promise<void> {
    const res = await fetch(`${ENDPOINTS.TRADES(ORG_ID, SITE_ID)}/${tradeId}/reject`, {
      method: 'POST',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("GORM trade rejection failed", res.status);
  },

  async postChatMessage(token: string | null, text: string): Promise<any> {
    const res = await fetch(ENDPOINTS.CHAT_MESSAGE(ORG_ID, SITE_ID, BYPASS_USER_ID, SHIFT_SESSION_ID), {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify({ message: text })
    });
    if (!res.ok) throw new ResponseError("Conversational agent message post failed", res.status);
    return res.json();
  },

  async fetchSchedulerStatus(token: string | null): Promise<any> {
    const res = await fetch(ENDPOINTS.SCHEDULER_STATUS, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Diagnostics polling status failed", res.status);
    return res.json();
  },

  async triggerSchedulerSweep(token: string | null): Promise<void> {
    const res = await fetch(ENDPOINTS.SCHEDULER_TRIGGER, {
      method: 'POST',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Forced background cron sweep trigger failed", res.status);
  },

  async triggerStreamingAlert(token: string | null, organizerID: string, eventType: string, description: string): Promise<any> {
    const res = await fetch(ENDPOINTS.ALERTS(ORG_ID, SITE_ID), {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify({ organizer_id: organizerID, event_type: eventType, description })
    });
    if (!res.ok) throw new ResponseError("Streaming alert trigger transaction failed", res.status);
    return res.json();
  }
};
