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
import { useAppContext } from '../../contexts/AppContext';

interface AssetMainProps {
  onEdit: (asset: any) => void;
  onCreate: () => void;
  onError: (err: any) => void;
}

export const AssetMain: React.FC<AssetMainProps> = ({ onEdit, onCreate, onError }) => {
  const { activeSiteID } = useAppContext();
  const [assets, setAssets] = useState<any[]>([]);
  const [locations, setLocations] = useState<any[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedStatus, setSelectedStatus] = useState('ALL');
  const [loading, setLoading] = useState(true);

  const loadData = async () => {
    try {
      setLoading(true);
      const [fetchedAssets, fetchedLocations] = await Promise.all([
        ApiClient.fetchAssets(),
        ApiClient.fetchLocations()
      ]);
      setAssets(fetchedAssets || []);
      setLocations(fetchedLocations || []);
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
    if (!window.confirm('Are you sure you want to delete this asset? Tasks requiring it might be blocked.')) return;
    try {
      await ApiClient.deleteAsset(id);
      loadData();
    } catch (err) {
      onError(err);
    }
  };

  const filteredAssets = assets.filter((a) => {
    const matchesSearch =
      (a.Name && a.Name.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (a.AssetTag && a.AssetTag.toLowerCase().includes(searchTerm.toLowerCase()));
    
    // Filter by global active site context
    if (activeSiteID && activeSiteID !== 'ALL') {
      const loc = locations.find((l) => l.ID === a.LocationID);
      const locSiteID = loc ? (loc.SiteID || loc.site_id) : null;
      if (locSiteID !== activeSiteID) return false;
    }

    if (selectedStatus === 'ALL') return matchesSearch;
    return matchesSearch && a.Status === selectedStatus;
  });

  const getLocationName = (locId: string) => {
    const l = locations.find((x) => x.ID === locId);
    return l ? l.Name : locId;
  };

  return (
    <div className="panel-card flex-1 flex flex-col min-h-0">
      <div className="panel-header flex justify-between items-center">
        <div>
          <h2 className="panel-title">Physical Assets & Equipment</h2>
          <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            Track machinery, tools, IoT devices, or compliance materials assigned to specific facility shelves.
          </p>
        </div>
        <button className="btn-primary" onClick={onCreate}>
          + Register Asset
        </button>
      </div>

      {/* Filters HUD */}
      <div className="p-4 flex gap-4 items-center border-b" style={{ borderColor: 'var(--panel-border)', background: 'rgba(255,255,255,0.01)' }}>
        <div className="flex-1 relative">
          <input
            type="text"
            className="site-meta-pill w-full"
            style={{ borderRadius: '8px', padding: '8px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)' }}
            placeholder="Search assets by name or asset tag..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 600 }}>Status:</span>
          <select
            className="site-meta-pill"
            style={{ borderRadius: '8px', padding: '6px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
            value={selectedStatus}
            onChange={(e) => setSelectedStatus(e.target.value)}
          >
            <option value="ALL" style={{ background: 'var(--bg-main)' }}>All Statuses</option>
            <option value="AVAILABLE" style={{ background: 'var(--bg-main)' }}>Available</option>
            <option value="IN_USE" style={{ background: 'var(--bg-main)' }}>In Use</option>
            <option value="MAINTENANCE" style={{ background: 'var(--bg-main)' }}>Maintenance</option>
            <option value="BROKEN" style={{ background: 'var(--bg-main)' }}>Broken</option>
          </select>
        </div>
      </div>

      {/* Table Area */}
      <div className="panel-body-scrollable flex-1">
        {loading ? (
          <div className="flex justify-center items-center h-32 text-muted">Loading assets database...</div>
        ) : filteredAssets.length === 0 ? (
          <div className="flex justify-center items-center h-32 text-muted">No assets found matching criteria.</div>
        ) : (
          <table className="a2ui-table">
            <thead>
              <tr style={{ borderBottom: '2px solid var(--panel-border)' }}>
                <th className="a2ui-label" style={{ padding: '12px' }}>Asset Name</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>Asset Tag</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>Current Location</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>Status</th>
                <th className="a2ui-label" style={{ padding: '12px', textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredAssets.map((a) => (
                <tr key={a.ID} className="hover:bg-white/5 transition-colors">
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px' }}>
                    <span className="font-semibold" style={{ color: 'var(--text-primary)' }}>{a.Name}</span>
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px', fontFamily: 'var(--font-mono)', fontSize: '0.85rem', color: 'var(--accent-primary)' }}>
                    {a.AssetTag}
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px', color: 'var(--text-secondary)' }}>
                    {getLocationName(a.LocationID)}
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px' }}>
                    <span
                      className="site-meta-pill text-xs font-bold"
                      style={{
                        borderColor:
                          a.Status === 'AVAILABLE'
                            ? 'var(--priority-standard)'
                            : a.Status === 'MAINTENANCE' || a.Status === 'IN_USE'
                            ? 'var(--priority-high)'
                            : 'var(--priority-critical)',
                        color:
                          a.Status === 'AVAILABLE'
                            ? 'var(--priority-standard)'
                            : a.Status === 'MAINTENANCE' || a.Status === 'IN_USE'
                            ? 'var(--priority-high)'
                            : 'var(--priority-critical)',
                        background: 'rgba(255,255,255,0.02)'
                      }}
                    >
                      {a.Status}
                    </span>
                  </td>
                  <td className="a2ui-value" style={{ padding: '12px' }}>
                    <div className="flex justify-end items-center gap-2">
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px' }} onClick={() => onEdit(a)}>
                        Edit
                      </button>
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px', borderColor: 'var(--priority-critical)', color: 'var(--priority-critical)' }} onClick={() => handleDelete(a.ID)}>
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
