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

interface LocationFormProps {
  loc?: any; // If provided, we are editing
  onSave: () => void;
  onCancel: () => void;
  onError: (err: any) => void;
}

export const LocationForm: React.FC<LocationFormProps> = ({ loc, onSave, onCancel, onError }) => {
  const [name, setName] = useState(loc?.Name || '');
  const [siteId, setSiteId] = useState(loc?.SiteID || '');
  const [parentId, setParentId] = useState(loc?.ParentID || '');
  const [locationType, setLocationType] = useState(loc?.LocationType || 'FIXTURE');
  const [locationFunctionType, setLocationFunctionType] = useState(loc?.LocationFunctionType || 'DISPLAY');
  const [x, setX] = useState(loc?.X !== undefined ? loc.X : 0.0);
  const [y, setY] = useState(loc?.Y !== undefined ? loc.Y : 0.0);
  const [z, setZ] = useState(loc?.Z !== undefined ? loc.Z : 0.0);
  const [metadataStr, setMetadataStr] = useState(loc?.Metadata ? JSON.stringify(loc.Metadata, null, 2) : '{}');

  const [sites, setSites] = useState<any[]>([]);
  const [siblingLocations, setSiblingLocations] = useState<any[]>([]);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    ApiClient.fetchSitesAdmin()
      .then((data) => {
        setSites(data || []);
        if (!loc && data && data.length > 0) {
          setSiteId(data[0].ID);
        }
      })
      .catch(onError);
  }, [loc]);

  // Load other locations under the same site to populate potential parent locations
  useEffect(() => {
    if (!siteId) return;
    ApiClient.fetchLocations()
      .then((data) => {
        const filtered = data.filter((x) => x.SiteID === siteId && x.ID !== loc?.ID);
        setSiblingLocations(filtered);
      })
      .catch(onError);
  }, [siteId, loc]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !siteId) {
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
      ...loc,
      Name: name,
      SiteID: siteId,
      ParentID: parentId || null,
      LocationType: locationType,
      LocationFunctionType: locationFunctionType,
      X: Number(x),
      Y: Number(y),
      Z: Number(z),
      Metadata: parsedMetadata,
    };

    try {
      setSubmitting(true);
      
      // Resolve Org ID of the selected Site for the creation endpoint
      const selectedSiteObj = sites.find((s) => s.ID === siteId);
      const orgId = selectedSiteObj?.OrganizationID || '33333333-3333-3333-3333-333333333333';

      if (loc?.ID) {
        await ApiClient.updateLocation(loc.ID, payload);
      } else {
        await ApiClient.createLocation(orgId, siteId, payload);
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
        <h2 className="panel-title">{loc ? 'Edit Sub-Location Maps' : 'Map New Facility Location'}</h2>
        <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
          Gridding coordinates enables precise spatial AI routing and proximity checks.
        </p>
      </div>

      <div className="panel-body-scrollable flex-1 p-6">
        <form onSubmit={handleSubmit} className="flex flex-col gap-6 max-w-3xl">
          {/* Location Name */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Location Identifier Name *</label>
            <input
              type="text"
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              placeholder="e.g. Aisle 4 Shelf B"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Site Context */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Site Context *</label>
              <select
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={siteId}
                onChange={(e) => setSiteId(e.target.value)}
                required
                disabled={!!loc} // Site is immutable after mapping
              >
                {sites.map((s) => (
                  <option key={s.ID} value={s.ID} style={{ background: 'var(--bg-main)' }}>{s.Name}</option>
                ))}
              </select>
            </div>

            {/* Parent Location */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Parent Location (Nesting Context)</label>
              <select
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={parentId}
                onChange={(e) => setParentId(e.target.value)}
              >
                <option value="" style={{ background: 'var(--bg-main)' }}>-- None (Top Level Shelf/Aisle) --</option>
                {siblingLocations.map((x) => (
                  <option key={x.ID} value={x.ID} style={{ background: 'var(--bg-main)' }}>{x.Name} ({x.LocationType})</option>
                ))}
              </select>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Location Type */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Location Type *</label>
              <select
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={locationType}
                onChange={(e) => setLocationType(e.target.value)}
                required
              >
                <option value="FIXTURE" style={{ background: 'var(--bg-main)' }}>Fixture</option>
                <option value="SHELF" style={{ background: 'var(--bg-main)' }}>Shelf</option>
                <option value="AISLE" style={{ background: 'var(--bg-main)' }}>Aisle</option>
                <option value="REGISTER" style={{ background: 'var(--bg-main)' }}>Register (POS)</option>
                <option value="BACKROOM" style={{ background: 'var(--bg-main)' }}>Backroom / Storage</option>
              </select>
            </div>

            {/* Function Type */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Location Functional Area *</label>
              <select
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={locationFunctionType}
                onChange={(e) => setLocationFunctionType(e.target.value)}
                required
              >
                <option value="DISPLAY" style={{ background: 'var(--bg-main)' }}>Display (Salesfloor)</option>
                <option value="STOCK_POINT" style={{ background: 'var(--bg-main)' }}>Stock Point (Inventory)</option>
                <option value="RECEIVING" style={{ background: 'var(--bg-main)' }}>Receiving (Dock)</option>
                <option value="TRANSIT" style={{ background: 'var(--bg-main)' }}>Transit Area</option>
              </select>
            </div>
          </div>

          {/* Spatial Coordinates offset relative to origin */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>X Offset (Horizontal Plan)</label>
              <input
                type="number"
                step="any"
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={x}
                onChange={(e) => setX(parseFloat(e.target.value) || 0.0)}
              />
            </div>

            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Y Offset (Vertical Plan)</label>
              <input
                type="number"
                step="any"
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={y}
                onChange={(e) => setY(parseFloat(e.target.value) || 0.0)}
              />
            </div>

            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Z Offset (Elevation / Floor Height)</label>
              <input
                type="number"
                step="any"
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={z}
                onChange={(e) => setZ(parseFloat(e.target.value) || 0.0)}
              />
            </div>
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
              {submitting ? 'Saving...' : 'Save Location'}
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
