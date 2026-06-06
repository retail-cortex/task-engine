import React, { useState, useEffect } from 'react';
import { ApiClient } from '../../api/client';

interface OrgMainProps {
  onEdit: (org: any) => void;
  onCreate: () => void;
  onError: (err: any) => void;
}

export const OrgMain: React.FC<OrgMainProps> = ({ onEdit, onCreate, onError }) => {
  const [orgs, setOrgs] = useState<any[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [loading, setLoading] = useState(true);

  const loadData = async () => {
    try {
      setLoading(true);
      const fetchedOrgs = await ApiClient.fetchOrganizations();
      setOrgs(fetchedOrgs || []);
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
    if (!window.confirm('Are you sure you want to delete this organization? Doing so will delete all sites associated with it.')) return;
    try {
      await ApiClient.deleteOrganization(id);
      loadData();
    } catch (err) {
      onError(err);
    }
  };

  const filteredOrgs = orgs.filter((o) => {
    return o.Name && o.Name.toLowerCase().includes(searchTerm.toLowerCase());
  });

  return (
    <div className="panel-card flex-1 flex flex-col min-h-0">
      <div className="panel-header flex justify-between items-center">
        <div>
          <h2 className="panel-title">Corporate Organizations</h2>
          <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            Configure corporate tenants, parent-child hierarchies, and regional structures.
          </p>
        </div>
        <button className="btn-primary" onClick={onCreate}>
          + Create Organization
        </button>
      </div>

      {/* Filters HUD */}
      <div className="p-4 flex gap-4 items-center border-b" style={{ borderColor: 'var(--panel-border)', background: 'rgba(255,255,255,0.01)' }}>
        <div className="flex-1 relative">
          <input
            type="text"
            className="site-meta-pill w-full"
            style={{ borderRadius: '8px', padding: '8px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)' }}
            placeholder="Search organizations by name..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      {/* Table Area */}
      <div className="panel-body-scrollable flex-1">
        {loading ? (
          <div className="flex justify-center items-center h-32 text-muted">Loading organizations...</div>
        ) : filteredOrgs.length === 0 ? (
          <div className="flex justify-center items-center h-32 text-muted">No organizations found.</div>
        ) : (
          <table className="a2ui-table">
            <thead>
              <tr style={{ borderBottom: '2px solid var(--panel-border)' }}>
                <th className="a2ui-label" style={{ padding: '12px' }}>Organization Name</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>UUID</th>
                <th className="a2ui-label" style={{ padding: '12px', textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredOrgs.map((o) => (
                <tr key={o.ID} className="hover:bg-white/5 transition-colors">
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px' }}>
                    <div className="font-semibold" style={{ color: 'var(--text-primary)' }}>{o.Name}</div>
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px', fontFamily: 'var(--font-mono)', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                    {o.ID}
                  </td>
                  <td className="a2ui-value" style={{ padding: '12px' }}>
                    <div className="flex justify-end items-center gap-2">
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px' }} onClick={() => onEdit(o)}>
                        Edit
                      </button>
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px', borderColor: 'var(--priority-critical)', color: 'var(--priority-critical)' }} onClick={() => handleDelete(o.ID)}>
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
