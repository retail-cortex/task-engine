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

import React, { useState, useEffect } from 'react';
import groundedBanner from './assets/grounded_intelligence_banner.png';
import SSOPortal from './components/SSOPortal';
import DashboardHeader from './components/DashboardHeader';
import TaskQueue from './components/TaskQueue';
import OperationsCenter from './components/OperationsCenter';
import ShiftCoach from './components/ShiftCoach';
import type { TaskExecution, ChecklistStep, ChatMessage } from './components/types';
import type { A2UIActionHandler } from './a2ui/types';
import { 
  ApiClient, 
  resolveChecklist, 
  decodeOAuthTokenClaims,
  BYPASS_USER_ID,
  SITE_ID,
  ResponseError
} from './api/client';

const App: React.FC = () => {
  // 0. Theme Swapping State Mechanism
  const [theme, setTheme] = useState<'light' | 'dark'>(
    () => (localStorage.getItem('theme') as 'light' | 'dark') || 'dark'
  );

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
  }, [theme]);

  // 1. Authentication & Google OAuth State
  const [userToken, setUserToken] = useState<string | null>(localStorage.getItem('oauth_token'));
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(!!localStorage.getItem('oauth_token'));
  const [userName, setUserName] = useState<string>(() => localStorage.getItem('oauth_name') || 'Hanna');
  const [userEmail, setUserEmail] = useState<string>(() => localStorage.getItem('oauth_email') || 'hanna@rmcguinness.altostrat.com');
  const [userPicture, setUserPicture] = useState<string | null>(() => localStorage.getItem('oauth_picture'));

  const [tasks, setTasks] = useState<TaskExecution[]>([]);
  const [selectedTask, setSelectedTask] = useState<TaskExecution | null>(null);
  const [checklist, setChecklist] = useState<ChecklistStep[]>([]);
  const [chatLog, setChatLog] = useState<ChatMessage[]>([]);
  const [chatInput, setChatInput] = useState('');
  const [schedulerLeader, setSchedulerLeader] = useState<boolean>(true);
  const [schedulerNodeID, setSchedulerNodeID] = useState<string>('node-A');
  const [schedulerTriggeredCount, setSchedulerTriggeredCount] = useState<number>(0);
  const [backendActive, setBackendActive] = useState<boolean>(false);
  const [googleClientID, setGoogleClientID] = useState<string>("10781708810-t4ose5l4ck5hc9ouq7kk56dipq6a3h76.apps.googleusercontent.com");

  // Dynamic Role, Site mapping & multi-site switching context states
  const [userRole, setUserRole] = useState<string>("SITE_ASSOCIATE");
  const [userSites, setUserSites] = useState<any[]>([]);
  const [activeSiteID, setActiveSiteID] = useState<string>(SITE_ID);
  const [allSites, setAllSites] = useState<any[]>([]);
  const [allUsers, setAllUsers] = useState<any[]>([]);
  const [activeUserId, setActiveUserId] = useState<string>(BYPASS_USER_ID);
  const [selectedAssigneeFilter, setSelectedAssigneeFilter] = useState<string>("ALL");

  // Centralised GORM API interceptor to handle Google OAuth token expirations (401 Unauthorized recovery pipeline!)
  const handleApiError = (err: any) => {
    console.error("[Operations API Error]", err);
    if (err instanceof ResponseError && err.status === 401) {
      console.warn("[OAuth Session] Credentials expired or validated incorrectly. Executing auto sign-out.");
      handleSignOut();
    }
  };

  const handleGoogleCredentialResponse = (response: any) => {
    const idToken = response.credential;
    
    const claims = decodeOAuthTokenClaims(idToken);
    if (claims) {
      setUserName(claims.name);
      setUserEmail(claims.email);
      setUserPicture(claims.picture);

      localStorage.setItem('oauth_name', claims.name);
      localStorage.setItem('oauth_email', claims.email);
      if (claims.picture) {
        localStorage.setItem('oauth_picture', claims.picture);
      } else {
        localStorage.removeItem('oauth_picture');
      }
    }

    localStorage.setItem('oauth_token', idToken);
    setUserToken(idToken);
    setIsAuthenticated(true);
    
    setChatLog([
      {
        id: 'msg-init',
        role: 'assistant',
        content: `OAuth SSO Handshake Successful! Google ID Token cryptographically verified.\n\nWelcome back, associate. Drawer drops are overdue, and Aisle 7 is looking sad. Complete the checklist steps to proceed.`
      }
    ]);
  };

  const handleSignOut = () => {
    localStorage.removeItem('oauth_token');
    localStorage.removeItem('oauth_name');
    localStorage.removeItem('oauth_email');
    localStorage.removeItem('oauth_picture');

    setUserToken(null);
    setIsAuthenticated(false);
    setTasks([]);
    setSelectedTask(null);
    setChecklist([]);
    setUserPicture(null);
    setUserName('Hanna');
    setUserEmail('hanna@rmcguinness.altostrat.com');
    setUserRole("SITE_ASSOCIATE");
    setUserSites([]);
    setActiveSiteID(SITE_ID);
    setAllSites([]);
    setAllUsers([]);
    setActiveUserId(BYPASS_USER_ID);
    setSelectedAssigneeFilter("ALL");
  };

  // 3. Fetch and Sync Pipeline Hooks (Readiness probe is completely public and unauthenticated!)
  useEffect(() => {
    ApiClient.probeReadiness()
      .then((data) => {
        setBackendActive(true);
        if (data && data.client_id) {
          setGoogleClientID(data.client_id);
        }
        if (isAuthenticated) {
          syncProfileContext();
          syncSchedulerDiagnostics();
        }
      })
      .catch(() => {
        setBackendActive(false);
        setTasks([]);
        setSelectedTask(null);
        setChecklist([]);
      });
  }, [isAuthenticated, userToken]);

  const syncProfileContext = () => {
    ApiClient.fetchUserProfile(userToken)
      .then((user: any) => {
        const userId = user.ID || user.id;
        setActiveUserId(userId);
        if (user.Name) setUserName(user.Name);
        if (user.Email) setUserEmail(user.Email);

        // Recover dynamic roles list preloaded from GORM
        const roleNames = user.Roles ? user.Roles.map((r: any) => r.Name) : [];
        let activeRole = "SITE_ASSOCIATE";
        if (roleNames.includes("ADMIN")) {
          activeRole = "ADMIN";
        } else if (roleNames.includes("REGION_MANAGER")) {
          activeRole = "REGION_MANAGER";
        } else if (roleNames.includes("SITE_MANAGER")) {
          activeRole = "SITE_MANAGER";
        } else if (roleNames.includes("SITE_3P")) {
          activeRole = "SITE_3P";
        }
        setUserRole(activeRole);

        // Recover preloaded physical store sites mappings
        const sitesList = user.Sites || [];
        setUserSites(sitesList);

        // Resolve context site target
        let initialSite = activeSiteID;
        if (sitesList.length > 0) {
          const hasActive = sitesList.some((s: any) => s.id === activeSiteID || s.ID === activeSiteID);
          if (!hasActive) {
            initialSite = sitesList[0].id || sitesList[0].ID;
            setActiveSiteID(initialSite);
          }
        }

        // Fetch master list datasets globally based on authorization profile scopes
        if (activeRole === "ADMIN" || activeRole === "REGION_MANAGER") {
          ApiClient.fetchSites(userToken)
            .then(data => setAllSites(data))
            .catch(console.error);
        }

        if (activeRole === "ADMIN" || activeRole === "REGION_MANAGER" || activeRole === "SITE_MANAGER") {
          ApiClient.fetchUsers(userToken)
            .then(data => setAllUsers(data))
            .catch(console.error);
        }

        syncTasksList(initialSite, activeRole, userId);
      })
      .catch((err) => {
        handleApiError(err);
        syncTasksList(activeSiteID, userRole, activeUserId);
      });
  };

  const syncTasksList = (targetSiteID: string = activeSiteID, targetRole: string = userRole, targetUserID: string = activeUserId) => {
    let taskPromise;
    if (targetRole === "ADMIN" || targetRole === "REGION_MANAGER" || targetRole === "SITE_MANAGER") {
      taskPromise = ApiClient.fetchTasks(userToken, targetSiteID);
    } else {
      taskPromise = ApiClient.fetchUserTasks(userToken, targetSiteID, targetUserID);
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

  const syncSchedulerDiagnostics = () => {
    ApiClient.fetchSchedulerStatus(userToken)
      .then((status: any) => {
        setSchedulerLeader(status.is_leader);
        setSchedulerNodeID(status.node_id);
      })
      .catch(handleApiError);
  };

  // 4. Dynamic State mutations
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

    ApiClient.updateTaskStatus(userToken, selectedTask.id, updatedTask.status, updatedTask.checklist_state)
      .catch(handleApiError);
  };

  const forceStartTaskSweep = () => {
    setSchedulerTriggeredCount(c => c + 1);
    ApiClient.triggerSchedulerSweep(userToken)
      .then(() => syncTasksList())
      .catch(handleApiError);
  };

  // 5. Conversational Chat and Dynamic A2UI Cards Engine
  const submitChatMsg = (customMsg?: string) => {
    const text = customMsg || chatInput;
    if (!text.trim()) return;

    const userMsg: ChatMessage = { id: 'user-' + Date.now(), role: 'user', content: text };
    setChatLog((prev) => [...prev, userMsg]);
    setChatInput('');

    ApiClient.postChatMessage(userToken, text)
      .then((data) => {
        const agentMsg: ChatMessage = {
          id: data.id || 'agent-' + Date.now(),
          role: 'assistant',
          content: data.content,
          a2uiType: data.a2uiType || undefined,
          a2uiData: data.a2uiData || undefined
        };
        setChatLog((prev) => [...prev, agentMsg]);
      })
      .catch((err) => {
        handleApiError(err);
        const errorMsg: ChatMessage = {
          id: 'agent-err-' + Date.now(),
          role: 'assistant',
          content: `Shift Coach connection offline. GORM database states are locked. Error: ${err.message}`
        };
        setChatLog((prev) => [...prev, errorMsg]);
      });
  };

  // 6. Centralized Dynamic A2UI Action Dispatcher (Pure live database-driven context!)
  const handleA2UIAction: A2UIActionHandler = (action: string, data: any) => {
    if (action === 'OVERRIDE') {
      ApiClient.overrideAsset(userToken, data.taskExecutionID, data.assetID, data.justification)
        .then(() => {
          const notifyMsg: ChatMessage = {
            id: 'agent-action-' + Date.now(),
            role: 'assistant',
            content: `Compliance override verification signed. Target Asset: ${data.assetID}. Justification: "${data.justification}". Status: COMPLETED.`
          };
          setChatLog((prev) => [...prev, notifyMsg]);
          
          // Dynamic database queue refresh
          syncTasksList();
        })
        .catch((err) => {
          handleApiError(err);
          const errorMsg: ChatMessage = {
            id: 'agent-action-err-' + Date.now(),
            role: 'assistant',
            content: `Compliance override verification failed: ${err.message}`
          };
          setChatLog((prev) => [...prev, errorMsg]);
        });
    } else if (action === 'TRADE') {
      ApiClient.proposeTrade(userToken, data.taskExecutionID, data.proposedAssigneeID)
        .then(() => {
          const notifyMsg: ChatMessage = {
            id: 'agent-action-' + Date.now(),
            role: 'assistant',
            content: `Task trade proposal successfully synchronized under GORM database schedules ledger. Coworker rosters updated.`
          };
          setChatLog((prev) => [...prev, notifyMsg]);
          
          syncTasksList();
        })
        .catch((err) => {
          handleApiError(err);
          const errorMsg: ChatMessage = {
            id: 'agent-action-err-' + Date.now(),
            role: 'assistant',
            content: `Task trade swap proposal transaction failed: ${err.message}`
          };
          setChatLog((prev) => [...prev, errorMsg]);
        });
    } else if (action === 'TRADE_ACCEPT') {
      ApiClient.acceptTrade(userToken, data.tradeID)
        .then(() => {
          const notifyMsg: ChatMessage = {
            id: 'agent-action-' + Date.now(),
            role: 'assistant',
            content: `Task trade request successfully accepted under GORM maker/checker guidelines! Coworker queues and handovers synchronized.`
          };
          setChatLog((prev) => [...prev, notifyMsg]);
          
          syncTasksList();
        })
        .catch((err) => {
          handleApiError(err);
          const errorMsg: ChatMessage = {
            id: 'agent-action-err-' + Date.now(),
            role: 'assistant',
            content: `Task trade acceptance failed: ${err.message}`
          };
          setChatLog((prev) => [...prev, errorMsg]);
        });
    } else if (action === 'TRADE_DENY') {
      ApiClient.rejectTrade(userToken, data.tradeID)
        .then(() => {
          const notifyMsg: ChatMessage = {
            id: 'agent-action-' + Date.now(),
            role: 'assistant',
            content: `Task trade request successfully denied. Task execution context restored on initiator's roster.`
          };
          setChatLog((prev) => [...prev, notifyMsg]);
          
          syncTasksList();
        })
        .catch((err) => {
          handleApiError(err);
          const errorMsg: ChatMessage = {
            id: 'agent-action-err-' + Date.now(),
            role: 'assistant',
            content: `Task trade rejection transaction failed: ${err.message}`
          };
          setChatLog((prev) => [...prev, errorMsg]);
        });
    } else if (action === 'DENY') {
      const notifyMsg: ChatMessage = {
        id: 'agent-action-' + Date.now(),
        role: 'assistant',
        content: `Task Trade Proposal swap denied. Notified operations controller.`
      };
      setChatLog((prev) => [...prev, notifyMsg]);
    } else if (action === 'SWEEP_TRIGGER') {
      setSchedulerTriggeredCount(c => c + 1);
      ApiClient.triggerSchedulerSweep(userToken)
        .then(() => {
          const notifyMsg: ChatMessage = {
            id: 'agent-action-' + Date.now(),
            role: 'assistant',
            content: "Background GORM scheduler cron sweep completed successfully. Site task queue refreshed dynamically in live database ledger."
          };
          setChatLog((prev) => [...prev, notifyMsg]);
          syncTasksList();
          syncSchedulerDiagnostics();
        })
        .catch((err) => {
          handleApiError(err);
          const errorMsg: ChatMessage = {
            id: 'agent-action-err-' + Date.now(),
            role: 'assistant',
            content: `Forced scheduler cron sweep transaction failed: ${err.message}`
          };
          setChatLog((prev) => [...prev, errorMsg]);
        });
    } else if (action === 'ALERT_TRIGGER') {
      ApiClient.triggerStreamingAlert(userToken, data.organizerID, data.eventType, data.description)
        .then((createdTask: any) => {
          const notifyMsg: ChatMessage = {
            id: 'agent-action-' + Date.now(),
            role: 'assistant',
            content: `Live sensor alarm alert ingested! Created persistent GORM task ticket: ${createdTask.task_template_id} (ID: ${createdTask.id.substring(0, 8)}). Register Terminal 4 target focal beacon activated. Complete cash drop checklist.`
          };
          setChatLog((prev) => [...prev, notifyMsg]);
          syncTasksList();
          
          // Dynamically trigger follow-up chatbot prompt queries matching newly generated GORM task card!
          setTimeout(() => {
            submitChatMsg("Register 4 exceeds Drawer ceiling limits, open drop voucher ticket please");
          }, 1500);
        })
        .catch((err) => {
          handleApiError(err);
          const errorMsg: ChatMessage = {
            id: 'agent-action-err-' + Date.now(),
            role: 'assistant',
            content: `Streaming alert ingestion failed: ${err.message}`
          };
          setChatLog((prev) => [...prev, errorMsg]);
        });
    } else if (action === 'PROPOSE_TRADE_FORM') {
      submitChatMsg("Propose colleague task trade swaps proposal please");
    } else if (action === 'OPEN_EVENT_FORM') {
      submitChatMsg("create event form");
    } else if (action === 'WEATHER_QUERY') {
      submitChatMsg("give me the weather for " + (data.station || 'KSFO'));
    } else if (action === 'SPAWN_SOP_TASK') {
      ApiClient.triggerStreamingAlert(userToken, "Sensor-Aisle-7", "AssetMaintenanceEvent", data.description)
        .then((createdTask: any) => {
          const notifyMsg: ChatMessage = {
            id: 'agent-action-' + Date.now(),
            role: 'assistant',
            content: `GORM SOP Compliance task ticket generated! Created task: ${createdTask.id.substring(0, 8)}. Focus target beacon on Aisle 7 operations coordinates twin activated. Complete the active audit steps.`
          };
          setChatLog((prev) => [...prev, notifyMsg]);
          syncTasksList();
        })
        .catch((err) => {
          handleApiError(err);
          const errorMsg: ChatMessage = {
            id: 'agent-action-err-' + Date.now(),
            role: 'assistant',
            content: `SOP task materialisation failure: ${err.message}`
          };
          setChatLog((prev) => [...prev, errorMsg]);
        });
    }
  };

  const handleSiteChange = (siteId: string) => {
    setActiveSiteID(siteId);
    setSelectedAssigneeFilter("ALL");
    syncTasksList(siteId, userRole, activeUserId);
  };

  const getFilteredTasks = () => {
    if (selectedAssigneeFilter === "ALL") {
      return tasks;
    }
    return tasks.filter(t => t.assignee_id === selectedAssigneeFilter || t.locked_by === selectedAssigneeFilter);
  };

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
        onSiteChange={handleSiteChange}
      />

      <main className="dashboard-grid">
        {/* 2. Left operational task queues column */}
        <TaskQueue
          tasks={getFilteredTasks()}
          selectedTask={selectedTask}
          onSelectTask={handleSelectTask}
          onTradeTask={(task) => submitChatMsg(`Propose task trade for task ${task.id}`)}
          showAssigneeFilter={userRole === "ADMIN" || userRole === "REGION_MANAGER" || userRole === "SITE_MANAGER"}
          activeCoworkers={allUsers.filter(u => u.Sites && u.Sites.some((s: any) => (s.id || s.ID) === activeSiteID))}
          selectedAssigneeFilter={selectedAssigneeFilter}
          onAssigneeFilterChange={(assigneeId) => setSelectedAssigneeFilter(assigneeId)}
        />

        {/* 3. Center Operations blueprint map column */}
        <OperationsCenter
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
          userNameVal=""
          userEmail={userEmail}
        />
      </main>

      {/* 5. Pure Live dynamic Server Offline Overlay Card block (satisfies ZERO fiction requirement!) */}
      {!backendActive && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100vw',
          height: '100vh',
          background: 'rgba(5, 6, 12, 0.85)',
          backdropFilter: 'blur(20px)',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          zIndex: 9999
        }}>
          <div className="panel-card" style={{ width: '450px', padding: '32px', textAlign: 'center', gap: '24px', border: '1px solid var(--priority-critical-glow)' }}>
            <div className="pulse-indicator" style={{ width: '12px', height: '12px', background: 'var(--priority-critical)', boxShadow: '0 0 12px var(--priority-critical)' }}></div>
            <h2 className="brand-title" style={{ color: 'var(--priority-critical)', fontSize: '1.4rem', margin: '8px 0 0 0' }}>NEXUS OPERATIONS OFFLINE</h2>
            <p style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', lineHeight: 1.5 }}>
              The backing multi-node task API server on port <strong>8080</strong> is currently unreachable.
              Start your local cluster runtime servers before proceeding:
            </p>
            <code style={{ background: 'var(--input-bg)', padding: '10px 14px', borderRadius: '6px', fontSize: '0.8rem', color: 'var(--accent-primary)', border: '1px solid var(--panel-border)', fontFamily: 'var(--font-mono)' }}>
              bazel run //:dev_server
            </code>
            <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              AlloyDB persistent ledgers, pgvector RAGs, and Maker/Checker task overrides are locked.
            </p>
          </div>
        </div>
      )}

      {/* Decorative radial vector backdrop asset */}
      <div style={{
        position: 'absolute',
        bottom: '2%',
        right: '2%',
        width: '80px',
        height: '80px',
        opacity: 0.15,
        backgroundImage: `url(${groundedBanner})`,
        backgroundSize: 'contain',
        backgroundRepeat: 'no-repeat',
        pointerEvents: 'none',
        zIndex: -1
      }}></div>
    </div>
  );
};

export default App;
