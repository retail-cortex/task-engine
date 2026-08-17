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

interface UserFormProps {
  user?: any; // If provided, we are editing
  onSave: () => void;
  onCancel: () => void;
  onError: (err: any) => void;
}

export const UserForm: React.FC<UserFormProps> = ({ user, onSave, onCancel, onError }) => {
  const [name, setName] = useState(user?.Name || '');
  const [email, setEmail] = useState(user?.Email || '');
  const [oauthProvider, setOauthProvider] = useState(user?.OAuthProvider || 'google');
  const [oauthId, setOauthId] = useState(user?.OAuthID || '');
  const [metadataStr, setMetadataStr] = useState(user?.Metadata ? JSON.stringify(user.Metadata, null, 2) : '{}');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !email || !oauthProvider || !oauthId) {
      alert('Please fill out all required fields.');
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
      ...user,
      Name: name,
      Email: email,
      OAuthProvider: oauthProvider,
      OAuthID: oauthId,
      Metadata: parsedMetadata,
    };

    try {
      setSubmitting(true);
      if (user?.ID) {
        await ApiClient.updateUser(user.ID, payload);
      } else {
        await ApiClient.createUser(payload);
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
        <h2 className="panel-title">{user ? 'Edit User Profile' : 'Register New User'}</h2>
        <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
          Ensure compliance parameters match the corporate Single Sign-On (SSO) directory.
        </p>
      </div>

      <div className="panel-body-scrollable flex-1 p-6">
        <form onSubmit={handleSubmit} className="flex flex-col gap-6 max-w-2xl">
          {/* Name */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Full Name *</label>
            <input
              type="text"
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              placeholder="e.g. John Doe"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          {/* Email */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Email Address *</label>
            <input
              type="email"
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              placeholder="e.g. jdoe@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>

          {/* OAuth Provider & ID */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>OAuth Provider *</label>
              <select
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={oauthProvider}
                onChange={(e) => setOauthProvider(e.target.value)}
                required
              >
                <option value="google" style={{ background: 'var(--bg-main)' }}>Google</option>
                <option value="github" style={{ background: 'var(--bg-main)' }}>GitHub</option>
                <option value="mock" style={{ background: 'var(--bg-main)' }}>Mock Developer Bypass</option>
              </select>
            </div>

            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>OAuth Subject ID *</label>
              <input
                type="text"
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                placeholder="SSO Subject Identifier"
                value={oauthId}
                onChange={(e) => setOauthId(e.target.value)}
                required
              />
            </div>
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
              {submitting ? 'Saving...' : 'Save User'}
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
