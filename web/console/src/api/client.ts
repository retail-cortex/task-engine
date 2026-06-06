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
import { resolveChecklist, normalizeTask, decodeOAuthTokenClaims } from '../lib/utils';

export { resolveChecklist, normalizeTask, decodeOAuthTokenClaims };

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

// Stateful Session Context inside ApiClient
let activeToken: string | null = null;
let activeSiteId: string = SITE_ID;
let activeUserId: string = BYPASS_USER_ID;

export const ApiClient = {
  // App-level Context Setters
  setToken(token: string | null) {
    activeToken = token;
  },
  setActiveSiteId(siteId: string) {
    activeSiteId = siteId;
  },
  setActiveUserId(userId: string) {
    activeUserId = userId;
  },

  getAuthHeaders(token: string | null = activeToken): Record<string, string> {
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

  async fetchTasks(token: string | null = activeToken, siteId: string = activeSiteId): Promise<TaskExecution[]> {
    const res = await fetch(ENDPOINTS.TASKS(ORG_ID, siteId), { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch active tasks failed", res.status);
    const data = await res.json();
    return Array.isArray(data) ? data.map(normalizeTask) : [];
  },

  async fetchUserProfile(token: string | null = activeToken): Promise<any> {
    const res = await fetch(ENDPOINTS.ME(ORG_ID), { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch user profile failed", res.status);
    return res.json();
  },

  async fetchSites(token: string | null = activeToken): Promise<any[]> {
    const res = await fetch(ENDPOINTS.SITES(ORG_ID), { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch active sites failed", res.status);
    return res.json();
  },

  async fetchUserTasks(token: string | null = activeToken, siteId: string = activeSiteId, userId: string = activeUserId): Promise<TaskExecution[]> {
    const res = await fetch(ENDPOINTS.USER_TASKS(ORG_ID, siteId, userId), { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch active user tasks failed", res.status);
    const data = await res.json();
    return Array.isArray(data) ? data.map(normalizeTask) : [];
  },

  async fetchUsers(token: string | null = activeToken): Promise<any[]> {
    const res = await fetch(`/api/v1/admin/users`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch system users failed", res.status);
    return res.json();
  },

  async updateTaskStatus(token: string | null = activeToken, siteId: string = activeSiteId, taskId: string = '', status: string = '', checklistState: string = ''): Promise<void> {
    const res = await fetch(ENDPOINTS.TASK_STATUS(ORG_ID, siteId, taskId), {
      method: 'PATCH',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify({ status, checklist_state: checklistState })
    });
    if (!res.ok) throw new ResponseError("Update task status failed", res.status);
  },

  async overrideAsset(token: string | null = activeToken, siteId: string = activeSiteId, taskId: string = '', assetId: string = '', justification: string = ''): Promise<void> {
    const res = await fetch(ENDPOINTS.TASK_OVERRIDE(ORG_ID, siteId, taskId), {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify({ asset_id: assetId, justification })
    });
    if (!res.ok) throw new ResponseError("Compliance override transaction failed", res.status);
  },

  async claimTask(token: string | null = activeToken, siteId: string = activeSiteId, taskId: string = ''): Promise<void> {
    const res = await fetch(`/api/v1/organizations/${ORG_ID}/sites/${siteId}/tasks/${taskId}/claim`, {
      method: 'POST',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Task claim action failed", res.status);
  },

  async proposeTrade(token: string | null = activeToken, siteId: string = activeSiteId, taskId: string = '', proposedAssigneeId: string = ''): Promise<void> {
    const res = await fetch(ENDPOINTS.TRADES(ORG_ID, siteId), {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify({ task_execution_id: taskId, proposed_assignee_id: proposedAssigneeId })
    });
    if (!res.ok) throw new ResponseError("Trades ledger proposal failed", res.status);
  },

  async acceptTrade(token: string | null = activeToken, siteId: string = activeSiteId, tradeId: string = ''): Promise<void> {
    const res = await fetch(`${ENDPOINTS.TRADES(ORG_ID, siteId)}/${tradeId}/accept`, {
      method: 'POST',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Trade acceptance failed", res.status);
  },

  async rejectTrade(token: string | null = activeToken, siteId: string = activeSiteId, tradeId: string = ''): Promise<void> {
    const res = await fetch(`${ENDPOINTS.TRADES(ORG_ID, siteId)}/${tradeId}/reject`, {
      method: 'POST',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Trade rejection failed", res.status);
  },

  async postChatMessage(token: string | null = activeToken, text: string = '', siteId: string = activeSiteId, userId: string = activeUserId): Promise<any> {
    const res = await fetch(ENDPOINTS.CHAT_MESSAGE(ORG_ID, siteId, userId, SHIFT_SESSION_ID), {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify({ message: text })
    });
    if (!res.ok) throw new ResponseError("Conversational agent message post failed", res.status);
    return res.json();
  },

  async fetchSchedulerStatus(token: string | null = activeToken): Promise<any> {
    const res = await fetch(ENDPOINTS.SCHEDULER_STATUS, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Diagnostics polling status failed", res.status);
    return res.json();
  },

  async triggerSchedulerSweep(token: string | null = activeToken): Promise<void> {
    const res = await fetch(ENDPOINTS.SCHEDULER_TRIGGER, {
      method: 'POST',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Forced background cron sweep trigger failed", res.status);
  },

  async triggerStreamingAlert(token: string | null = activeToken, siteId: string = activeSiteId, organizerID: string = '', eventType: string = '', description: string = ''): Promise<any> {
    const res = await fetch(ENDPOINTS.ALERTS(ORG_ID, siteId), {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify({ organizer_id: organizerID, event_type: eventType, description })
    });
    if (!res.ok) throw new ResponseError("Streaming alert trigger transaction failed", res.status);
    return res.json();
  },

  // --- ADMIN API EXTENSIONS ---

  // Users
  async fetchUser(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/users/${id}`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch user failed", res.status);
    return res.json();
  },
  async createUser(user: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/users`, {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(user)
    });
    if (!res.ok) throw new ResponseError("Create user failed", res.status);
    return res.json();
  },
  async updateUser(id: string, user: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/users/${id}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(user)
    });
    if (!res.ok) throw new ResponseError("Update user failed", res.status);
    return res.json();
  },
  async deleteUser(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/users/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Delete user failed", res.status);
    return res.json();
  },
  async assignUserRole(id: string, roleId: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/users/${id}/roles`, {
      method: 'PUT',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify({ role_id: roleId })
    });
    if (!res.ok) throw new ResponseError("Assign role failed", res.status);
    return res.json();
  },

  // Roles
  async fetchRoles(token: string | null = activeToken): Promise<any[]> {
    const res = await fetch(`/api/v1/admin/roles`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch roles failed", res.status);
    return res.json();
  },
  async fetchRole(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/roles/${id}`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch role failed", res.status);
    return res.json();
  },
  async createRole(role: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/roles`, {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(role)
    });
    if (!res.ok) throw new ResponseError("Create role failed", res.status);
    return res.json();
  },
  async updateRole(id: string, role: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/roles/${id}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(role)
    });
    if (!res.ok) throw new ResponseError("Update role failed", res.status);
    return res.json();
  },
  async deleteRole(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/roles/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Delete role failed", res.status);
    return res.json();
  },

  // Organizations
  async fetchOrganizations(token: string | null = activeToken): Promise<any[]> {
    const res = await fetch(`/api/v1/admin/organizations`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch organizations failed", res.status);
    return res.json();
  },
  async fetchOrganization(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/organizations/${id}`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch organization failed", res.status);
    return res.json();
  },
  async createOrganization(org: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/organizations`, {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(org)
    });
    if (!res.ok) throw new ResponseError("Create organization failed", res.status);
    return res.json();
  },
  async updateOrganization(id: string, org: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/organizations/${id}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(org)
    });
    if (!res.ok) throw new ResponseError("Update organization failed", res.status);
    return res.json();
  },
  async deleteOrganization(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/organizations/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Delete organization failed", res.status);
    return res.json();
  },
  async assignUserToOrganization(orgId: string, userId: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/organizations/${orgId}/users/${userId}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Assign user to organization failed", res.status);
    return res.json();
  },

  // Sites
  async fetchSitesAdmin(token: string | null = activeToken): Promise<any[]> {
    const res = await fetch(`/api/v1/admin/sites`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch admin sites failed", res.status);
    return res.json();
  },
  async fetchSite(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/sites/${id}`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch site failed", res.status);
    return res.json();
  },
  async createSite(orgId: string, site: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/organizations/${orgId}/sites`, {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(site)
    });
    if (!res.ok) throw new ResponseError("Create site failed", res.status);
    return res.json();
  },
  async updateSite(id: string, site: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/sites/${id}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(site)
    });
    if (!res.ok) throw new ResponseError("Update site failed", res.status);
    return res.json();
  },
  async deleteSite(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/sites/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Delete site failed", res.status);
    return res.json();
  },

  // Locations
  async fetchLocations(token: string | null = activeToken): Promise<any[]> {
    const res = await fetch(`/api/v1/admin/locations`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch locations failed", res.status);
    return res.json();
  },
  async fetchLocation(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/locations/${id}`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch location failed", res.status);
    return res.json();
  },
  async createLocation(orgId: string, siteId: string, location: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/organizations/${orgId}/sites/${siteId}/locations`, {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(location)
    });
    if (!res.ok) throw new ResponseError("Create location failed", res.status);
    return res.json();
  },
  async updateLocation(id: string, location: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/locations/${id}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(location)
    });
    if (!res.ok) throw new ResponseError("Update location failed", res.status);
    return res.json();
  },
  async deleteLocation(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/locations/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Delete location failed", res.status);
    return res.json();
  },

  // Assets
  async fetchAssets(token: string | null = activeToken): Promise<any[]> {
    const res = await fetch(`/api/v1/admin/assets`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch assets failed", res.status);
    return res.json();
  },
  async fetchAsset(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/assets/${id}`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch asset failed", res.status);
    return res.json();
  },
  async createAsset(orgId: string, siteId: string, locationId: string, asset: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/organizations/${orgId}/sites/${siteId}/locations/${locationId}/assets`, {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(asset)
    });
    if (!res.ok) throw new ResponseError("Create asset failed", res.status);
    return res.json();
  },
  async updateAsset(id: string, asset: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/assets/${id}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(asset)
    });
    if (!res.ok) throw new ResponseError("Update asset failed", res.status);
    return res.json();
  },
  async deleteAsset(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/assets/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Delete asset failed", res.status);
    return res.json();
  },

  // Task Templates
  async fetchTaskTemplates(token: string | null = activeToken): Promise<any[]> {
    const res = await fetch(`/api/v1/admin/tasks/templates`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch task templates failed", res.status);
    return res.json();
  },
  async fetchTaskTemplate(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/tasks/templates/${id}`, { headers: this.getAuthHeaders(token) });
    if (!res.ok) throw new ResponseError("Fetch task template failed", res.status);
    return res.json();
  },
  async createTaskTemplate(template: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/tasks/templates`, {
      method: 'POST',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(template)
    });
    if (!res.ok) throw new ResponseError("Create task template failed", res.status);
    return res.json();
  },
  async updateTaskTemplate(id: string, template: any, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/tasks/templates/${id}`, {
      method: 'PUT',
      headers: this.getAuthHeaders(token),
      body: JSON.stringify(template)
    });
    if (!res.ok) throw new ResponseError("Update task template failed", res.status);
    return res.json();
  },
  async deleteTaskTemplate(id: string, token: string | null = activeToken): Promise<any> {
    const res = await fetch(`/api/v1/admin/tasks/templates/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders(token)
    });
    if (!res.ok) throw new ResponseError("Delete task template failed", res.status);
    return res.json();
  }
};
