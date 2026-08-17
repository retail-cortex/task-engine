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
import { useAppContext } from '../contexts/AppContext';

// Import all entity page controllers
import { UserMain } from '../pages/users/main';
import { UserForm } from '../pages/users/form';
import { RoleMain } from '../pages/roles/main';
import { RoleForm } from '../pages/roles/form';
import { OrgMain } from '../pages/organizations/main';
import { OrgForm } from '../pages/organizations/form';
import { SiteMain } from '../pages/sites/main';
import { SiteForm } from '../pages/sites/form';
import { LocationMain } from '../pages/locations/main';
import { LocationForm } from '../pages/locations/form';
import { AssetMain } from '../pages/assets/main';
import { AssetForm } from '../pages/assets/form';
import { TaskMain } from '../pages/tasks/main';
import { TaskForm } from '../pages/tasks/form';

interface AdminPanelProps {
  onExit: () => void;
}

type EntityType = 'users' | 'roles' | 'organizations' | 'sites' | 'locations' | 'assets' | 'tasks';
type ViewModeType = 'list' | 'form';

const AdminPanel: React.FC<AdminPanelProps> = ({ onExit }) => {
  const { userRole, handleApiError } = useAppContext();
  const [activeEntity, setActiveEntity] = useState<EntityType>('users');
  const [viewMode, setViewMode] = useState<ViewModeType>('list');
  const [editingItem, setEditingItem] = useState<any>(null);

  // 1. FRONTEND SECURITY SHIELD
  if (userRole !== 'ADMIN') {
    return (
      <main className="flex justify-center items-center h-full w-full p-8" style={{ background: 'var(--bg-main)' }}>
        <div className="panel-card max-w-lg p-8 flex flex-col items-center text-center gap-6" style={{ borderColor: 'var(--priority-critical)' }}>
          <div className="pulse-indicator" style={{ background: 'var(--priority-critical)', boxShadow: '0 0 15px var(--priority-critical)' }}></div>
          <h2 className="brand-title" style={{ background: 'linear-gradient(135deg, #ffffff 0%, var(--priority-critical) 100%)', webkitTextFillColor: 'transparent', webkitBackgroundClip: 'text' }}>
            SECURITY PROTOCOL TRIGGERED
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.95rem', lineHeight: 1.5 }}>
            Access to the administrative control systems is strictly reserved for users authenticated with the <strong>ADMIN</strong> role. Your credentials have been logged.
          </p>
          <button className="btn-primary" onClick={onExit} style={{ background: 'var(--priority-critical-dark)' }}>
            Exit Secure Terminal
          </button>
        </div>
      </main>
    );
  }

  const navigateToList = () => {
    setViewMode('list');
    setEditingItem(null);
  };

  const navigateToCreate = () => {
    setEditingItem(null);
    setViewMode('form');
  };

  const navigateToEdit = (item: any) => {
    setEditingItem(item);
    setViewMode('form');
  };

  const renderActiveView = () => {
    switch (activeEntity) {
      case 'users':
        return viewMode === 'list' ? (
          <UserMain onEdit={navigateToEdit} onCreate={navigateToCreate} onError={handleApiError} />
        ) : (
          <UserForm user={editingItem} onSave={navigateToList} onCancel={navigateToList} onError={handleApiError} />
        );
      case 'roles':
        return viewMode === 'list' ? (
          <RoleMain onEdit={navigateToEdit} onCreate={navigateToCreate} onError={handleApiError} />
        ) : (
          <RoleForm role={editingItem} onSave={navigateToList} onCancel={navigateToList} onError={handleApiError} />
        );
      case 'organizations':
        return viewMode === 'list' ? (
          <OrgMain onEdit={navigateToEdit} onCreate={navigateToCreate} onError={handleApiError} />
        ) : (
          <OrgForm org={editingItem} onSave={navigateToList} onCancel={navigateToList} onError={handleApiError} />
        );
      case 'sites':
        return viewMode === 'list' ? (
          <SiteMain onEdit={navigateToEdit} onCreate={navigateToCreate} onError={handleApiError} />
        ) : (
          <SiteForm site={editingItem} onSave={navigateToList} onCancel={navigateToList} onError={handleApiError} />
        );
      case 'locations':
        return viewMode === 'list' ? (
          <LocationMain onEdit={navigateToEdit} onCreate={navigateToCreate} onError={handleApiError} />
        ) : (
          <LocationForm loc={editingItem} onSave={navigateToList} onCancel={navigateToList} onError={handleApiError} />
        );
      case 'assets':
        return viewMode === 'list' ? (
          <AssetMain onEdit={navigateToEdit} onCreate={navigateToCreate} onError={handleApiError} />
        ) : (
          <AssetForm asset={editingItem} onSave={navigateToList} onCancel={navigateToList} onError={handleApiError} />
        );
      case 'tasks':
        return viewMode === 'list' ? (
          <TaskMain onEdit={navigateToEdit} onCreate={navigateToCreate} onError={handleApiError} />
        ) : (
          <TaskForm template={editingItem} onSave={navigateToList} onCancel={navigateToList} onError={handleApiError} />
        );
      default:
        return <div className="text-muted">Entity view not implemented.</div>;
    }
  };

  const sidebarItems: { id: EntityType; label: string; icon: React.ReactNode }[] = [
    {
      id: 'users',
      label: 'Personnel',
      icon: (
        <svg className="admin-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
          <circle cx="9" cy="7" r="4" />
          <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
          <path d="M16 3.13a4 4 0 0 1 0 7.75" />
        </svg>
      )
    },
    {
      id: 'locations',
      label: 'Sub-Locations',
      icon: (
        <svg className="admin-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" />
          <line x1="22" y1="12" x2="18" y2="12" />
          <line x1="6" y1="12" x2="2" y2="12" />
          <line x1="12" y1="6" x2="12" y2="2" />
          <line x1="12" y1="22" x2="12" y2="18" />
        </svg>
      )
    },
    {
      id: 'assets',
      label: 'Physical Assets',
      icon: (
        <svg className="admin-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
          <polyline points="3.27 6.96 12 12.01 20.73 6.96" />
          <line x1="12" y1="22.08" x2="12" y2="12" />
        </svg>
      )
    },
    {
      id: 'roles',
      label: 'Security Roles',
      icon: (
        <svg className="admin-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
          <path d="M7 11V7a5 5 0 0 1 10 0v4" />
        </svg>
      )
    },
    {
      id: 'organizations',
      label: 'Organizations',
      icon: (
        <svg className="admin-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
        </svg>
      )
    },
    {
      id: 'sites',
      label: 'Physical Sites',
      icon: (
        <svg className="admin-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
          <polyline points="9 22 9 12 15 12 15 22" />
        </svg>
      )
    },
    {
      id: 'tasks',
      label: 'Task Blueprints',
      icon: (
        <svg className="admin-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
          <polyline points="14 2 14 8 20 8" />
          <line x1="16" y1="13" x2="8" y2="13" />
          <line x1="16" y1="17" x2="8" y2="17" />
          <polyline points="10 9 9 9 8 9" />
        </svg>
      )
    }
  ];

  return (
    <div className="admin-container">
      {/* Sidebar Navigation */}
      <aside className="panel-card admin-sidebar">
        <div style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0 8px' }}>
            <h3 className="brand-title text-sm font-bold tracking-wider" style={{ color: 'var(--text-secondary)' }}>
              MDM CONTROL PANEL
            </h3>
          </div>

          <nav className="admin-nav" style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
            {/* Top Scoped Admin Functions */}
            {sidebarItems.slice(0, 3).map((item) => (
              <button
                key={item.id}
                onClick={() => {
                  setActiveEntity(item.id);
                  navigateToList();
                }}
                className={`admin-nav-item ${activeEntity === item.id ? 'active' : ''}`}
                style={{
                  textAlign: 'left'
                }}
              >
                {item.icon}
                <span className="text-sm">{item.label}</span>
              </button>
            ))}

            {/* Line Separator */}
            <hr style={{ margin: '12px 8px', borderColor: 'var(--panel-border)', opacity: 0.4 }} />

            {/* General System Admin Functions */}
            {sidebarItems.slice(3).map((item) => (
              <button
                key={item.id}
                onClick={() => {
                  setActiveEntity(item.id);
                  navigateToList();
                }}
                className={`admin-nav-item ${activeEntity === item.id ? 'active' : ''}`}
                style={{
                  textAlign: 'left'
                }}
              >
                {item.icon}
                <span className="text-sm">{item.label}</span>
              </button>
            ))}
          </nav>
        </div>

        {/* Exit Button */}
        <button
          onClick={onExit}
          className="admin-nav-item"
          style={{ color: 'var(--priority-critical)', marginTop: 'auto', border: '1px solid transparent' }}
        >
          <svg className="admin-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
            <polyline points="16 17 21 12 16 7" />
            <line x1="21" y1="12" x2="9" y2="12" />
          </svg>
          <span className="text-sm font-semibold">Exit Control</span>
        </button>
      </aside>

      {/* Primary Work Area */}
      <section className="admin-content">
        {renderActiveView()}
      </section>
    </div>
  );
};

export default AdminPanel;
