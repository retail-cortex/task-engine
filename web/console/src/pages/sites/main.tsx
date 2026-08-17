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

interface SiteMainProps {
  onEdit: (site: any) => void;
  onCreate: () => void;
  onError: (err: any) => void;
}

export const SiteMain: React.FC<SiteMainProps> = ({ onEdit, onCreate, onError }) => {
  const [sites, setSites] = useState<any[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedType, setSelectedType] = useState('ALL');
  const [loading, setLoading] = useState(true);

  const loadData = async () => {
    try {
      setLoading(true);
      const fetchedSites = await ApiClient.fetchSitesAdmin();
      setSites(fetchedSites || []);
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
    if (!window.confirm('Are you sure you want to delete this site? All mapped sub-locations and assets will be deleted.')) return;
    try {
      await ApiClient.deleteSite(id);
      loadData();
    } catch (err) {
      onError(err);
    }
  };

  const filteredSites = sites.filter((s) => {
    const matchesSearch = s.Name && s.Name.toLowerCase().includes(searchTerm.toLowerCase());
    if (selectedType === 'ALL') return matchesSearch;
    return matchesSearch && s.SiteType === selectedType;
  });

  return (
    <div className="panel-card flex-1 flex flex-col min-h-0">
      <div className="panel-header flex justify-between items-center">
        <div>
          <h2 className="panel-title">Physical Sites</h2>
          <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            Manage stores, warehouses, regional headquarters, and geo-mapping boundaries.
          </p>
        </div>
        <button className="btn-primary" onClick={onCreate}>
          + Register Site
        </button>
      </div>

      {/* Filters HUD */}
      <div className="p-4 flex gap-4 items-center border-b" style={{ borderColor: 'var(--panel-border)', background: 'rgba(255,255,255,0.01)' }}>
        <div className="flex-1 relative">
          <input
            type="text"
            className="site-meta-pill w-full"
            style={{ borderRadius: '8px', padding: '8px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)' }}
            placeholder="Search sites by name..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 600 }}>Type:</span>
          <select
            className="site-meta-pill"
            style={{ borderRadius: '8px', padding: '6px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
            value={selectedType}
            onChange={(e) => setSelectedType(e.target.value)}
          >
            <option value="ALL" style={{ background: 'var(--bg-main)' }}>All Types</option>
            <option value="STORE" style={{ background: 'var(--bg-main)' }}>Store</option>
            <option value="WAREHOUSE" style={{ background: 'var(--bg-main)' }}>Warehouse</option>
            <option value="OFFICE" style={{ background: 'var(--bg-main)' }}>Office</option>
          </select>
        </div>
      </div>

      {/* Table Area */}
      <div className="panel-body-scrollable flex-1">
        {loading ? (
          <div className="flex justify-center items-center h-32 text-muted">Loading sites database...</div>
        ) : filteredSites.length === 0 ? (
          <div className="flex justify-center items-center h-32 text-muted">No sites found.</div>
        ) : (
          <table className="a2ui-table">
            <thead>
              <tr style={{ borderBottom: '2px solid var(--panel-border)' }}>
                <th className="a2ui-label" style={{ padding: '12px' }}>Site Name</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>Type</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>Coordinates / Address</th>
                <th className="a2ui-label" style={{ padding: '12px', textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredSites.map((s) => (
                <tr key={s.ID} className="hover:bg-white/5 transition-colors">
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px' }}>
                    <div>
                      <div className="font-semibold" style={{ color: 'var(--text-primary)' }}>{s.Name}</div>
                      <div className="text-xs" style={{ color: 'var(--text-muted)' }}>TZ: {s.TimeZone || 'UTC'} | ICAO: {s.ICAOCode || 'N/A'}</div>
                    </div>
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px' }}>
                    <span className="site-meta-pill text-xs font-bold uppercase" style={{ background: 'rgba(255,255,255,0.03)', color: s.SiteType === 'STORE' ? 'var(--priority-standard)' : 'var(--accent-primary)' }}>
                      {s.SiteType}
                    </span>
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px', color: 'var(--text-secondary)' }}>
                    <div>
                      <div className="text-xs">{s.Address || 'No Address Provided'}</div>
                      <div className="text-xs" style={{ color: 'var(--text-muted)' }}>Lat: {s.Latitude?.toFixed(4)}, Lng: {s.Longitude?.toFixed(4)}</div>
                    </div>
                  </td>
                  <td className="a2ui-value" style={{ padding: '12px' }}>
                    <div className="flex justify-end items-center gap-2">
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px' }} onClick={() => onEdit(s)}>
                        Edit
                      </button>
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px', borderColor: 'var(--priority-critical)', color: 'var(--priority-critical)' }} onClick={() => handleDelete(s.ID)}>
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
