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

interface SiteFormProps {
  site?: any; // If provided, we are editing
  onSave: () => void;
  onCancel: () => void;
  onError: (err: any) => void;
}

export const SiteForm: React.FC<SiteFormProps> = ({ site, onSave, onCancel, onError }) => {
  const [name, setName] = useState(site?.Name || '');
  const [orgId, setOrgId] = useState(site?.OrganizationID || '');
  const [siteType, setSiteType] = useState(site?.SiteType || 'STORE');
  const [address, setAddress] = useState(site?.Address || '');
  const [latitude, setLatitude] = useState(site?.Latitude !== undefined ? site.Latitude : 0.0);
  const [longitude, setLongitude] = useState(site?.Longitude !== undefined ? site.Longitude : 0.0);
  const [altitude, setAltitude] = useState(site?.AltitudeMeters !== undefined ? site.AltitudeMeters : 0.0);
  const [icaoCode, setIcaoCode] = useState(site?.ICAOCode || '');
  const [timeZone, setTimeZone] = useState(site?.TimeZone || 'UTC');
  const [metadataStr, setMetadataStr] = useState(site?.Metadata ? JSON.stringify(site.Metadata, null, 2) : '{}');

  const [orgs, setOrgs] = useState<any[]>([]);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    ApiClient.fetchOrganizations()
      .then((data) => {
        setOrgs(data || []);
        if (!site && data && data.length > 0) {
          setOrgId(data[0].ID);
        }
      })
      .catch(onError);
  }, [site]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !orgId) {
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
      ...site,
      Name: name,
      OrganizationID: orgId,
      SiteType: siteType,
      Address: address,
      Latitude: Number(latitude),
      Longitude: Number(longitude),
      AltitudeMeters: Number(altitude),
      ICAOCode: icaoCode,
      TimeZone: timeZone,
      Metadata: parsedMetadata,
    };

    try {
      setSubmitting(true);
      if (site?.ID) {
        await ApiClient.updateSite(site.ID, payload);
      } else {
        await ApiClient.createSite(orgId, payload);
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
        <h2 className="panel-title">{site ? 'Edit Physical Site' : 'Register New Site'}</h2>
        <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
          Establish operational storefronts, regional yards, or logistics warehouses.
        </p>
      </div>

      <div className="panel-body-scrollable flex-1 p-6">
        <form onSubmit={handleSubmit} className="flex flex-col gap-6 max-w-3xl">
          {/* Site Name */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Site Name *</label>
            <input
              type="text"
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              placeholder="e.g. Dallas Store #1000"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Organization */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Organization Tenant *</label>
              <select
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={orgId}
                onChange={(e) => setOrgId(e.target.value)}
                required
                disabled={!!site} // Organization is immutable after creation
              >
                {orgs.map((o) => (
                  <option key={o.ID} value={o.ID} style={{ background: 'var(--bg-main)' }}>{o.Name}</option>
                ))}
              </select>
            </div>

            {/* Site Type */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Site Type *</label>
              <select
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={siteType}
                onChange={(e) => setSiteType(e.target.value)}
                required
              >
                <option value="STORE" style={{ background: 'var(--bg-main)' }}>Store (Retail)</option>
                <option value="WAREHOUSE" style={{ background: 'var(--bg-main)' }}>Warehouse (Logistics)</option>
                <option value="OFFICE" style={{ background: 'var(--bg-main)' }}>Office (Corporate)</option>
              </select>
            </div>
          </div>

          {/* Address */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Physical Address</label>
            <input
              type="text"
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              placeholder="123 Corporate Way, City, State"
              value={address}
              onChange={(e) => setAddress(e.target.value)}
            />
          </div>

          {/* Geographic Coordinates */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Latitude</label>
              <input
                type="number"
                step="any"
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={latitude}
                onChange={(e) => setLatitude(parseFloat(e.target.value) || 0.0)}
              />
            </div>

            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Longitude</label>
              <input
                type="number"
                step="any"
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={longitude}
                onChange={(e) => setLongitude(parseFloat(e.target.value) || 0.0)}
              />
            </div>

            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Elevation (Meters)</label>
              <input
                type="number"
                step="any"
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={altitude}
                onChange={(e) => setAltitude(parseFloat(e.target.value) || 0.0)}
              />
            </div>
          </div>

          {/* ICAO / Weather Station & TimeZone */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>ICAO Reporting Code (e.g. KDFW)</label>
              <input
                type="text"
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                placeholder="Weather reporting station identifier"
                value={icaoCode}
                onChange={(e) => setIcaoCode(e.target.value)}
              />
            </div>

            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>TimeZone Identifier (e.g. America/Chicago)</label>
              <input
                type="text"
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                placeholder="IANA Time Zone Name"
                value={timeZone}
                onChange={(e) => setTimeZone(e.target.value)}
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
              {submitting ? 'Saving...' : 'Save Site'}
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
