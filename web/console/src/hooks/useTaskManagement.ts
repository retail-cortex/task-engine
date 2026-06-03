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

import { useState, useEffect } from 'react';
import type { TaskExecution, ChecklistStep } from '../components/types';
import { ApiClient, resolveChecklist } from '../api/client';

interface UseTaskManagementProps {
  activeSiteID: string;
  userRole: string;
  activeUserId: string;
  isAuthenticated: boolean;
  handleApiError: (err: any) => void;
  allUsers: any[];
}

export const useTaskManagement = ({
  activeSiteID,
  userRole,
  activeUserId,
  isAuthenticated,
  handleApiError,
  allUsers,
}: UseTaskManagementProps) => {
  const [tasks, setTasks] = useState<TaskExecution[]>([]);
  const [selectedTask, setSelectedTask] = useState<TaskExecution | null>(null);
  const [checklist, setChecklist] = useState<ChecklistStep[]>([]);
  const [selectedAssigneeFilter, setSelectedAssigneeFilter] = useState<string>("ALL");
  const [selectedRoleFilter, setSelectedRoleFilter] = useState<string>("ALL");

  useEffect(() => {
    setSelectedAssigneeFilter("ALL");
  }, [selectedRoleFilter]);

  const syncTasksList = (
    targetSiteID: string = activeSiteID,
    targetRole: string = userRole,
    targetUserID: string = activeUserId
  ) => {
    let taskPromise;
    if (targetRole === "ADMIN" || targetRole === "REGION_MANAGER" || targetRole === "SITE_MANAGER") {
      taskPromise = ApiClient.fetchTasks(null, targetSiteID);
    } else {
      taskPromise = ApiClient.fetchUserTasks(null, targetSiteID, targetUserID);
    }

    taskPromise
      .then((normalized) => {
        if (normalized && normalized.length > 0) {
          setTasks(normalized);
          
          // Select task matched to currently active selection index, fallback to first row
          const cachedId = localStorage.getItem('selected_task_id');
          const matched = normalized.find(t => t.id === cachedId) || normalized[0];
          setSelectedTask(matched);
          setChecklist(resolveChecklist(matched.checklist_state));
        } else {
          setTasks([]);
          setSelectedTask(null);
          setChecklist([]);
        }
      })
      .catch(handleApiError);
  };

  useEffect(() => {
    if (isAuthenticated) {
      syncTasksList(activeSiteID, userRole, activeUserId);
    } else {
      setTasks([]);
      setSelectedTask(null);
      setChecklist([]);
    }
  }, [activeSiteID, userRole, activeUserId, isAuthenticated]);

  const handleSelectTask = (task: TaskExecution) => {
    setSelectedTask(task);
    setChecklist(resolveChecklist(task.checklist_state));
    localStorage.setItem('selected_task_id', task.id);
  };

  const toggleChecklistStep = (stepIdx: number) => {
    if (!selectedTask) return;

    const updatedChecklist = checklist.map((step, idx) => {
      if (idx === stepIdx) {
        return { ...step, completed: !step.completed };
      }
      return step;
    });

    setChecklist(updatedChecklist);

    const updatedTask = { ...selectedTask, checklist_state: JSON.stringify(updatedChecklist) };
    
    const allDone = updatedChecklist.every(s => s.completed);
    updatedTask.status = allDone ? 'COMPLETED' : 'IN_PROGRESS';

    setSelectedTask(updatedTask);
    setTasks(tasks.map(t => t.id === updatedTask.id ? updatedTask : t));

    ApiClient.updateTaskStatus(undefined, undefined, selectedTask.id, updatedTask.status, updatedTask.checklist_state)
      .catch(handleApiError);
  };

  const getFilteredTasks = () => {
    let filtered = tasks;

    // Filter by Role first
    if (selectedRoleFilter !== "ALL") {
      filtered = filtered.filter(t => {
        const assignee = allUsers.find(u => u.id === t.assignee_id || u.ID === t.assignee_id);
        if (!assignee) return false;
        const roleNames = assignee.Roles ? assignee.Roles.map((r: any) => r.Name || r.name) : [];
        return roleNames.includes(selectedRoleFilter);
      });
    }

    // Filter by specific assignee
    if (selectedAssigneeFilter !== "ALL") {
      filtered = filtered.filter(t => t.assignee_id === selectedAssigneeFilter || t.locked_by === selectedAssigneeFilter);
    }

    return filtered;
  };

  return {
    tasks,
    selectedTask,
    checklist,
    selectedAssigneeFilter,
    setSelectedAssigneeFilter,
    selectedRoleFilter,
    setSelectedRoleFilter,
    syncTasksList,
    handleSelectTask,
    toggleChecklistStep,
    getFilteredTasks,
  };
};
