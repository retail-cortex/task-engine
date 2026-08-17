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
import { ApiClient } from '../../api/client';

interface TaskMainProps {
  onEdit: (template: any) => void;
  onCreate: () => void;
  onError: (err: any) => void;
}

export const TaskMain: React.FC<TaskMainProps> = ({ onEdit, onCreate, onError }) => {
  const [templates, setTemplates] = useState<any[]>([]);
  const [roles, setRoles] = useState<any[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedPriority, setSelectedPriority] = useState('ALL');
  const [loading, setLoading] = useState(true);

  const loadData = async () => {
    try {
      setLoading(true);
      const [fetchedTemplates, fetchedRoles] = await Promise.all([
        ApiClient.fetchTaskTemplates(),
        ApiClient.fetchRoles()
      ]);
      setTemplates(fetchedTemplates || []);
      setRoles(fetchedRoles || []);
    } catch (err) {
      onError(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this task template? All currently scheduled sweeps of this task will be cancelled.')) return;
    try {
      await ApiClient.deleteTaskTemplate(id);
      loadData();
    } catch (err) {
      onError(err);
    }
  };

  const getRoleName = (roleId: string | null) => {
    if (!roleId) return 'All Personnel';
    const r = roles.find((x) => x.ID === roleId);
    return r ? r.Name : 'All Personnel';
  };

  const getPriorityLabel = (priority: number) => {
    if (priority === 1) return 'Critical';
    if (priority === 2) return 'High';
    return 'Standard';
  };

  const getPriorityBadgeClass = (priority: number) => {
    if (priority === 1) return 'priority-badge-critical';
    if (priority === 2) return 'priority-badge-high';
    return 'priority-badge-standard';
  };

  const filteredTemplates = templates.filter((t) => {
    const matchesSearch =
      (t.Name && t.Name.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (t.Description && t.Description.toLowerCase().includes(searchTerm.toLowerCase()));
    
    if (selectedPriority === 'ALL') return matchesSearch;
    return matchesSearch && t.Priority === Number(selectedPriority);
  });

  return (
    <div className="panel-card flex-1 flex flex-col min-h-0">
      <div className="panel-header flex justify-between items-center">
        <div>
          <h2 className="panel-title">Task Workflow Templates</h2>
          <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            Design operational checklists, target execution times, and safety roles.
          </p>
        </div>
        <button className="btn-primary" onClick={onCreate}>
          + Create Template
        </button>
      </div>

      {/* Filters HUD */}
      <div className="p-4 flex gap-4 items-center border-b" style={{ borderColor: 'var(--panel-border)', background: 'rgba(255,255,255,0.01)' }}>
        <div className="flex-1 relative">
          <input
            type="text"
            className="site-meta-pill w-full"
            style={{ borderRadius: '8px', padding: '8px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)' }}
            placeholder="Search templates by name or description..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 600 }}>Priority:</span>
          <select
            className="site-meta-pill"
            style={{ borderRadius: '8px', padding: '6px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
            value={selectedPriority}
            onChange={(e) => setSelectedPriority(e.target.value)}
          >
            <option value="ALL" style={{ background: 'var(--bg-main)' }}>All Priorities</option>
            <option value="1" style={{ background: 'var(--bg-main)' }}>Critical</option>
            <option value="2" style={{ background: 'var(--bg-main)' }}>High</option>
            <option value="3" style={{ background: 'var(--bg-main)' }}>Standard</option>
          </select>
        </div>
      </div>

      {/* Table Area */}
      <div className="panel-body-scrollable flex-1">
        {loading ? (
          <div className="flex justify-center items-center h-32 text-muted">Loading task templates...</div>
        ) : filteredTemplates.length === 0 ? (
          <div className="flex justify-center items-center h-32 text-muted">No task templates found.</div>
        ) : (
          <table className="a2ui-table">
            <thead>
              <tr style={{ borderBottom: '2px solid var(--panel-border)' }}>
                <th className="a2ui-label" style={{ padding: '12px' }}>Template Blueprint</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>Required Role</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>Priority</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>SLA Duration</th>
                <th className="a2ui-label" style={{ padding: '12px', textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredTemplates.map((t) => (
                <tr key={t.ID} className="hover:bg-white/5 transition-colors">
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px', maxWidth: '320px' }}>
                    <div>
                      <div className="font-semibold" style={{ color: 'var(--text-primary)' }}>{t.Name}</div>
                      <div className="text-xs text-muted line-clamp-2">{t.Description || 'No description provided.'}</div>
                    </div>
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px', color: 'var(--text-secondary)' }}>
                    <span className="site-meta-pill text-xs" style={{ background: 'rgba(255,255,255,0.03)' }}>
                      {getRoleName(t.TargetRoleID)}
                    </span>
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px' }}>
                    <span className={`task-priority-badge ${getPriorityBadgeClass(t.Priority)}`}>
                      {getPriorityLabel(t.Priority)}
                    </span>
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px', color: 'var(--text-secondary)' }}>
                    {t.EstimatedDurationMinutes || 15} mins
                  </td>
                  <td className="a2ui-value" style={{ padding: '12px' }}>
                    <div className="flex justify-end items-center gap-2">
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px' }} onClick={() => onEdit(t)}>
                        Edit
                      </button>
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px', borderColor: 'var(--priority-critical)', color: 'var(--priority-critical)' }} onClick={() => handleDelete(t.ID)}>
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
