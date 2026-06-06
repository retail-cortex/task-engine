import React, { useState, useEffect } from 'react';
import { ApiClient } from '../../api/client';

interface RoleMainProps {
  onEdit: (role: any) => void;
  onCreate: () => void;
  onError: (err: any) => void;
}

export const RoleMain: React.FC<RoleMainProps> = ({ onEdit, onCreate, onError }) => {
  const [roles, setRoles] = useState<any[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [loading, setLoading] = useState(true);

  const loadData = async () => {
    try {
      setLoading(true);
      const fetchedRoles = await ApiClient.fetchRoles();
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
    if (!window.confirm('Are you sure you want to delete this role?')) return;
    try {
      await ApiClient.deleteRole(id);
      loadData();
    } catch (err) {
      onError(err);
    }
  };

  const filteredRoles = roles.filter((r) => {
    return (
      (r.Name && r.Name.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (r.Description && r.Description.toLowerCase().includes(searchTerm.toLowerCase()))
    );
  });

  return (
    <div className="panel-card flex-1 flex flex-col min-h-0">
      <div className="panel-header flex justify-between items-center">
        <div>
          <h2 className="panel-title">System Roles</h2>
          <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            Configure RBAC levels and policy groupings across storefront facilities.
          </p>
        </div>
        <button className="btn-primary" onClick={onCreate}>
          + Create Role
        </button>
      </div>

      {/* Filters HUD */}
      <div className="p-4 flex gap-4 items-center border-b" style={{ borderColor: 'var(--panel-border)', background: 'rgba(255,255,255,0.01)' }}>
        <div className="flex-1 relative">
          <input
            type="text"
            className="site-meta-pill w-full"
            style={{ borderRadius: '8px', padding: '8px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)' }}
            placeholder="Search roles by name or description..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      {/* Table Area */}
      <div className="panel-body-scrollable flex-1">
        {loading ? (
          <div className="flex justify-center items-center h-32 text-muted">Loading roles database...</div>
        ) : filteredRoles.length === 0 ? (
          <div className="flex justify-center items-center h-32 text-muted">No roles found.</div>
        ) : (
          <table className="a2ui-table">
            <thead>
              <tr style={{ borderBottom: '2px solid var(--panel-border)' }}>
                <th className="a2ui-label" style={{ padding: '12px' }}>Role Name</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>Description</th>
                <th className="a2ui-label" style={{ padding: '12px', textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredRoles.map((r) => (
                <tr key={r.ID} className="hover:bg-white/5 transition-colors">
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px' }}>
                    <span className="site-meta-pill text-xs font-bold uppercase" style={{ borderColor: 'var(--accent-primary)', color: 'var(--accent-primary)' }}>
                      {r.Name}
                    </span>
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px', color: 'var(--text-secondary)' }}>
                    {r.Description || 'No description provided.'}
                  </td>
                  <td className="a2ui-value" style={{ padding: '12px' }}>
                    <div className="flex justify-end items-center gap-2">
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px' }} onClick={() => onEdit(r)}>
                        Edit
                      </button>
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px', borderColor: 'var(--priority-critical)', color: 'var(--priority-critical)' }} onClick={() => handleDelete(r.ID)}>
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
