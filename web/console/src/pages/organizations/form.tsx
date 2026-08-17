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

interface OrgFormProps {
  org?: any; // If provided, we are editing
  onSave: () => void;
  onCancel: () => void;
  onError: (err: any) => void;
}

export const OrgForm: React.FC<OrgFormProps> = ({ org, onSave, onCancel, onError }) => {
  const [name, setName] = useState(org?.Name || '');
  const [parentId, setParentId] = useState(org?.ParentID || '');
  const [metadataStr, setMetadataStr] = useState(org?.Metadata ? JSON.stringify(org.Metadata, null, 2) : '{}');
  const [allOrgs, setAllOrgs] = useState<any[]>([]);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    ApiClient.fetchOrganizations()
      .then((data) => {
        // Exclude the current org from being its own parent to prevent circular loops
        const filtered = data.filter((o) => o.ID !== org?.ID);
        setAllOrgs(filtered);
      })
      .catch(onError);
  }, [org]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name) {
      alert('Please enter an organization name.');
      return;
    }

    let parsedMetadata = {};
    try {
      parsedMetadata = JSON.parse(metadataStr);
    } catch (err) {
      alert('Invalid JSON in Metadata field.');
      return;
    }

    const payload = {
      ...org,
      Name: name,
      ParentID: parentId || null,
      Metadata: parsedMetadata,
    };

    try {
      setSubmitting(true);
      if (org?.ID) {
        await ApiClient.updateOrganization(org.ID, payload);
      } else {
        await ApiClient.createOrganization(payload);
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
        <h2 className="panel-title">{org ? 'Edit Corporate Organization' : 'Create Organization'}</h2>
        <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
          Establish corporate tenancy structures and metadata fields.
        </p>
      </div>

      <div className="panel-body-scrollable flex-1 p-6">
        <form onSubmit={handleSubmit} className="flex flex-col gap-6 max-w-2xl">
          {/* Org Name */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Organization Name *</label>
            <input
              type="text"
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              placeholder="e.g. Altostrat Retail West"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          {/* Parent Organization */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Parent Organization (Optional)</label>
            <select
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              value={parentId}
              onChange={(e) => setParentId(e.target.value)}
            >
              <option value="" style={{ background: 'var(--bg-main)' }}>-- None (Top Level) --</option>
              {allOrgs.map((o) => (
                <option key={o.ID} value={o.ID} style={{ background: 'var(--bg-main)' }}>{o.Name}</option>
              ))}
            </select>
          </div>

          {/* Metadata */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Metadata (JSONB) *</label>
            <textarea
              className="site-meta-pill font-mono text-sm"
              style={{ borderRadius: '8px', padding: '12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)', minHeight: '120px', resize: 'vertical' }}
              placeholder="{}"
              value={metadataStr}
              onChange={(e) => setMetadataStr(e.target.value)}
              required
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
              {submitting ? 'Saving...' : 'Save Organization'}
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
