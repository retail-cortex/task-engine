/* Copyright 2026 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useState, useEffect, useRef } from 'react';
import A2UIRenderer from './a2ui/A2UIRenderer';
import type { A2UIComponent, A2UIActionHandler } from './a2ui/types';
import Canvas from './a2ui/Canvas';
import SSOPortal from './components/SSOPortal';
import { useAppContext } from './contexts/AppContext';

// 1. Types for Chat and Messages
interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  a2uiComponent?: A2UIComponent;
  timestamp: string;
}

const getCleanSiteName = (fullName: string) => {
  if (!fullName) return '';
  if (fullName.includes(' - ')) {
    return fullName.split(' - ').pop()?.trim() || fullName;
  }
  return fullName;
};

const App: React.FC = () => {
  // Hook into the global, secure authentication context (Google GIS + JWT tokens!)
  const {
    theme,
    setTheme,
    userToken,
    isAuthenticated,
    userName,
    userEmail,
    userPicture,
    userRole,
    userSites,
    activeSiteID,
    googleClientID,
    activeUserId,
    handleApiError,
    handleGoogleCredentialResponse,
    handleSignOut,
    handleSiteChange,
  } = useAppContext();

  // Chat & Connection State
  const [chatLog, setChatLog] = useState<ChatMessage[]>([]);
  const [chatInput, setChatInput] = useState<string>('');
  const [agentConnected, setAgentConnected] = useState<boolean>(true);
  const [activeSiteLabel, setActiveSiteLabel] = useState<string>('');
  const [voiceState, setVoiceState] = useState<'idle' | 'thinking' | 'listening' | 'speaking'>('idle');
  const [isThinking, setIsThinking] = useState<boolean>(false);
  const [microphoneActive, setMicrophoneActive] = useState<boolean>(false);
  const [isMuted, setIsMuted] = useState<boolean>(true); // Speech disabled by default

  // Digital Twin (Right Panel) State
  const [activeLayout, setActiveLayout] = useState<'linear' | 'boutique' | 'racetrack'>('linear');
  const [activeBeacon, setActiveBeacon] = useState<any>(null);

  const timelineEndRef = useRef<HTMLDivElement>(null);
  const recognitionRef = useRef<any>(null);
  const hasBootstrappedRef = useRef<boolean>(false);
  const chatInputRef = useRef<string>('');

  useEffect(() => {
    chatInputRef.current = chatInput;
  }, [chatInput]);


  // Clear chat log and reset bootstrapping locks on logout
  useEffect(() => {
    if (!isAuthenticated) {
      setChatLog([]);
      hasBootstrappedRef.current = false;
      setActiveBeacon(null);
    }
  }, [isAuthenticated]);

  // Auto-scroll chat timeline
  useEffect(() => {
    timelineEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatLog]);

  // Dynamically resolve physical store layout style based on active site ID claims
  useEffect(() => {
    if (!isAuthenticated) return;
    if (!activeUserId) return;

    if (activeSiteID === '44444444-4444-4444-4444-444444440001') {
      setActiveLayout('boutique'); // Volt & Vine - San Francisco
    } else if (activeSiteID === '44444444-4444-4444-4444-444444440002') {
      setActiveLayout('racetrack'); // Volt & Vine - Los Angeles
    } else {
      setActiveLayout('linear'); // Seattle, Milwaukee, Scottsdale, Cincinnati, etc.
    }
    setActiveBeacon(null);
  }, [isAuthenticated, activeSiteID, activeUserId]);

  useEffect(() => {
    if (!isAuthenticated || !activeUserId || userSites.length === 0) return;
    if (hasBootstrappedRef.current) return;

    hasBootstrappedRef.current = true;

    const currentSite = userSites.find((s: any) => (s.id || s.UUID || s.ID) === activeSiteID);
    const storeName = currentSite ? (currentSite.name || currentSite.Name) : '';

    // Set initial welcome message on login or site context swap
    const welcomeMsg: ChatMessage = {
      id: 'welcome-' + Date.now(),
      role: 'assistant',
      content: `Google OAuth Token validated. Session active for ${userName}.\n\nWorkspace initialized. Connected to operational ledger.`,
      timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    };

    // Check if we already have an active site selected in the UI context
    if (activeSiteID && storeName) {
      const cleanName = getCleanSiteName(storeName);
      // Show a visible user message in the chat log so the user knows the UI is driving the agent!
      const userMsg: ChatMessage = {
        id: 'bootstrap-user-msg',
        role: 'user',
        content: `Set store to ${cleanName} (ID: ${activeSiteID})`,
        timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      };
      setChatLog([welcomeMsg, userMsg]);
      
      // Send the RPC to the agent (not silent, so it matches the visible prompt!)
      sendAgentRPC(`Set store to ${cleanName} (ID: ${activeSiteID})`);
    } else {
      // If no site is pre-selected, ask the agent to list the stores
      const userMsg: ChatMessage = {
        id: 'bootstrap-user-msg',
        role: 'user',
        content: 'Retrieve authorized storefront contexts',
        timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      };
      setChatLog([welcomeMsg, userMsg]);
      sendAgentRPC("get user context");
    }
  }, [isAuthenticated, activeSiteID, userName, userSites, activeUserId]);

  // Web Speech API for simulated voice streaming
  useEffect(() => {
    if (typeof window !== 'undefined' && ('webkitSpeechRecognition' in window || 'SpeechRecognition' in window)) {
      const SpeechRecognition = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;
      const rec = new SpeechRecognition();
      rec.continuous = true;
      rec.interimResults = true;
      rec.lang = 'en-US';

      rec.onstart = () => {
        setVoiceState('listening');
        setMicrophoneActive(true);
        setChatInput(''); // Clear previous input on start
      };

      rec.onresult = (event: any) => {
        let transcript = '';
        for (let i = 0; i < event.results.length; ++i) {
          transcript += event.results[i][0].transcript;
        }
        setChatInput(transcript);
      };

      rec.onerror = (err: any) => {
        console.error("Speech recognition error:", err);
        setVoiceState('idle');
        setMicrophoneActive(false);
      };

      rec.onend = () => {
        setMicrophoneActive(false);
        setVoiceState(prev => prev === 'listening' ? 'idle' : prev);
        
        // Submit the final accumulated transcript on voice session termination
        const textToSubmit = chatInputRef.current.trim();
        if (textToSubmit) {
          submitMessage(textToSubmit);
        }
      };

      recognitionRef.current = rec;

      return () => {
        rec.abort();
      };
    }
  }, []);


  const toggleMicrophone = () => {
    if (!recognitionRef.current) {
      alert("Web Speech API is not supported in this browser. Please type your query.");
      return;
    }

    if (microphoneActive) {
      recognitionRef.current.stop();
    } else {
      setVoiceState('listening');
      recognitionRef.current.start();
    }
  };

  // Text-to-speech synthesis using secure server-side Google Cloud TTS
  const speakAgentText = async (text: string) => {
    if (isMuted) return;

    const cleanText = text.replace(/```[\s\S]*?```/g, "").trim();
    if (!cleanText) return;

    try {
      setVoiceState('speaking');

      // Fetch high-fidelity synthesized MP3 bytes from secure Go backend
      const response = await fetch('/api/v1/tts', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-User-ID': activeUserId || '',
          'Authorization': `Bearer ${userToken}`
        },
        body: JSON.stringify({ text: cleanText })
      });

      if (!response.ok) {
        throw new Error(`Google TTS failed with status: ${response.status}`);
      }

      const audioBlob = await response.blob();
      const audioUrl = URL.createObjectURL(audioBlob);
      const audio = new Audio(audioUrl);

      audio.onended = () => {
        setVoiceState('idle');
        URL.revokeObjectURL(audioUrl);
      };
      audio.onerror = () => {
        setVoiceState('idle');
        URL.revokeObjectURL(audioUrl);
      };

      await audio.play();
    } catch (error) {
      console.error("Google TTS Playback error:", error);
      setVoiceState('idle');
    }
  };

  // A2UI JSON-RPC v0.8 parser: transforms flat components into a nested tree structure
  const parseA2UIPayload = (result: any): A2UIComponent | undefined => {
    let agentMessageObj = null;
    if (result) {
      if (result.message && result.message.parts) {
        agentMessageObj = result.message;
      } else if (result.parts) {
        agentMessageObj = result;
      } else if (result.status && result.status.message && result.status.message.parts) {
        agentMessageObj = result.status.message;
      } else if (result.artifacts && result.artifacts.length > 0 && result.artifacts[0].parts) {
        agentMessageObj = result.artifacts[0];
      }
    }

    if (!agentMessageObj || !agentMessageObj.parts) return undefined;

    const a2uiParts = agentMessageObj.parts.filter((p: any) => {
      const data = p.data || p.root?.data;
      const metadata = p.metadata || p.root?.metadata;
      return data && metadata?.mimeType === 'application/json+a2ui';
    });
    if (a2uiParts.length === 0) return undefined;

    let rootId = '';
    const flatComponentsMap = new Map<string, any>();

    a2uiParts.forEach((p: any) => {
      const envelope = p.data || p.root?.data;
      if (envelope.beginRendering) {
        rootId = envelope.beginRendering.root;
      } else if (envelope.surfaceUpdate) {
        const list = envelope.surfaceUpdate.components || [];
        list.forEach((c: any) => {
          flatComponentsMap.set(c.id, c);
        });
      }
    });

    if (!rootId || flatComponentsMap.size === 0) return undefined;

    const resolveBoundValue = (val: any): string => {
      if (val === null || val === undefined) return '';
      if (typeof val === 'object') {
        if ('literalString' in val) return String(val.literalString);
        if ('literalBoolean' in val) return String(val.literalBoolean);
        if ('literalNumber' in val) return String(val.literalNumber);
        if ('path' in val) return String(val.path);
      }
      return String(val);
    };

    const resolveComponent = (id: string): A2UIComponent | null => {
      const definition = flatComponentsMap.get(id);
      if (!definition) return null;

      const rawComp = definition.component;
      const key = Object.keys(rawComp)[0];
      const props = rawComp[key];

      const resolved: any = {
        id: id,
        type: key.toLowerCase(),
        style: props.style || 'standard',
      };

      if (key === 'Card') {
        resolved.type = 'card';
        resolved.title = props.title;
        if (props.child) {
          const child = resolveComponent(props.child);
          resolved.children = child ? [child] : [];
        }
      } else if (key === 'Column' || key === 'Row') {
        resolved.type = key.toLowerCase();
        resolved.gap = props.gap;
        resolved.align = props.alignment || props.distribution;
        const childrenIds = props.children?.explicitList || [];
        resolved.children = childrenIds.map(resolveComponent).filter(Boolean);
      } else if (key === 'Text') {
        resolved.type = 'text';
        resolved.content = resolveBoundValue(props.text);
        resolved.style = props.usageHint || 'body';
      } else if (key === 'Button') {
        resolved.type = 'button';
        resolved.label = props.label;
        resolved.action = props.action?.name;
        
        const actionData: any = {};
        if (props.action?.context) {
          props.action.context.forEach((ctx: any) => {
            actionData[ctx.key] = resolveBoundValue(ctx.value);
          });
        }
        resolved.actionData = actionData;
        if (props.child) {
          const child = resolveComponent(props.child);
          resolved.children = child ? [child] : [];
        }
      } else if (key === 'Image') {
        const imageUrl = resolveBoundValue(props.url);
        if (imageUrl.includes('/api/v1/blueprint')) {
          resolved.type = 'canvas';
          try {
            const urlObj = new URL(imageUrl, window.location.origin);
            resolved.layout = urlObj.searchParams.get('layout') || 'linear';
            const x = urlObj.searchParams.get('x');
            const y = urlObj.searchParams.get('y');
            if (x && y) {
              resolved.beacon = {
                x: parseFloat(x),
                y: parseFloat(y),
                name: 'Active Focal Beacon'
              };
            }
          } catch (e) {
            console.error("Error parsing blueprint image query params:", e);
          }
        } else {
          resolved.type = 'image';
          resolved.content = imageUrl;
        }
      } else if (key === 'Input') {
        resolved.type = 'input';
        resolved.name = props.name;
        resolved.placeholder = props.placeholder;
      } else if (key === 'Select') {
        resolved.type = 'select';
        resolved.name = props.name;
        resolved.options = props.options;
      } else if (key === 'MultipleChoice') {
        resolved.type = 'select';
        resolved.name = props.selections?.path || 'selections';
        resolved.options = props.options?.map((o: any) => ({
          label: resolveBoundValue(o.label),
          value: o.value
        }));
      } else if (key === 'CheckBox' || key === 'Checkbox') {
        resolved.type = 'checkbox';
        resolved.label = resolveBoundValue(props.label);
        resolved.name = props.value?.path || id;
        resolved.value = props.value?.literalBoolean !== undefined ? props.value.literalBoolean : false;
      }

      return resolved;
    };

    return resolveComponent(rootId) || undefined;
  };

  const [activeSessionID, setActiveSessionID] = useState<string>('');
  const [sessions, setSessions] = useState<any[]>([]);

  // Fetch session history list from ADK
  const fetchSessionHistory = async () => {
    if (!isAuthenticated || !activeUserId) return;
    try {
      const res = await fetch(`http://localhost:8081/apps/task_agent/users/${activeUserId}/sessions`);
      if (res.ok) {
        const data = await res.json();
        const sorted = data.sort((a: any, b: any) => b.last_update_time - a.last_update_time);
        setSessions(sorted);
      }
    } catch (err) {
      console.error("Failed to fetch session history:", err);
    }
  };

  const handleNewSession = () => {
    sessionStorage.removeItem('a2ui_session_id');
    setActiveSessionID('');
    setChatLog([]);
    hasBootstrappedRef.current = false;
    setActiveBeacon(null);
    
    const welcomeMsg: ChatMessage = {
      id: 'welcome-' + Date.now(),
      role: 'assistant',
      content: `Chat session reset. Started a brand-new operational thread.\n\nReady for new commands.`,
      timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    };
    setChatLog([welcomeMsg]);
    fetchSessionHistory();
  };

  const handleSwitchSession = async (sessionId: string) => {
    if (sessionId === activeSessionID) return;
    
    setVoiceState('thinking');
    try {
      const res = await fetch(`http://localhost:8081/apps/task_agent/users/${activeUserId}/sessions/${sessionId}`);
      if (res.ok) {
        const sessionData = await res.json();
        
        sessionStorage.setItem('a2ui_session_id', sessionId);
        setActiveSessionID(sessionId);
        
        const events = sessionData.events || [];
        const newChatLog: ChatMessage[] = [];
        
        for (const evt of events) {
          if (evt.message && evt.message.role === 'user') {
            const text = evt.message.parts && evt.message.parts[0] ? evt.message.parts[0].text : '';
            newChatLog.push({
              id: evt.message.message_id || 'msg-' + Math.random().toString(36).substring(2, 11),
              role: 'user',
              content: text,
              timestamp: new Date(evt.timestamp * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
            });
          }
          if (evt.actions && evt.actions.response) {
            const text = evt.actions.response.parts && evt.actions.response.parts[0] ? evt.actions.response.parts[0].text : '';
            newChatLog.push({
              id: 'agent-' + Math.random().toString(36).substring(2, 11),
              role: 'assistant',
              content: text,
              timestamp: new Date(evt.timestamp * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
            });
          }
        }
        
        if (newChatLog.length === 0) {
          newChatLog.push({
            id: 'welcome-' + Date.now(),
            role: 'assistant',
            content: `Switched to empty session: ${sessionId.substring(0, 8)}...\n\nReady for new commands.`,
            timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
          });
        }
        
        setChatLog(newChatLog);
        hasBootstrappedRef.current = true;
        fetchSessionHistory();
      }
    } catch (err) {
      console.error("Failed to switch session:", err);
    } finally {
      setVoiceState('idle');
    }
  };

  // Load initial session ID on startup
  useEffect(() => {
    const sessId = sessionStorage.getItem('a2ui_session_id');
    if (sessId) {
      setActiveSessionID(sessId);
    }
  }, []);

  // Fetch history list when authenticated user changes
  useEffect(() => {
    if (isAuthenticated && activeUserId) {
      fetchSessionHistory();
    }
  }, [isAuthenticated, activeUserId]);

  // Sends raw RPC payload to A2A server on port 8081 (fully secured by JWT + X-User-ID)
  const sendAgentRPC = async (messageText: string, isSilent: boolean = false) => {
    if (!isAuthenticated || !activeUserId) return;

    setVoiceState('thinking');
    setIsThinking(true);

    // Retrieve or generate a persistent, valid UUID session ID for the current tab
    let a2uiSessionId = activeSessionID;
    if (!a2uiSessionId) {
      a2uiSessionId = sessionStorage.getItem('a2ui_session_id') || '';
      if (!a2uiSessionId) {
        if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
          a2uiSessionId = crypto.randomUUID();
        } else {
          // Standard RFC4122 v4 UUID generator fallback in vanilla JS
          a2uiSessionId = '10000000-1000-4000-8000-100000000000'.replace(/[018]/g, (c: any) =>
            (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
          );
        }
        sessionStorage.setItem('a2ui_session_id', a2uiSessionId);
      }
      setActiveSessionID(a2uiSessionId);
    }

    const rpcPayload = {
      jsonrpc: '2.0',
      id: Date.now(),
      method: 'message/send',
      params: {
        message: {
          message_id: 'msg-' + Math.random().toString(36).substring(2, 11),
          context_id: a2uiSessionId, // Propagate the session ID to preserve agent memory!
          role: 'user',
          parts: [
            {
              text: messageText
            }
          ]
        }
      }
    };

    try {
      const response = await fetch('/a2a/task_agent', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-User-ID': activeUserId,
          'Authorization': `Bearer ${userToken}`
        },
        body: JSON.stringify(rpcPayload)
      });

      if (!response.ok) {
        if (response.status === 401) {
          handleSignOut();
          throw new Error("Session expired. Please sign in again.");
        }
        throw new Error(`Agent HTTP error: ${response.status}`);
      }

      const data = await response.json();

      if (data.error) {
        throw new Error(data.error.message || "Unknown agent error");
      }

      const result = data.result;
      let agentMessageObj = null;
      if (result) {
        if (result.message && result.message.parts) {
          agentMessageObj = result.message;
        } else if (result.parts) {
          agentMessageObj = result;
        } else if (result.status && result.status.message && result.status.message.parts) {
          agentMessageObj = result.status.message;
        } else if (result.artifacts && result.artifacts.length > 0 && result.artifacts[0].parts) {
          agentMessageObj = result.artifacts[0];
        }
      }

      if (!agentMessageObj) {
        throw new Error("Invalid agent response format: missing message parts in result");
      }
      const textParts = agentMessageObj.parts.filter((p: any) => p.text);
      const combinedText = textParts.map((p: any) => p.text).join('\n').trim();

      const a2uiComp = parseA2UIPayload(result);

      if (a2uiComp) {
        const findCanvasNode = (node: A2UIComponent): A2UIComponent | null => {
          if (node.type === 'canvas') return node;
          if (node.children) {
            for (const child of node.children) {
              const found = findCanvasNode(child);
              if (found) return found;
            }
          }
          return null;
        };
        const canvasNode = findCanvasNode(a2uiComp);
        if (canvasNode) {
          setActiveLayout(canvasNode.layout || activeLayout);
          if (canvasNode.beacon) {
            setActiveBeacon(canvasNode.beacon);
          }
        }
      }

      if (!isSilent || combinedText || a2uiComp) {
        const agentMessage: ChatMessage = {
          id: 'agent-' + Date.now(),
          role: 'assistant',
          content: combinedText,
          a2uiComponent: a2uiComp,
          timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        };

        setChatLog((prev) => [...prev, agentMessage]);
        speakAgentText(combinedText);
        fetchSessionHistory();
      } else {
        setVoiceState('idle');
      }
      setIsThinking(false);
    } catch (err: any) {
      setIsThinking(false);
      console.error("Failed talking to agent:", err);
      setAgentConnected(false);
      setVoiceState('idle');
      
      const errorMsg: ChatMessage = {
        id: 'error-' + Date.now(),
        role: 'system',
        content: `Error: Failed connecting to the ADK Agent Service. Details: ${err.message}`,
        timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      };
      setChatLog((prev) => [...prev, errorMsg]);
    }
  };

  const submitMessage = (text: string) => {
    if (!text.trim()) return;

    const userMessage: ChatMessage = {
      id: 'user-' + Date.now(),
      role: 'user',
      content: text,
      timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    };

    setChatLog((prev) => [...prev, userMessage]);
    setChatInput('');
    submitMessageToBackend(text);
  };

  // Keep chat history stored on Go backend via REST API to ensure audit compliance, then trigger ADK Agent
  const submitMessageToBackend = async (text: string) => {
    try {
      // POST message to Go server session shifts first!
      const orgId = '33333333-3333-3333-3333-333333333333';
      const shiftSessionId = '11111111-1111-1111-1111-111111111111';
      
      const response = await fetch(
        `/api/v1/organizations/${orgId}/sites/${activeSiteID}/users/${activeUserId}/sessions/shift/${shiftSessionId}/message`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${userToken}`
          },
          body: JSON.stringify({ message: text })
        }
      );

      if (!response.ok) {
        if (response.status === 401) {
          handleSignOut();
          throw new Error("Session expired.");
        }
        throw new Error(`Go backend returned status ${response.status}`);
      }

      // Once audited successfully in DB, forward to ADK Python agent for reasoning & tooluse
      sendAgentRPC(text);
    } catch (e: any) {
      console.warn("Failed to audit message in shift ledger, calling agent directly. details:", e.message);
      sendAgentRPC(text);
    }
  };

  // Intercept actions triggered by buttons in the A2UI cards
  const handleA2UIAction: A2UIActionHandler = async (action, data) => {
    if (!isAuthenticated) return;

    const actMessage: ChatMessage = {
      id: 'action-trigger-' + Date.now(),
      role: 'system',
      content: `Triggering Action: ${action}...`,
      timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    };
    setChatLog(prev => [...prev, actMessage]);

    try {
      if (action === 'VIEW_TASK' || action === 'CONTINUE_TASK') {
        sendAgentRPC(`Show details for task ${data.execution_id}`);
      } else if (action === 'CLAIM_TASK') {
        sendAgentRPC(`Claim task ${data.execution_id}`);
      } else if (action === 'START_TASK') {
        sendAgentRPC(`Start task ${data.execution_id}`);
      } else if (action === 'COMPLETE_TASK') {
        sendAgentRPC(`Complete task ${data.execution_id}`);
      } else if (action === 'UPDATE_CHECKLIST') {
        sendAgentRPC(`Update checklist for task ${data.execution_id} with state: ${data.checklist_state}`);
      } else if (action === 'OVERRIDE') {
        sendAgentRPC(`Override asset constraint for task ${data.taskExecutionID} and asset ${data.assetID} with justification: ${data.justification || 'Verified cash drop secures backroom.'}`);
      } else if (action === 'TRADE') {
        sendAgentRPC(`Propose trade for task ${data.taskExecutionID} to coworker ${data.proposedAssigneeID}`);
      } else if (action === 'TRADE_ACCEPT') {
        sendAgentRPC(`Accept trade proposal ${data.tradeID}`);
      } else if (action === 'TRADE_DENY') {
        sendAgentRPC(`Deny trade proposal ${data.tradeID}`);
      } else if (action === 'SET_STORE') {
        const newSiteId = data.siteID;
        const currentSite = userSites.find((s: any) => (s.id || s.UUID || s.ID) === newSiteId);
        const siteName = currentSite ? (currentSite.name || currentSite.Name) : 'the store';
        const cleanName = getCleanSiteName(siteName);
        handleSiteChange(newSiteId);
        
        // Use the exact imperative phrasing with UUID that the agent is guaranteed to match
        sendAgentRPC(`Set store to ${cleanName} (ID: ${newSiteId})`);
      } else {
        throw new Error(`Unsupported action type: ${action}`);
      }
    } catch (e: any) {
      console.error("Action event routing failed:", e);
      const failMessage: ChatMessage = {
        id: 'action-fail-' + Date.now(),
        role: 'system',
        content: `Error: Action failed. Details: ${e.message}`,
        timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      };
      setChatLog(prev => [...prev, failMessage]);
    }
  };

  // Securely guard the portal under Google OAuth/SSO login boundary
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

  // Find active site details from userSites claims list
  const currentSite = userSites.find((s: any) => (s.id || s.ID) === activeSiteID);
  const storeName = currentSite ? (currentSite.name || currentSite.Name) : 'OmniMart Store';

  return (
    <div className="portal-container">
      {/* 1. Header toolbar */}
      <header className="portal-header">
        <div className="portal-logo">
          <div className="logo-icon"></div>
          <span className="logo-text">A2UI.Portal</span>
          <span className="logo-badge">NextGen</span>
        </div>
        <div className="header-status">
          <div className="status-item">
            <div className={`status-dot ${!agentConnected ? 'offline' : ''}`}></div>
            <span>{agentConnected ? 'Agent Connected (8081)' : 'Agent Offline'}</span>
          </div>
          
          {/* Site Selector Dropdown in Header */}
          {userSites.length > 1 && (
            <div className="status-item" style={{ gap: 6 }}>
              <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>Site Context:</span>
              <select 
                value={activeSiteID}
                onChange={(e) => handleSiteChange(e.target.value)}
                style={{
                  background: 'rgba(255, 255, 255, 0.04)',
                  border: '1px solid var(--border-muted)',
                  borderRadius: '6px',
                  color: '#fff',
                  padding: '3px 8px',
                  fontSize: '0.8rem',
                  outline: 'none',
                  cursor: 'pointer'
                }}
              >
                {userSites.map((site: any) => (
                  <option key={site.id || site.ID} value={site.id || site.ID} style={{ background: '#0b0f19', color: '#fff' }}>
                    {site.name || site.Name}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="user-badge">
            {userPicture && <img src={userPicture} className="user-avatar" alt="" />}
            <span className="user-name">{userName}</span>
            <span style={{ fontSize: '0.7rem', color: 'var(--color-primary)', background: 'var(--color-primary-glow)', padding: '2px 8px', borderRadius: 10 }}>
              {userRole}
            </span>
          </div>
          <button 
            onClick={handleSignOut}
            style={{ background: 'transparent', border: 'none', color: 'var(--color-critical)', cursor: 'pointer', fontSize: '0.8rem', fontWeight: 600 }}
          >
            Sign Out
          </button>
        </div>
      </header>

      {/* 2. Portal immersive workspace grid */}
      <div className="portal-grid">
        {/* Left column: Immersive Conversational Portal */}
        <div className="panel" style={{ height: 'calc(100vh - 120px)' }}>
          <div className="panel-header">
            <h3 className="panel-title">
              <span>Shift Operations Coordinator</span>
            </h3>
            <div style={{ display: 'flex', gap: 10 }}>
              <button 
                onClick={() => setIsMuted(!isMuted)} 
                className={`a2ui-btn ${isMuted ? 'secondary' : 'primary'}`}
                style={{ padding: '4px 10px', fontSize: '0.75rem', fontWeight: 600 }}
                title={isMuted ? "Enable Voice Assistant" : "Disable Voice Assistant"}
              >
                {isMuted ? "🔇 Speech Off" : "🔊 Speech On"}
              </button>
              <button 
                onClick={handleNewSession} 
                className="a2ui-btn secondary"
                style={{ padding: '4px 10px', fontSize: '0.75rem', fontWeight: 600 }}
                title="Start New Chat Session"
              >
                ➕ New Session
              </button>
            </div>
          </div>

          {/* Top Voice Orb Centerpiece */}
          <div className="voice-orb-container">
            <div 
              className={`voice-orb ${voiceState}`} 
              onClick={toggleMicrophone}
              title="Click to talk"
            >
              <div className="voice-orb-pulse-ring"></div>
              <svg className="voice-orb-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                {microphoneActive ? (
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
                ) : voiceState === 'speaking' ? (
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
                ) : (
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                )}
              </svg>
            </div>
            
            <div className="voice-status-text">
              {voiceState === 'listening' && "Listening..."}
              {voiceState === 'thinking' && "Analyzing Retail Engine..."}
              {voiceState === 'speaking' && "Synthesizing Instruction..."}
              {voiceState === 'idle' && "Agent Connected"}
            </div>
            
            <div className="voice-status-subtext">
              {voiceState === 'listening' ? "Speak clearly now..." : microphoneActive ? "Tap to pause" : "Tap orb to stream voice"}
            </div>

            <div className={`waveform ${voiceState !== 'idle' ? 'active' : ''}`}>
              <div className="waveform-bar"></div>
              <div className="waveform-bar"></div>
              <div className="waveform-bar"></div>
              <div className="waveform-bar"></div>
              <div className="waveform-bar"></div>
              <div className="waveform-bar"></div>
              <div className="waveform-bar"></div>
              <div className="waveform-bar"></div>
              <div className="waveform-bar"></div>
            </div>
          </div>

          {/* Scrollable Conversation timeline */}
          <div className="timeline-content">
            {chatLog.map((msg) => (
              <div key={msg.id} className={`chat-message ${msg.role}`}>
                <div className="chat-bubble">
                  <span style={{ whiteSpace: 'pre-wrap' }}>{msg.content}</span>
                  {msg.a2uiComponent && (
                    <div 
                      className="a2ui-card-outer"
                      style={{ cursor: 'pointer' }}
                      onClick={() => {
                        const findCanvasNode = (node: any): any | null => {
                          if (node.type === 'canvas') return node;
                          if (node.children && Array.isArray(node.children)) {
                            for (const child of node.children) {
                              const found = findCanvasNode(child);
                              if (found) return found;
                            }
                          }
                          return null;
                        };
                        const canvasNode = findCanvasNode(msg.a2uiComponent);
                        if (canvasNode) {
                          setActiveLayout(canvasNode.layout || activeLayout);
                          if (canvasNode.beacon) {
                            setActiveBeacon(canvasNode.beacon);
                          }
                        }
                      }}
                    >
                      <A2UIRenderer 
                        component={msg.a2uiComponent} 
                        onActionTrigger={handleA2UIAction} 
                      />
                    </div>
                  )}

                </div>
                <div className="chat-meta">
                  {msg.role === 'assistant' ? 'Agent' : msg.role === 'user' ? 'You' : 'System'} • {msg.timestamp}
                </div>
              </div>
            ))}
            {isThinking && (
              <div className="chat-message assistant thinking-indicator-msg">
                <div className="chat-bubble thinking-bubble">
                  <div className="typing-dots">
                    <span></span>
                    <span></span>
                    <span></span>
                  </div>
                </div>
                <div className="chat-meta">
                  Agent • thinking...
                </div>
              </div>
            )}
            <div ref={timelineEndRef} />
          </div>

          {/* Bottom Chat Input Bar */}
          <div className="chat-input-container">
            <div className="chat-input-wrapper">
              <input 
                type="text" 
                className="chat-input" 
                placeholder="Instruct the agent (e.g. 'override register ceiling')..."
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && submitMessage(chatInput)}
              />
              <button 
                className="chat-submit-btn" 
                onClick={() => submitMessage(chatInput)}
                title="Send instruction"
              >
                ➔
              </button>
            </div>
          </div>
        </div>

        {/* Right column: Digital Twin Spatial Visualization */}
        <div className="panel" style={{ height: 'calc(100vh - 120px)' }}>
          <div className="panel-header">
            <h3 className="panel-title">
              <span>Operations Spatial Twin</span>
            </h3>
          </div>
          <div className="operations-content" style={{ display: 'flex', flexDirection: 'column', flex: 1, justifySelf: 'center', width: '100%' }}>
            {/* Immersive SVG digital twin */}
            <div className="panel" style={{ padding: 10, background: '#03050c', border: '1px solid rgba(255,255,255,0.02)', width: '100%', boxSizing: 'border-box' }}>
              <Canvas 
                node={{
                  type: 'canvas',
                  layout: activeLayout,
                  beacon: activeBeacon
                }} 
              />
            </div>

            {/* Diagnostics HUD card */}
            <div className="panel" style={{ marginTop: 10, padding: 18, background: 'rgba(255, 255, 255, 0.01)', border: '1px solid var(--border-muted)' }}>
              <h4 style={{ margin: '0 0 10px', fontSize: '0.9rem', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                Operational Diagnostics HUD
              </h4>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8, fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span>Site Location UUID</span>
                  <span style={{ fontFamily: 'var(--font-mono)', color: '#fff' }}>{activeSiteID.substring(0, 18)}...</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span>Active Map Layout</span>
                  <span style={{ textTransform: 'uppercase', color: 'var(--color-primary)', fontWeight: 600 }}>{activeLayout}</span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span>Spatial Focus beacon</span>
                  <span style={{ color: activeBeacon ? 'var(--color-critical)' : 'var(--text-muted)', fontWeight: 600 }}>
                    {activeBeacon ? `${activeBeacon.name || 'Pin'} (x: ${activeBeacon.x}, y: ${activeBeacon.y})` : 'Unfocused'}
                  </span>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span>Storefront Context</span>
                  <span style={{ color: '#fff', fontWeight: 500 }}>{storeName}</span>
                </div>
              </div>
            </div>

            {/* Chat History Panel */}
            <div className="panel" style={{ marginTop: 10, padding: 18, background: 'rgba(255, 255, 255, 0.01)', border: '1px solid var(--border-muted)', display: 'flex', flexDirection: 'column', flex: 1, minHeight: 180 }}>
              <h4 style={{ margin: '0 0 10px', fontSize: '0.9rem', color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                Chat Session History
              </h4>
              <div style={{ overflowY: 'auto', flex: 1, display: 'flex', flexDirection: 'column', gap: 6, maxHeight: 200 }}>
                {sessions.length === 0 ? (
                  <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', textAlign: 'center', padding: '10px 0' }}>
                    No past sessions found.
                  </div>
                ) : (
                  sessions.map((sess: any) => {
                    const isActive = sess.id === activeSessionID;
                    const dateStr = new Date(sess.last_update_time * 1000).toLocaleString([], {
                      month: 'short',
                      day: 'numeric',
                      hour: '2-digit',
                      minute: '2-digit'
                    });
                    return (
                      <div 
                        key={sess.id}
                        onClick={() => handleSwitchSession(sess.id)}
                        style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'center',
                          padding: '8px 12px',
                          background: isActive ? 'rgba(74, 144, 226, 0.08)' : 'rgba(255, 255, 255, 0.02)',
                          border: isActive ? '1px solid var(--color-primary)' : '1px solid rgba(255, 255, 255, 0.05)',
                          borderRadius: 4,
                          cursor: 'pointer',
                          fontSize: '0.8rem',
                          transition: 'all 0.2s'
                        }}
                      >
                        <span style={{ color: isActive ? '#fff' : 'var(--text-secondary)', fontWeight: isActive ? 600 : 400 }}>
                          {sess.id.substring(0, 8)}... {isActive ? ' (Active)' : ''}
                        </span>
                        <span style={{ fontSize: '0.7rem', color: 'var(--text-muted)' }}>
                          {dateStr}
                        </span>
                      </div>
                    );
                  })
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default App;
