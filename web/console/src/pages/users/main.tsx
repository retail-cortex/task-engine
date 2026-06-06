import React, { useState, useEffect } from 'react';
import { ApiClient } from '../../api/client';
import { useAppContext } from '../../contexts/AppContext';

interface UserMainProps {
  onEdit: (user: any) => void;
  onCreate: () => void;
  onError: (err: any) => void;
}

export const UserMain: React.FC<UserMainProps> = ({ onEdit, onCreate, onError }) => {
  const { activeSiteID } = useAppContext();
  const [users, setUsers] = useState<any[]>([]);
  const [roles, setRoles] = useState<any[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedRole, setSelectedRole] = useState('ALL');
  const [loading, setLoading] = useState(true);

  const loadData = async () => {
    try {
      setLoading(true);
      const [fetchedUsers, fetchedRoles] = await Promise.all([
        ApiClient.fetchUsers(),
        ApiClient.fetchRoles()
      ]);
      setUsers(fetchedUsers || []);
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
    if (!window.confirm('Are you sure you want to delete this user?')) return;
    try {
      await ApiClient.deleteUser(id);
      loadData();
    } catch (err) {
      onError(err);
    }
  };

  const handleRoleChange = async (userId: string, roleId: string) => {
    try {
      await ApiClient.assignUserRole(userId, roleId);
      loadData();
    } catch (err) {
      onError(err);
    }
  };

  const filteredUsers = users.filter((u) => {
    const matchesSearch =
      (u.Name && u.Name.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (u.Email && u.Email.toLowerCase().includes(searchTerm.toLowerCase()));
    
    // Filter by global active site context
    if (activeSiteID && activeSiteID !== 'ALL') {
      const belongsToSite = u.Sites && u.Sites.some((s: any) => s.ID === activeSiteID);
      if (!belongsToSite) return false;
    }

    if (selectedRole === 'ALL') return matchesSearch;
    
    const hasRole = u.Roles && u.Roles.some((r: any) => r.ID === selectedRole || r.Name === selectedRole);
    return matchesSearch && hasRole;
  });

  return (
    <div className="panel-card flex-1 flex flex-col min-h-0">
      <div className="panel-header flex justify-between items-center">
        <div>
          <h2 className="panel-title">System Users</h2>
          <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            Manage personnel directory, security credentials, and access scopes.
          </p>
        </div>
        <button className="btn-primary" onClick={onCreate}>
          + Register User
        </button>
      </div>

      {/* Filters HUD */}
      <div className="p-4 flex gap-4 items-center border-b" style={{ borderColor: 'var(--panel-border)', background: 'rgba(255,255,255,0.01)' }}>
        <div className="flex-1 relative">
          <input
            type="text"
            className="site-meta-pill w-full"
            style={{ borderRadius: '8px', padding: '8px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)' }}
            placeholder="Search by name or email..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 600 }}>Role:</span>
          <select
            className="site-meta-pill"
            style={{ borderRadius: '8px', padding: '6px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
            value={selectedRole}
            onChange={(e) => setSelectedRole(e.target.value)}
          >
            <option value="ALL" style={{ background: 'var(--bg-main)' }}>All Roles</option>
            {roles.map((r) => (
              <option key={r.ID} value={r.ID} style={{ background: 'var(--bg-main)' }}>{r.Name}</option>
            ))}
          </select>
        </div>
      </div>

      {/* Table Area */}
      <div className="panel-body-scrollable flex-1">
        {loading ? (
          <div className="flex justify-center items-center h-32 text-muted">Loading user database...</div>
        ) : filteredUsers.length === 0 ? (
          <div className="flex justify-center items-center h-32 text-muted">No users found matching search criteria.</div>
        ) : (
          <table className="a2ui-table">
            <thead>
              <tr style={{ borderBottom: '2px solid var(--panel-border)' }}>
                <th className="a2ui-label" style={{ padding: '12px' }}>Name / Identity</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>OAuth Source</th>
                <th className="a2ui-label" style={{ padding: '12px' }}>Assigned Roles</th>
                <th className="a2ui-label" style={{ padding: '12px', textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredUsers.map((u) => (
                <tr key={u.ID} className="hover:bg-white/5 transition-colors">
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px' }}>
                    <div>
                      <div className="font-semibold" style={{ color: 'var(--text-primary)' }}>{u.Name || 'Unnamed Profile'}</div>
                      <div className="text-xs" style={{ color: 'var(--text-muted)' }}>{u.Email}</div>
                    </div>
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px', color: 'var(--text-secondary)' }}>
                    <span className="site-meta-pill text-xs uppercase" style={{ background: 'rgba(255,255,255,0.03)' }}>
                      {u.OAuthProvider || 'mock'}:{u.OAuthID?.substring(0, 8) || 'bypass'}
                    </span>
                  </td>
                  <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px' }}>
                    <div className="flex flex-wrap gap-1">
                      {u.Roles && u.Roles.length > 0 ? (
                        u.Roles.map((r: any) => (
                          <span key={r.ID} className="site-meta-pill text-xs font-bold" style={{ borderColor: 'var(--accent-primary)', color: 'var(--accent-primary)' }}>
                            {r.Name}
                          </span>
                        ))
                      ) : (
                        <span className="text-xs text-muted">No Roles Assigned</span>
                      )}
                    </div>
                  </td>
                  <td className="a2ui-value" style={{ padding: '12px' }}>
                    <div className="flex justify-end items-center gap-2">
                      {/* Quick Role Assignment Dropdown */}
                      <select
                        className="site-meta-pill text-xs"
                        style={{ padding: '2px 6px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)' }}
                        value={u.Roles && u.Roles.length > 0 ? u.Roles[0].ID : ''}
                        onChange={(e) => handleRoleChange(u.ID, e.target.value)}
                      >
                        <option value="" disabled>+ Assign Role</option>
                        {roles.map((r) => (
                          <option key={r.ID} value={r.ID} style={{ background: 'var(--bg-main)' }}>{r.Name}</option>
                        ))}
                      </select>
                      
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px' }} onClick={() => onEdit(u)}>
                        Edit
                      </button>
                      <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px', borderColor: 'var(--priority-critical)', color: 'var(--priority-critical)' }} onClick={() => handleDelete(u.ID)}>
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
