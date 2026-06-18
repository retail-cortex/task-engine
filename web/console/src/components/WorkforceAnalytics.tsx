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
import type { TaskExecution, ChecklistStep } from './types';

interface WorkforceAnalyticsProps {
  tasks: TaskExecution[];
  onExit: () => void;
}

const WorkforceAnalytics: React.FC<WorkforceAnalyticsProps> = ({ tasks, onExit }: WorkforceAnalyticsProps) => {
  const [expandedTaskId, setExpandedTaskId] = useState<string | null>(null);

  // 1. Filter completed tasks
  const completedRuns = tasks.filter(t => t.status === 'COMPLETED');

  // 2. Perform SLA compliance metrics calculations
  let totalDeltaSec = 0;
  let compliantCount = 0;
  let bottleneckCount = 0;

  const runsMetrics = completedRuns.map(t => {
    const start = t.started_at ? new Date(t.started_at).getTime() : 0;
    const end = t.completed_at ? new Date(t.completed_at).getTime() : 0;
    const durationSec = start && end ? Math.floor((end - start) / 1000) : 0;
    const pausedSec = t.total_paused_seconds || 0;
    const netSec = durationSec - pausedSec;
    const sloSec = t.Task ? t.Task.EstimatedDurationMinutes * 60 : 0;
    const delta = netSec - sloSec;

    if (netSec > 0 && sloSec > 0) {
      totalDeltaSec += delta;
      if (delta <= 0) compliantCount++;
      else bottleneckCount++;
    }

    return {
      task: t,
      netSec,
      sloSec,
      delta: netSec > 0 && sloSec > 0 ? delta : null
    };
  });

  const totalSloRuns = runsMetrics.filter(m => m.delta !== null).length;
  const avgDelta = totalSloRuns > 0 ? Math.round(totalDeltaSec / totalSloRuns) : 0;
  const complianceRate = totalSloRuns > 0 ? Math.round((compliantCount / totalSloRuns) * 100) : 0;

  const formatDelta = (seconds: number): string => {
    const abs = Math.abs(seconds);
    const m = Math.floor(abs / 60);
    const s = abs % 60;
    return m > 0 ? `${m}m ${s}s` : `${s}s`;
  };

  const toggleExpand = (id: string) => {
    setExpandedTaskId(expandedTaskId === id ? null : id);
  };

  return (
    <div className="admin-panel-overlay" style={{ padding: '24px', display: 'flex', flexDirection: 'column', gap: 24, overflowY: 'auto' }}>
      
      {/* Header Row */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 className="panel-title" style={{ fontSize: '1.75rem', margin: 0 }}>Workforce Compliance & SLO Analytics</h1>
          <p style={{ margin: '4px 0 0 0', color: 'var(--text-secondary)', fontSize: '0.85rem' }}>
            Real-time labor compliance audit, time-tracking, and SLA performance metrics.
          </p>
        </div>
        <button 
          onClick={onExit}
          style={{
            background: 'var(--bg-input)',
            border: '1px solid var(--panel-border)',
            color: 'var(--text-primary)',
            padding: '8px 16px',
            borderRadius: '8px',
            fontWeight: 600,
            cursor: 'pointer',
            transition: 'all 0.2s ease-in-out'
          }}
        >
          Close Dashboard
        </button>
      </div>

      {/* KPI Cards Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 20 }}>
        
        <div className="panel-card" style={{ padding: '18px', display: 'flex', flexDirection: 'column', gap: 6 }}>
          <span style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', letterSpacing: '0.05em' }}>
            COMPLIANCE COMPLETIONS
          </span>
          <span style={{ fontSize: '2rem', fontWeight: 900, color: 'var(--text-primary)' }}>
            {completedRuns.length}
          </span>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            Finished store operational runs
          </span>
        </div>

        <div className="panel-card" style={{ padding: '18px', display: 'flex', flexDirection: 'column', gap: 6 }}>
          <span style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', letterSpacing: '0.05em' }}>
            AVG COMPLIANCE DELTA
          </span>
          <span style={{ 
            fontSize: '2rem', 
            fontWeight: 900, 
            color: avgDelta <= 0 ? '#10b981' : '#ef4444' 
          }}>
            {avgDelta <= 0 ? `-${formatDelta(avgDelta)}` : `+${formatDelta(avgDelta)}`}
          </span>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            {avgDelta <= 0 ? 'Ahead of SLO targets (Efficient)' : 'Behind SLO targets (Overdue)'}
          </span>
        </div>

        <div className="panel-card" style={{ padding: '18px', display: 'flex', flexDirection: 'column', gap: 6 }}>
          <span style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', letterSpacing: '0.05em' }}>
            COMPLIANCE SLA RATE
          </span>
          <span style={{ 
            fontSize: '2rem', 
            fontWeight: 900, 
            color: complianceRate >= 80 ? '#10b981' : '#ec9f22' 
          }}>
            {complianceRate}%
          </span>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            Target threshold SLA is 85%
          </span>
        </div>

        <div className="panel-card" style={{ padding: '18px', display: 'flex', flexDirection: 'column', gap: 6 }}>
          <span style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-secondary)', letterSpacing: '0.05em' }}>
            LABOR BOTTLENECK RUNS
          </span>
          <span style={{ 
            fontSize: '2rem', 
            fontWeight: 900, 
            color: bottleneckCount === 0 ? 'var(--text-muted)' : '#ef4444' 
          }}>
            {bottleneckCount}
          </span>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            Runs exceeding target task durations
          </span>
        </div>

      </div>

      {/* Main labor compliance audit list */}
      <div className="panel-card" style={{ display: 'flex', flexDirection: 'column', padding: 0, overflow: 'hidden' }}>
        <div className="panel-header" style={{ borderBottom: '1px solid var(--panel-border)', padding: '16px 20px' }}>
          <h2 className="panel-title" style={{ margin: 0 }}>Labor Compliance Audit Trail</h2>
        </div>
        
        <div style={{ overflowX: 'auto' }}>
          {completedRuns.length === 0 ? (
            <div style={{ padding: '40px', textAlign: 'center', color: 'var(--text-muted)' }}>
              No completed operational task runs found for audit.
            </div>
          ) : (
            <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.9rem' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--panel-border)', color: 'var(--text-secondary)', fontWeight: 600, fontSize: '0.75rem', textTransform: 'uppercase' }}>
                  <th style={{ padding: '12px 20px' }}>Task Description</th>
                  <th style={{ padding: '12px 20px' }}>Associate</th>
                  <th style={{ padding: '12px 20px' }}>Net Duration</th>
                  <th style={{ padding: '12px 20px' }}>Target SLO</th>
                  <th style={{ padding: '12px 20px', textAlign: 'right' }}>SLA Performance Delta</th>
                </tr>
              </thead>
              <tbody>
                {runsMetrics.map(({ task, netSec, sloSec, delta }) => {
                  const isExpanded = expandedTaskId === task.id;
                  const isCompliant = delta !== null && delta <= 0;
                  
                  // Parse checklist steps
                  let checklistSteps: ChecklistStep[] = [];
                  if (task.checklist_state) {
                    try {
                      checklistSteps = JSON.parse(task.checklist_state);
                    } catch (e) {}
                  }

                  return (
                    <React.Fragment key={task.id}>
                      <tr 
                        onClick={() => toggleExpand(task.id)}
                        style={{ 
                          borderBottom: '1px solid var(--panel-border)', 
                          cursor: 'pointer',
                          background: isExpanded ? 'rgba(255,255,255,0.02)' : 'transparent',
                          transition: 'background 0.2s'
                        }}
                        className="audit-row-hover"
                      >
                        <td style={{ padding: '16px 20px', fontWeight: 600, color: 'var(--text-primary)' }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                            <span style={{ fontSize: '0.7rem', color: 'var(--text-muted)', transition: 'transform 0.2s', transform: isExpanded ? 'rotate(90deg)' : 'none' }}>
                              ▶
                            </span>
                            {task.Task ? task.Task.Name : task.task_template_id}
                          </div>
                        </td>
                        <td style={{ padding: '16px 20px', color: 'var(--text-primary)' }}>
                          👤 {task.Assignee ? task.Assignee.name : 'Unknown Associate'}
                        </td>
                        <td style={{ padding: '16px 20px', color: 'var(--text-secondary)' }}>
                          {formatDelta(netSec)} (Paused {formatDelta(task.total_paused_seconds || 0)})
                        </td>
                        <td style={{ padding: '16px 20px', color: 'var(--text-secondary)' }}>
                          {formatDelta(sloSec)}
                        </td>
                        <td style={{ padding: '16px 20px', textAlign: 'right' }}>
                          {delta !== null ? (
                            <span style={{ 
                              fontSize: '0.75rem', 
                              fontWeight: 700, 
                              padding: '4px 8px', 
                              borderRadius: 4, 
                              background: isCompliant ? 'rgba(16, 185, 129, 0.15)' : 'rgba(239, 68, 68, 0.15)',
                              color: isCompliant ? '#10b981' : '#ef4444'
                            }}>
                              {isCompliant ? `-${formatDelta(delta)} SLO` : `+${formatDelta(delta)} SLO`}
                            </span>
                          ) : (
                            <span style={{ color: 'var(--text-muted)' }}>No SLO</span>
                          )}
                        </td>
                      </tr>
                      
                      {/* Expanded Checklist Step-Level Audits Drawers */}
                      {isExpanded && (
                        <tr>
                          <td colSpan={5} style={{ padding: '0 0 16px 0', background: 'rgba(255,255,255,0.01)' }}>
                            <div style={{ padding: '16px 40px', borderBottom: '1px dashed var(--panel-border)' }}>
                              <h3 style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--accent-primary)', margin: '0 0 12px 0', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                                Detailed Step-Level Checklist Compliance Logs
                              </h3>
                              
                              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                                {checklistSteps.length === 0 ? (
                                  <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>No step logs recorded.</span>
                                ) : (
                                  checklistSteps.map((step, sIdx) => {
                                    const sDelta = step.slo_delta_seconds;
                                    const sCompliant = sDelta !== undefined && sDelta <= 0;

                                    return (
                                      <div 
                                        key={sIdx}
                                        style={{ 
                                          display: 'flex', 
                                          justifyContent: 'space-between', 
                                          alignItems: 'center',
                                          background: 'var(--bg-input)',
                                          border: '1px solid var(--panel-border)',
                                          padding: '10px 14px',
                                          borderRadius: 6,
                                          fontSize: '0.85rem'
                                        }}
                                      >
                                        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                                          <span style={{ fontWeight: 800, color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                                            STEP {step.step}
                                          </span>
                                          <span style={{ color: 'var(--text-primary)', fontWeight: 600 }}>
                                            {step.action}
                                          </span>
                                        </div>
                                        
                                        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                                          <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                                            Completed by: 👤 {step.completed_by_id === task.assignee_id ? (task.Assignee?.name || 'Assignee') : 'Supervisor'}
                                          </span>
                                          
                                          {sDelta !== undefined ? (
                                            <span style={{ 
                                              fontSize: '0.72rem', 
                                              fontWeight: 700, 
                                              padding: '2px 6px', 
                                              borderRadius: 4, 
                                              background: sCompliant ? 'rgba(16, 185, 129, 0.15)' : 'rgba(239, 68, 68, 0.15)',
                                              color: sCompliant ? '#10b981' : '#ef4444'
                                            }}>
                                              {sCompliant ? `-${formatDelta(sDelta)}` : `+${formatDelta(sDelta)}`}
                                            </span>
                                          ) : (
                                            <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>Pending</span>
                                          )}
                                        </div>
                                      </div>
                                    );
                                  })
                                )}
                              </div>
                            </div>
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      </div>

    </div>
  );
};

export default WorkforceAnalytics;
