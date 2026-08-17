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
import { ApiClient } from '../../api/client';

interface RoleFormProps {
  role?: any; // If provided, we are editing
  onSave: () => void;
  onCancel: () => void;
  onError: (err: any) => void;
}

export const RoleForm: React.FC<RoleFormProps> = ({ role, onSave, onCancel, onError }) => {
  const [name, setName] = useState(role?.Name || '');
  const [description, setDescription] = useState(role?.Description || '');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name) {
      alert('Please enter a role name.');
      return;
    }

    const payload = {
      ...role,
      Name: name,
      Description: description,
    };

    try {
      setSubmitting(true);
      if (role?.ID) {
        await ApiClient.updateRole(role.ID, payload);
      } else {
        await ApiClient.createRole(payload);
      }
      onSave();
    } catch (err) {
      onError(err);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="panel-card flex-1 flex flex-col min-h-0">
      <div className="panel-header">
        <h2 className="panel-title">{role ? 'Edit Role Details' : 'Create System Role'}</h2>
        <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
          Create standard security groups mapped to personnel tasks and approvals.
        </p>
      </div>

      <div className="panel-body-scrollable flex-1 p-6">
        <form onSubmit={handleSubmit} className="flex flex-col gap-6 max-w-2xl">
          {/* Role Name */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Role Name *</label>
            <input
              type="text"
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              placeholder="e.g. SHIFT_LEAD"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          {/* Description */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Description</label>
            <textarea
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)', minHeight: '100px', resize: 'vertical' }}
              placeholder="Provide context on what capabilities this role holds..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          {/* Actions */}
          <div className="flex gap-4 mt-4">
            <button
              type="submit"
              className="btn-primary"
              disabled={submitting}
              style={{ padding: '10px 24px' }}
            >
              {submitting ? 'Saving...' : 'Save Role'}
            </button>
            <button
              type="button"
              className="a2ui-btn-action"
              onClick={onCancel}
              style={{ padding: '10px 24px' }}
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
