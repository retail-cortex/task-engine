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

import React, { useState, useEffect, useRef } from 'react';

interface DashboardHeaderProps {
  userName: string;
  userEmail: string;
  userPicture: string | null;
  theme: 'light' | 'dark';
  setTheme: React.Dispatch<React.SetStateAction<'light' | 'dark'>>;
  backendActive: boolean;
  schedulerNodeID: string;
  schedulerLeader: boolean;
  handleSignOut: () => void;
  // Dynamic role-based site switcher properties:
  userRole: string;
  userSites: any[];
  allSites: any[];
  activeSiteID: string;
  onSiteChange?: (siteID: string) => void;
}

const DashboardHeader: React.FC<DashboardHeaderProps> = ({
  userName,
  userEmail,
  userPicture,
  theme,
  setTheme,
  backendActive,
  schedulerNodeID,
  schedulerLeader,
  handleSignOut,
  userRole,
  userSites,
  allSites,
  activeSiteID,
  onSiteChange
}: DashboardHeaderProps) => {

  const [isDropdownOpen, setIsDropdownOpen] = useState<boolean>(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Bind mouse-click event listener to handle click-outs and close settings card
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  const triggerSignOut = () => {
    setIsDropdownOpen(false);
    handleSignOut();
  };

  const getSiteOptions = () => {
    const list = (userRole === 'ADMIN' || userRole === 'REGION_MANAGER') ? allSites : userSites;
    if (list.length === 0) {
      return [{ id: activeSiteID, name: activeSiteID === '55555555-5555-5555-5555-555555550000' ? 'OmniMart Dallas Store #1000' : 'Volt & Vine Seattle' }];
    }
    return list;
  };

  const sitesOptions = getSiteOptions();
  const currentSite = sitesOptions.find((s: any) => (s.id || s.ID) === activeSiteID);
  const activeSiteLabel = currentSite ? (currentSite.name || currentSite.Name) : 'Operational Store Context';

  return (
    <header className="dashboard-header">
      <div className="brand-section">
        <div className="pulse-indicator"></div>
        <h1 className="brand-title">NEXUS INTEGRATION ENGINE HUB</h1>
        
        {sitesOptions.length > 1 && (userRole === 'ADMIN' || userRole === 'REGION_MANAGER' || userRole === 'SITE_MANAGER') ? (
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, position: 'relative' }}>
            <span style={{ fontSize: '0.78rem', color: 'var(--text-secondary)', fontWeight: 600 }}>Store:</span>
            <select
              className="site-meta-pill"
              style={{
                background: 'var(--input-bg)',
                border: '1px solid var(--panel-border)',
                color: 'var(--accent-primary)',
                padding: '2px 10px',
                borderRadius: 6,
                cursor: 'pointer',
                fontWeight: 700,
                fontSize: '0.78rem',
                fontFamily: 'inherit',
                outline: 'none',
                textTransform: 'uppercase',
                transition: 'all 0.2s ease-in-out'
              }}
              value={activeSiteID}
              onChange={(e) => onSiteChange && onSiteChange(e.target.value)}
            >
              {sitesOptions.map((s: any) => (
                <option key={s.id || s.ID} value={s.id || s.ID} style={{ background: '#0c0e1c', color: 'var(--text-primary)' }}>
                  {s.name || s.Name}
                </option>
              ))}
            </select>
          </div>
        ) : (
          <span className="site-meta-pill">Store: {activeSiteLabel}</span>
        )}

        <span className="site-meta-pill" style={{ textTransform: 'uppercase', background: 'var(--priority-high-glow)', color: 'var(--priority-high)', border: '1px solid var(--priority-high)', fontWeight: 700 }}>
          {userRole}
        </span>

        {backendActive ? (
          <span className="site-meta-pill" style={{ borderColor: 'var(--priority-standard)', color: 'var(--priority-standard)' }}>API Active</span>
        ) : (
          <span className="site-meta-pill" style={{ borderColor: 'var(--priority-high)', color: 'var(--priority-high)' }}>Local Sandbox</span>
        )}
      </div>
      <div className="header-user-card">
        <div className="profile-container" ref={dropdownRef}>
          <div className="avatar-wrapper" onClick={() => setIsDropdownOpen(prev => !prev)} title="Account Settings">
            {userPicture ? (
              <img src={userPicture} alt={userName} className="user-avatar-img" />
            ) : (
              <div className="user-avatar">
                {userName.slice(0, 1).toUpperCase()}
              </div>
            )}
          </div>

          {isDropdownOpen && (
            <div className="profile-dropdown-menu">
              <div className="dropdown-user-info">
                <h4 className="dropdown-user-name">{userName}</h4>
                <p className="dropdown-user-email">{userEmail}</p>
              </div>

              <div className="appearance-switch-row" onClick={() => setTheme(t => (t === 'light' ? 'dark' : 'light'))}>
                <div className="switch-label">
                  {theme === 'light' ? (
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="dropdown-menu-icon" style={{ color: 'var(--accent-primary)' }}>
                      <circle cx="12" cy="12" r="5" />
                      <line x1="12" y1="1" x2="12" y2="3" />
                      <line x1="12" y1="21" x2="12" y2="23" />
                      <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
                      <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
                      <line x1="1" y1="12" x2="3" y2="12" />
                      <line x1="21" y1="12" x2="23" y2="12" />
                      <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
                      <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
                    </svg>
                  ) : (
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="dropdown-menu-icon" style={{ color: 'var(--accent-primary)' }}>
                      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
                    </svg>
                  )}
                  <span>{theme === 'light' ? 'Light Theme' : 'Dark Theme'}</span>
                </div>
                <label className="switch-control" onClick={(e) => e.stopPropagation()}>
                  <input 
                    type="checkbox" 
                    checked={theme === 'dark'} 
                    onChange={() => setTheme(t => (t === 'light' ? 'dark' : 'light'))} 
                  />
                  <span className="switch-slider"></span>
                </label>
              </div>

              {/* Leader Election status widgets inside settings panel */}
              <div className="appearance-switch-row" style={{ cursor: 'default' }}>
                <div className="switch-label">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="dropdown-menu-icon" style={{ color: 'var(--priority-standard)' }}>
                    <rect x="2" y="2" width="20" height="8" rx="2" ry="2" />
                    <rect x="2" y="14" width="20" height="8" rx="2" ry="2" />
                    <line x1="6" y1="6" x2="6.01" y2="6" />
                    <line x1="6" y1="18" x2="6.01" y2="18" />
                  </svg>
                  <span>Node: <strong>{schedulerNodeID}</strong></span>
                </div>
                {schedulerLeader && (
                  <span className="panel-title-count" style={{ background: 'var(--priority-critical-glow)', color: 'var(--priority-critical)', fontSize: '0.7rem' }}>
                    LEADER
                  </span>
                )}
              </div>

              <div className="dropdown-divider"></div>

              <button type="button" className="dropdown-menu-item" onClick={triggerSignOut}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="dropdown-menu-icon" style={{ color: 'var(--priority-critical)' }}>
                    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                    <polyline points="16 17 21 12 16 7" />
                    <line x1="21" y1="12" x2="9" y2="12" />
                  </svg>
                  <span>Clock Out</span>
                </div>
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
};

export default DashboardHeader;
