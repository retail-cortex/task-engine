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

import React, { useState } from 'react';
import type { TaskExecution } from './types';

interface TaskQueueProps {
  tasks: TaskExecution[];
  selectedTask: TaskExecution | null;
  onSelectTask: (task: TaskExecution) => void;
  onTradeTask?: (task: TaskExecution) => void;
  showAssigneeFilter?: boolean;
  activeCoworkers?: any[];
  selectedAssigneeFilter?: string;
  onAssigneeFilterChange?: (assigneeID: string) => void;
}

const TaskQueue: React.FC<TaskQueueProps> = ({
  tasks,
  selectedTask,
  onSelectTask,
  onTradeTask,
  showAssigneeFilter = false,
  activeCoworkers = [],
  selectedAssigneeFilter = "ALL",
  onAssigneeFilterChange
}: TaskQueueProps) => {

  // Local state tracking copied ID targets (enables beautiful post-copy dynamic visual confirmations!)
  const [copiedID, setCopiedID] = useState<string | null>(null);

  const getPriorityClass = (priority: number): string => {
    if (priority === 1) return 'priority-badge-critical';
    if (priority === 2) return 'priority-badge-high';
    return 'priority-badge-standard';
  };

  const getPriorityLabel = (priority: number): string => {
    if (priority === 1) return 'Critical';
    if (priority === 2) return 'High';
    return 'Standard';
  };

  const getCleanDescription = (desc: string): string => {
    if (desc.includes('[Grounded SOP Context]') || desc.includes('[Grounded SOP Compliance Context]')) {
      return desc.split(/\[Grounded SOP/)[0].trim();
    }
    return desc;
  };

  const handleCopyID = (e: React.MouseEvent, id: string) => {
    e.stopPropagation(); // Prevents copying UUID from selecting the parent task card layout!
    navigator.clipboard.writeText(id)
      .then(() => {
        setCopiedID(id);
        setTimeout(() => setCopiedID(null), 1500);
      })
      .catch(() => {});
  };

  return (
    <section className="panel-card span-left">
      <div className="panel-header" style={{ display: 'flex', flexDirection: 'column', gap: 10, alignItems: 'stretch' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
          <h2 className="panel-title">Operational Task Queue</h2>
          <span className="panel-title-count">{tasks.length} Active</span>
        </div>

        {showAssigneeFilter && activeCoworkers.length > 0 && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, background: 'var(--input-bg)', padding: '6px 12px', borderRadius: 8, border: '1px solid var(--panel-border)', width: '100%', marginTop: 4 }}>
            <span style={{ fontSize: '0.78rem', color: 'var(--text-secondary)', fontWeight: 700 }}>Assignee:</span>
            <select
              className="a2ui-form-select"
              style={{
                flex: 1,
                background: 'transparent',
                border: 'none',
                color: 'var(--text-primary)',
                fontSize: '0.78rem',
                fontWeight: 700,
                cursor: 'pointer',
                outline: 'none',
                padding: 0
              }}
              value={selectedAssigneeFilter}
              onChange={(e) => onAssigneeFilterChange && onAssigneeFilterChange(e.target.value)}
            >
              <option value="ALL" style={{ background: '#0c0e1c', color: 'var(--text-primary)' }}>All Site Coworkers</option>
              {activeCoworkers.map((u: any) => (
                <option key={u.id || u.ID} value={u.id || u.ID} style={{ background: '#0c0e1c', color: 'var(--text-primary)' }}>
                  {u.name || u.Name}
                </option>
              ))}
            </select>
          </div>
        )}
      </div>
      <div className="panel-body-scrollable">
        {tasks.map((t) => {
          const pClass = getPriorityClass(t.priority);
          const pLabel = getPriorityLabel(t.priority);
          const cleanDesc = getCleanDescription(t.description);

          return (
            <div
              key={t.id}
              className={`task-item-card ${selectedTask?.id === t.id ? 'active' : ''}`}
              onClick={() => onSelectTask(t)}
            >
              <div className="task-card-row1">
                <h3 className="task-title">{t.Task ? t.Task.Name : t.task_template_id}</h3>
                <span className={`task-priority-badge ${pClass}`}>
                  {pLabel}
                </span>
              </div>
              <p style={{ margin: 0, fontSize: '0.85rem', color: 'var(--text-secondary)', lineHeight: 1.4 }}>
                {cleanDesc}
              </p>
              <div className="task-details-footer">
                {/* Dynamic Copy-to-clipboard UUID Wrapper (Wow Aesthetics!) */}
                <div 
                  className="task-id-wrapper"
                  onClick={(e) => handleCopyID(e, t.id)}
                  title="Copy Full 36-character Task UUID ID"
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 6,
                    cursor: 'pointer',
                    color: copiedID === t.id ? 'var(--priority-standard)' : 'var(--text-secondary)',
                    transition: 'color 0.25s ease-in-out'
                  }}
                >
                  <span style={{ fontFamily: 'var(--font-mono)' }}>
                    {copiedID === t.id ? 'COPIED!' : `${t.id.substring(0, 15)}...`}
                  </span>
                  {copiedID === t.id ? (
                    // Success checkmark SVG icon!
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3.5" strokeLinecap="round" strokeLinejoin="round">
                      <polyline points="20 6 9 17 4 12" />
                    </svg>
                  ) : (
                    // Standard duplicate/copy SVG icon!
                    <svg 
                      width="12" 
                      height="12" 
                      viewBox="0 0 24 24" 
                      fill="none" 
                      stroke="currentColor" 
                      strokeWidth="2.5" 
                      strokeLinecap="round" 
                      strokeLinejoin="round"
                      style={{ opacity: 0.65 }}
                    >
                      <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                    </svg>
                  )}
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: 8, justifyContent: 'flex-end' }}>
                  {t.status !== 'COMPLETED' && t.status !== 'TRADE_PENDING' && (
                    <button
                      type="button"
                      className="task-trade-action-btn"
                      onClick={(e) => {
                        e.stopPropagation();
                        if (onTradeTask) onTradeTask(t);
                      }}
                      title="Initiate Peer Task Trade handover request"
                      style={{
                        background: 'transparent',
                        border: '1px solid var(--panel-border)',
                        borderRadius: 4,
                        color: 'var(--accent-primary)',
                        fontSize: '0.72rem',
                        padding: '2px 8px',
                        cursor: 'pointer',
                        fontWeight: 600,
                        letterSpacing: '0.02em',
                        transition: 'all 0.2s ease-in-out'
                      }}
                    >
                      TRADE
                    </button>
                  )}

                  {t.status === 'COMPLETED' ? (
                    <span style={{ color: 'var(--priority-standard)', fontWeight: 600 }}>COMPLETED</span>
                  ) : t.status === 'TRADE_PENDING' ? (
                    <span style={{ color: '#ec9f22', fontWeight: 600 }}>TRADE PENDING</span>
                  ) : t.status === 'IN_PROGRESS' ? (
                    <span style={{ color: 'var(--priority-high)', fontWeight: 600 }}>IN_PROGRESS</span>
                  ) : (
                    <span style={{ color: 'var(--text-muted)' }}>PENDING</span>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
};

export default TaskQueue;
