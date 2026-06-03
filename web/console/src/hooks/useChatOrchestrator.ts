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

import { useState } from 'react';
import type { ChatMessage } from '../components/types';
import type { A2UIActionHandler } from '../a2ui/types';
import { ApiClient } from '../api/client';

interface UseChatOrchestratorProps {
  activeSiteID: string;
  activeUserId: string;
  userRole: string;
  syncTasksList: () => void;
  handleApiError: (err: any) => void;
  schedulerNodeID: string;
  schedulerLeader: boolean;
  schedulerTriggeredCount: number;
  setSchedulerTriggeredCount: React.Dispatch<React.SetStateAction<number>>;
  syncSchedulerDiagnostics: () => void;
  onSiteChange?: (siteId: string) => void;
  onRoleChange?: (role: string) => void;
  onAssigneeChange?: (assigneeId: string) => void;
}

export const useChatOrchestrator = ({
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
  onSiteChange,
  onRoleChange,
  onAssigneeChange,
}: UseChatOrchestratorProps) => {
  const [chatLog, setChatLog] = useState<ChatMessage[]>([
    {
      id: 'msg-init',
      role: 'assistant',
      content: `OAuth SSO Handshake Successful! Google ID Token cryptographically verified.\n\nWelcome back, associate. Drawer drops are overdue, and Aisle 7 is looking sad. Complete the checklist steps to proceed.`
    }
  ]);
  const [chatInput, setChatInput] = useState('');

  const submitChatMsg = (customMsg?: string) => {
    const text = customMsg || chatInput;
    if (!text.trim()) return;

    const userMsg: ChatMessage = { id: 'user-' + Date.now(), role: 'user', content: text };
    setChatLog((prev) => [...prev, userMsg]);
    setChatInput('');

    ApiClient.postChatMessage(null, text, activeSiteID, activeUserId)
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
          content: `Shift Coach connection offline. Database states are locked. Error: ${err.message}`
        };
        setChatLog((prev) => [...prev, errorMsg]);
      });
  };

  const handleA2UIAction: A2UIActionHandler = (action: string, data: any) => {
    if (action === 'OVERRIDE') {
      ApiClient.overrideAsset(undefined, undefined, data.taskExecutionID, data.assetID, data.justification)
        .then(() => {
          const notifyMsg: ChatMessage = {
            id: 'agent-action-' + Date.now(),
            role: 'assistant',
            content: `Compliance override verification signed. Target Asset: ${data.assetID}. Justification: "${data.justification}". Status: COMPLETED.`
          };
          setChatLog((prev) => [...prev, notifyMsg]);
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
      ApiClient.proposeTrade(undefined, undefined, data.taskExecutionID, data.proposedAssigneeID)
        .then(() => {
          const notifyMsg: ChatMessage = {
            id: 'agent-action-' + Date.now(),
            role: 'assistant',
            content: `Task trade proposal successfully synchronized under system schedules ledger. Coworker rosters updated.`
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
      ApiClient.acceptTrade(undefined, undefined, data.tradeID)
        .then(() => {
          const notifyMsg: ChatMessage = {
            id: 'agent-action-' + Date.now(),
            role: 'assistant',
            content: `Task trade request successfully accepted! Coworker queues and handovers synchronized.`
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
      ApiClient.rejectTrade(undefined, undefined, data.tradeID)
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
      ApiClient.triggerSchedulerSweep()
        .then(() => {
          const notifyMsg: ChatMessage = {
            id: 'agent-action-' + Date.now(),
            role: 'assistant',
            content: "Background scheduler sweep completed successfully. Site task queue refreshed dynamically in database ledger."
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
      ApiClient.triggerStreamingAlert(undefined, undefined, data.organizerID, data.eventType, data.description)
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
    } else if (action === 'OPEN_WEATHER_FORM') {
      submitChatMsg("airport codes");
    } else if (action === 'OPEN_STORE_SELECTOR') {
      submitChatMsg("select store");
    } else if (action === 'OPEN_ROLE_SELECTOR') {
      submitChatMsg("filter by role");
    } else if (action === 'OPEN_ASSIGNEE_SELECTOR') {
      submitChatMsg("filter by assignee");
    } else if (action === 'OPEN_SOP_SEARCH') {
      submitChatMsg("sop guidelines");
    } else if (action === 'WEATHER_QUERY') {
      submitChatMsg("give me the weather for " + (data.station || 'KSFO'));
    } else if (action === 'SPAWN_SOP_TASK') {
      ApiClient.triggerStreamingAlert(undefined, undefined, "Sensor-Aisle-7", "AssetMaintenanceEvent", data.description)
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
    } else if (action === 'SET_STORE') {
      if (onSiteChange && data.siteID) {
        onSiteChange(data.siteID);
        const notifyMsg: ChatMessage = {
          id: 'agent-action-' + Date.now(),
          role: 'assistant',
          content: `Context site successfully updated to storefront: "${data.siteLabel || data.siteID}". Task queue and associate roster synchronized.`
        };
        setChatLog((prev) => [...prev, notifyMsg]);
      }
    } else if (action === 'SET_ROLE') {
      if (onRoleChange && data.role) {
        onRoleChange(data.role);
        const notifyMsg: ChatMessage = {
          id: 'agent-action-' + Date.now(),
          role: 'assistant',
          content: `Context role filter successfully updated to: "${data.roleLabel || data.role}". Queue scoped.`
        };
        setChatLog((prev) => [...prev, notifyMsg]);
      }
    } else if (action === 'SET_ASSIGNEE') {
      if (onAssigneeChange && data.assigneeID) {
        onAssigneeChange(data.assigneeID);
        const notifyMsg: ChatMessage = {
          id: 'agent-action-' + Date.now(),
          role: 'assistant',
          content: `Context assignee filter successfully updated to coworker: "${data.assigneeName || data.assigneeID}".`
        };
        setChatLog((prev) => [...prev, notifyMsg]);
      }
    }
  };

  const forceStartTaskSweep = () => {
    setSchedulerTriggeredCount(c => c + 1);
    ApiClient.triggerSchedulerSweep()
      .then(() => syncTasksList())
      .catch(handleApiError);
  };

  return {
    chatLog,
    setChatLog,
    chatInput,
    setChatInput,
    submitChatMsg,
    handleA2UIAction,
    forceStartTaskSweep,
  };
};
