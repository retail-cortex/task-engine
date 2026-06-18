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

import React from 'react';
import SSOPortal from './components/SSOPortal';
import DashboardHeader from './components/DashboardHeader';
import TaskQueue from './components/TaskQueue';
import OperationsCenter from './components/OperationsCenter';
import ShiftCoach from './components/ShiftCoach';
import OfflineOverlay from './components/OfflineOverlay';
import AdminPanel from './components/AdminPanel';
import WorkforceAnalytics from './components/WorkforceAnalytics';
import { useAppContext } from './contexts/AppContext';
import { useTaskManagement } from './hooks/useTaskManagement';
import { useChatOrchestrator } from './hooks/useChatOrchestrator';
import { ApiClient } from './api/client';

const App: React.FC = () => {
  const [currentPage, setCurrentPage] = React.useState<string>('dashboard');
  const {
    theme,
    setTheme,
    isAuthenticated,
    userName,
    userEmail,
    userPicture,
    userRole,
    userSites,
    activeSiteID,
    allSites,
    allUsers,
    activeUserId,
    backendActive,
    googleClientID,
    schedulerLeader,
    schedulerNodeID,
    schedulerTriggeredCount,
    setSchedulerTriggeredCount,
    handleApiError,
    handleGoogleCredentialResponse,
    handleSignOut,
    syncSchedulerDiagnostics,
    handleSiteChange,
    allOrganizations,
    activeOrgID,
    setActiveOrgID,
  } = useAppContext();

  const {
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
  } = useTaskManagement({
    activeSiteID,
    userRole,
    activeUserId,
    isAuthenticated,
    handleApiError,
    allUsers,
  });

  const onSiteChangeWrapper = (siteId: string) => {
    handleSiteChange(siteId);
    setSelectedAssigneeFilter("ALL");
  };

  const onOrgChangeWrapper = (orgId: string) => {
    setActiveOrgID(orgId);
    if (orgId === 'ALL') {
      if (allSites.length > 0) {
        handleSiteChange(allSites[0].id || allSites[0].ID);
      }
    } else {
      const orgSites = allSites.filter(s => (s.OrganizationID || s.organization_id) === orgId);
      if (orgSites.length > 0) {
        handleSiteChange(orgSites[0].id || orgSites[0].ID);
      }
    }
    setSelectedAssigneeFilter("ALL");
  };

  const handleTakeTask = (task: any) => {
    ApiClient.claimTask(null, activeSiteID, task.id)
      .then(() => {
        syncTasksList();
        const notifyMsg = {
          id: 'system-' + Date.now(),
          role: 'assistant' as const,
          content: `System Notification: User successfully claimed and took ownership of task: "${task.Task?.Name || task.task_template_id}" (ID: task-${task.id.substring(0, 8)}). Trade pending status resolved.`
        };
        setChatLog((prev) => [...prev, notifyMsg]);
      })
      .catch(handleApiError);
  };

  const {
    chatLog,
    setChatLog,
    chatInput,
    setChatInput,
    submitChatMsg,
    handleA2UIAction,
    forceStartTaskSweep,
  } = useChatOrchestrator({
    activeSiteID,
    activeUserId,
    userRole,
    syncTasksList,
    handleApiError,
    schedulerNodeID,
    schedulerLeader,
    schedulerTriggeredCount,
    setSchedulerTriggeredCount,
    syncSchedulerDiagnostics,
    onSiteChange: onSiteChangeWrapper,
    onRoleChange: (role) => setSelectedRoleFilter(role),
    onAssigneeChange: (assigneeId) => setSelectedAssigneeFilter(assigneeId),
  });

  // If user is not authenticated: render the Single Sign-On (SSO) Portal
  if (!isAuthenticated) {
    return (
      <SSOPortal
        theme={theme}
        setTheme={setTheme}
        googleClientID={googleClientID}
        handleGoogleCredentialResponse={handleGoogleCredentialResponse}
      />
    );
  }

  return (
    <div id="root">
      {/* 1. Master dashboard settings and variables header */}
      <DashboardHeader
        userName={userName}
        userEmail={userEmail}
        userPicture={userPicture}
        theme={theme}
        setTheme={setTheme}
        backendActive={backendActive}
        schedulerNodeID={schedulerNodeID}
        schedulerLeader={schedulerLeader}
        handleSignOut={handleSignOut}
        userRole={userRole}
        userSites={userSites}
        allSites={allSites}
        activeSiteID={activeSiteID}
        onSiteChange={onSiteChangeWrapper}
        allOrganizations={allOrganizations}
        activeOrgID={activeOrgID}
        onOrgChange={onOrgChangeWrapper}
        activeCoworkers={allUsers.filter(u => u.Sites && u.Sites.some((s: any) => (s.id || s.ID) === activeSiteID))}
        selectedAssigneeFilter={selectedAssigneeFilter}
        onAssigneeFilterChange={(assigneeId) => setSelectedAssigneeFilter(assigneeId)}
        selectedRoleFilter={selectedRoleFilter}
        onRoleFilterChange={(role) => setSelectedRoleFilter(role)}
        onNavigateToAdmin={() => setCurrentPage('admin')}
        onNavigateToAnalytics={() => setCurrentPage('analytics')}
      />

      {currentPage === 'admin' ? (
        <AdminPanel onExit={() => setCurrentPage('dashboard')} />
      ) : currentPage === 'analytics' ? (
        <WorkforceAnalytics tasks={tasks} onExit={() => setCurrentPage('dashboard')} />
      ) : (
        <main className="dashboard-grid">
          {/* 2. Left operational task queues column */}
          <TaskQueue
            tasks={getFilteredTasks()}
            selectedTask={selectedTask}
            onSelectTask={handleSelectTask}
            onTradeTask={(task) => submitChatMsg(`Propose task trade for task ${task.id}`)}
            onTakeTask={handleTakeTask}
          />

          {/* 3. Center Operations blueprint map column */}
          <OperationsCenter
            activeSiteID={activeSiteID}
            selectedTask={selectedTask}
            checklist={checklist}
            onToggleStep={toggleChecklistStep}
          />

          {/* 4. Right Shift Coach conversation feed column */}
          <ShiftCoach
            chatLog={chatLog}
            chatInput={chatInput}
            setChatInput={setChatInput}
            submitChatMsg={submitChatMsg}
            schedulerNodeID={schedulerNodeID}
            schedulerLeader={schedulerLeader}
            schedulerTriggeredCount={schedulerTriggeredCount}
            forceStartTaskSweep={forceStartTaskSweep}
            onA2UIActionTrigger={handleA2UIAction}
            userName={userName}
            userEmail={userEmail}
          />
        </main>
      )}

      {/* 5. Pure Live dynamic Server Offline Overlay Card block */}
      <OfflineOverlay backendActive={backendActive} />
    </div>
  );
};

export default App;
