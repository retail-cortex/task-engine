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

interface AssetFormProps {
  asset?: any; // If provided, we are editing
  onSave: () => void;
  onCancel: () => void;
  onError: (err: any) => void;
}

export const AssetForm: React.FC<AssetFormProps> = ({ asset, onSave, onCancel, onError }) => {
  const [name, setName] = useState(asset?.Name || '');
  const [assetTag, setAssetTag] = useState(asset?.AssetTag || '');
  const [locationId, setLocationId] = useState(asset?.LocationID || '');
  const [status, setStatus] = useState(asset?.Status || 'AVAILABLE');
  const [metadataStr, setMetadataStr] = useState(asset?.Metadata ? JSON.stringify(asset.Metadata, null, 2) : '{}');

  const [locations, setLocations] = useState<any[]>([]);
  const [sites, setSites] = useState<any[]>([]);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    Promise.all([
      ApiClient.fetchLocations(),
      ApiClient.fetchSitesAdmin()
    ])
      .then(([fetchedLocations, fetchedSites]) => {
        setLocations(fetchedLocations || []);
        setSites(fetchedSites || []);
        if (!asset && fetchedLocations && fetchedLocations.length > 0) {
          setLocationId(fetchedLocations[0].ID);
        }
      })
      .catch(onError);
  }, [asset]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !assetTag || !locationId) {
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
      ...asset,
      Name: name,
      AssetTag: assetTag,
      LocationID: locationId,
      Status: status,
      Metadata: parsedMetadata,
    };

    try {
      setSubmitting(true);
      if (asset?.ID) {
        await ApiClient.updateAsset(asset.ID, payload);
      } else {
        // Resolve Org ID and Site ID from Selected Location
        const selectedLoc = locations.find((l) => l.ID === locationId);
        const siteId = selectedLoc?.SiteID || '';
        const selectedSite = sites.find((s) => s.ID === siteId);
        const orgId = selectedSite?.OrganizationID || '33333333-3333-3333-3333-333333333333';

        if (!siteId) {
          alert('Error: Could not resolve site context for the selected location.');
          return;
        }

        await ApiClient.createAsset(orgId, siteId, locationId, payload);
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
        <h2 className="panel-title">{asset ? 'Edit Equipment Details' : 'Register New Asset'}</h2>
        <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
          Assigned assets are dynamically matched against task checklist step requirements.
        </p>
      </div>

      <div className="panel-body-scrollable flex-1 p-6">
        <form onSubmit={handleSubmit} className="flex flex-col gap-6 max-w-3xl">
          {/* Asset Name */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Asset / Equipment Name *</label>
            <input
              type="text"
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              placeholder="e.g. Crown Power Pallet Jack C-5"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Asset Tag */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Asset Barcode / Tag *</label>
              <input
                type="text"
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                placeholder="e.g. EQ-PALLET-005"
                value={assetTag}
                onChange={(e) => setAssetTag(e.target.value)}
                required
              />
            </div>

            {/* Status */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Operational Status *</label>
              <select
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={status}
                onChange={(e) => setStatus(e.target.value)}
                required
              >
                <option value="AVAILABLE" style={{ background: 'var(--bg-main)' }}>Available (Ready for Tasks)</option>
                <option value="IN_USE" style={{ background: 'var(--bg-main)' }}>In Use (Actively claimed)</option>
                <option value="MAINTENANCE" style={{ background: 'var(--bg-main)' }}>Maintenance (Lockout Tagout)</option>
                <option value="BROKEN" style={{ background: 'var(--bg-main)' }}>Broken (Decommissioned)</option>
              </select>
            </div>
          </div>

          {/* Current Grounding Location */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Grounding Sub-Location (Storage Point) *</label>
            <select
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              value={locationId}
              onChange={(e) => setLocationId(e.target.value)}
              required
            >
              {locations.map((l) => (
                <option key={l.ID} value={l.ID} style={{ background: 'var(--bg-main)' }}>
                  {l.Name} ({l.LocationType})
                </option>
              ))}
            </select>
          </div>

          {/* Metadata */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Metadata (JSONB) *</label>
            <textarea
              className="site-meta-pill font-mono text-sm"
              style={{ borderRadius: '8px', padding: '12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)', minHeight: '100px', resize: 'vertical' }}
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
              {submitting ? 'Saving...' : 'Save Asset'}
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
