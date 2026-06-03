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

import React, { useRef, useEffect } from 'react';
import type { ChatMessage } from './types';
import A2UIRenderer from '../a2ui/A2UIRenderer';
import type { A2UIComponent, A2UIActionHandler } from '../a2ui/types';

interface ShiftCoachProps {
  chatLog: ChatMessage[];
  chatInput: string;
  setChatInput: React.Dispatch<React.SetStateAction<string>>;
  submitChatMsg: (customMsg?: string) => void;
  schedulerNodeID: string;
  schedulerLeader: boolean;
  schedulerTriggeredCount: number;
  forceStartTaskSweep: () => void;
  onA2UIActionTrigger: A2UIActionHandler;
  userName: string;
  userEmail: string;
}

const ShiftCoach: React.FC<ShiftCoachProps> = ({
  chatLog,
  chatInput,
  setChatInput,
  submitChatMsg,
  schedulerNodeID,
  schedulerLeader,
  schedulerTriggeredCount,
  forceStartTaskSweep,
  onA2UIActionTrigger,
  userName,
  userEmail
}: ShiftCoachProps) => {

  const chatEndRef = useRef<HTMLDivElement>(null);

  // Microtask post-paint smooth scrolling hook to defend against layout paint race conditions
  useEffect(() => {
    // Scroll instantly
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });

    // Schedule scroll after a short delay to guarantee layout calculations stabilize
    const timer = setTimeout(() => {
      chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, 100);

    return () => clearTimeout(timer);
  }, [chatLog]);

  return (
    <section className="panel-card span-right">
      <div className="panel-header" style={{ position: 'relative' }}>
        <h2 className="panel-title">Hanna - Operations Shift Coach</h2>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span className="panel-title-count" style={{ background: 'var(--priority-standard-glow)', color: 'var(--priority-standard)', fontFamily: 'var(--font-mono)' }}>
            {schedulerNodeID}
          </span>
          {schedulerLeader && (
            <span className="panel-title-count" style={{ background: 'var(--priority-critical-glow)', color: 'var(--priority-critical)' }}>
              LEADER
            </span>
          )}
        </div>
      </div>
      <div className="chat-wrapper">
        {/* Dynamic Chat Log Feed */}
        <div className="chat-log-feed">
          {chatLog.map((msg) => (
            <React.Fragment key={msg.id}>
              <div className={`chat-bubble ${msg.role === 'user' ? 'bubble-user' : 'bubble-agent'}`}>
                {msg.role === 'assistant' && <span className="agent-name-tag">Hanna</span>}
                <span style={{ whiteSpace: 'pre-line' }}>{msg.content}</span>
              </div>

              {/* Pure dynamic stateless A2UI component layout tree render dispatcher */}
              {msg.a2uiData && msg.a2uiData.type && (
                <A2UIRenderer
                  component={msg.a2uiData as A2UIComponent}
                  onActionTrigger={onA2UIActionTrigger}
                />
              )}

              {/* Legacy dynamic A2UI fallback templates compatibility layer */}
              {msg.a2uiType === 'VAULT_DROP' && (!msg.a2uiData || !msg.a2uiData.type) && (
                <div className="a2ui-card-container">
                  <div className="a2ui-header">
                    <span>CASH VAULT DROP VERIFICATION TICKET</span>
                    <span style={{ color: 'var(--priority-critical)', fontFamily: 'var(--font-mono)' }}>ALERT</span>
                  </div>
                  <div className="a2ui-body">
                    <table className="a2ui-table">
                      <tbody>
                        <tr>
                          <td className="a2ui-label">Register Channel</td>
                          <td className="a2ui-value">{msg.a2uiData.registerID}</td>
                        </tr>
                        <tr>
                          <td className="a2ui-label">Audit Ceiling</td>
                          <td className="a2ui-value" style={{ color: 'var(--priority-critical)' }}>{msg.a2uiData.currentCeiling}</td>
                        </tr>
                        <tr>
                          <td className="a2ui-label">Target Secure Pouch</td>
                          <td className="a2ui-value">{msg.a2uiData.pouchID}</td>
                        </tr>
                        <tr>
                          <td className="a2ui-label">Deposit Safe Vault</td>
                          <td className="a2ui-value">{msg.a2uiData.dropLocation}</td>
                        </tr>
                      </tbody>
                    </table>
                    <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 12 }}>
                      <button
                        type="button"
                        className="a2ui-btn-action"
                        onClick={() => onA2UIActionTrigger('OVERRIDE', msg.a2uiData)}
                      >
                        Force Vault Compliance Verify & Override
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {msg.a2uiType === 'TRADE' && (!msg.a2uiData || !msg.a2uiData.type) && (
                <div className="a2ui-card-container">
                  <div className="a2ui-header">
                    <span>PEER TASK TRADE PROPOSAL</span>
                    <span style={{ color: 'var(--accent-primary)', fontFamily: 'var(--font-mono)' }}>SWAP</span>
                  </div>
                  <div className="a2ui-body">
                    <div className="a2ui-trade-box">
                      <div className="a2ui-trade-row">
                        <span className="a2ui-label" style={{ flex: 1.2 }}>Proposing Coworker</span>
                        <div style={{ textAlign: 'right', flex: 2 }}>
                          <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{msg.a2uiData.target || 'Colleague Associate'}</div>
                        </div>
                      </div>
                      <div className="a2ui-trade-row">
                        <span className="a2ui-label" style={{ flex: 1.2 }}>Task Description</span>
                        <span className="a2ui-value" style={{ fontSize: '0.75rem', flex: 2, textAlign: 'right' }}>{msg.a2uiData.taskTitle || 'Retail Operations Handover'}</span>
                      </div>
                    </div>

                    <div style={{ display: 'flex', justifyContent: 'center', margin: '8px 0', color: 'var(--text-muted)' }}>
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                        <polyline points="17 1 21 5 17 9" />
                        <path d="M3 11V9a4 4 0 0 1 4-4h14" />
                        <polyline points="7 23 3 19 7 15" />
                        <path d="M21 13v2a4 4 0 0 1-4 4H3" />
                      </svg>
                    </div>

                    <div className="a2ui-trade-box">
                      <div className="a2ui-trade-row">
                        <span className="a2ui-label" style={{ flex: 1.2 }}>Proposed Assignee</span>
                        <div style={{ textAlign: 'right', flex: 2 }}>
                          <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{userName}</div>
                        </div>
                      </div>
                      <div className="a2ui-trade-row">
                        <span className="a2ui-label" style={{ flex: 1.2 }}>Assignee Role</span>
                        <span className="a2ui-value" style={{ fontSize: '0.75rem', flex: 2, textAlign: 'right' }}>{userEmail.includes('ryan') ? 'Shift Supervisor' : 'Floor Operations Associate'}</span>
                      </div>
                    </div>

                    <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 12 }}>
                      <button
                        type="button"
                        className="a2ui-btn-action"
                        onClick={() => onA2UIActionTrigger('TRADE_ACCEPT', { tradeID: msg.a2uiData.tradeID })}
                      >
                        Accept Trade Swap
                      </button>
                      <button
                        type="button"
                        className="a2ui-btn-action"
                        style={{ borderColor: 'var(--priority-critical)', color: 'var(--priority-critical)' }}
                        onClick={() => onA2UIActionTrigger('TRADE_DENY', { tradeID: msg.a2uiData.tradeID })}
                      >
                        Deny
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {msg.a2uiType === 'WEATHER' && (!msg.a2uiData || !msg.a2uiData.type) && (
                <div className="a2ui-card-container">
                  <div className="a2ui-header">
                    <span>METAR AIRPORT WIND AUDIT</span>
                    <span style={{ color: 'var(--accent-primary)', fontFamily: 'var(--font-mono)' }}>{msg.a2uiData.station}</span>
                  </div>
                  <div className="a2ui-body">
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                      <div className="site-meta-pill">Temp: <strong>{msg.a2uiData.temp}</strong></div>
                      <div className="site-meta-pill">Wind: <strong>{msg.a2uiData.wind}</strong></div>
                      <div className="site-meta-pill">Pressure: <strong>{msg.a2uiData.pressure}</strong></div>
                      <div className="site-meta-pill">Visibility: <strong>{msg.a2uiData.visibility}</strong></div>
                    </div>
                  </div>
                </div>
              )}
            </React.Fragment>
          ))}
          <div ref={chatEndRef} />
        </div>

        {/* suggestion chips */}
        <div className="quick-action-row">
          <button
            type="button"
            className="action-suggestion-chip"
            onClick={() => submitChatMsg("Register 4 exceeds Drop ceilings!")}
          >
            Till Drop pouch check
          </button>
          <button
            type="button"
            className="action-suggestion-chip"
            onClick={() => submitChatMsg("Verify active task trades")}
          >
            Task trade swap audit
          </button>
          <button
            type="button"
            className="action-suggestion-chip"
            onClick={() => submitChatMsg("Show Dallas METAR Weather check")}
          >
            DFW Airport Wind audit
          </button>
          <button
            type="button"
            className="action-suggestion-chip"
            style={{ borderColor: 'var(--priority-high)', color: 'var(--priority-high)' }}
            onClick={forceStartTaskSweep}
          >
            Force Cron sweep ({schedulerTriggeredCount})
          </button>
        </div>

        {/* Inputs bar */}
        <form
          className="chat-inputs-bar"
          onSubmit={(e) => {
            e.preventDefault();
            submitChatMsg();
          }}
        >
          <input
            type="text"
            className="chat-input-text"
            placeholder="Ask Hanna about SOP guidelines..."
            value={chatInput}
            onChange={(e) => setChatInput(e.target.value)}
          />
          <button type="submit" className="chat-btn-send">
            Send
          </button>
        </form>
      </div>
    </section>
  );
};

export default ShiftCoach;
