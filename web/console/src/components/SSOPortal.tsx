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

import React, { useEffect } from 'react';
import groundedBanner from '../assets/grounded_intelligence_banner.png';

interface SSOPortalProps {
  theme: 'light' | 'dark';
  setTheme: React.Dispatch<React.SetStateAction<'light' | 'dark'>>;
  googleClientID: string;
  handleGoogleCredentialResponse: (response: any) => void;
}

const SSOPortal: React.FC<SSOPortalProps> = ({
  theme,
  setTheme,
  googleClientID,
  handleGoogleCredentialResponse
}: SSOPortalProps) => {

  // Google Identity Services (GIS) Sign-In Initializer inside authenticated portal limits
  useEffect(() => {
    const interval = setInterval(() => {
      const google = (window as any).google;
      if (google && google.accounts && google.accounts.id) {
        clearInterval(interval);
        
        google.accounts.id.initialize({
          client_id: googleClientID,
          callback: handleGoogleCredentialResponse,
          auto_select: false,
          use_fedcm_for_prompt: false
        });

        const targetBtn = document.getElementById("google-signin-btn");
        if (targetBtn) {
          google.accounts.id.renderButton(
            targetBtn,
            { theme: "filled_blue", size: "large", width: 280 }
          );
        }
      }
    }, 300);
    
    return () => clearInterval(interval);
  }, [googleClientID, handleGoogleCredentialResponse]);

  return (
    <div id="root" style={{ height: '100vh', justifyContent: 'center', alignItems: 'center', position: 'relative' }}>
      <div style={{ position: 'absolute', top: 24, right: 24, zIndex: 999 }}>
        <button
          type="button"
          className="theme-toggle-btn"
          onClick={() => setTheme(t => (t === 'light' ? 'dark' : 'light'))}
          title={theme === 'light' ? 'Switch to Dark Mode' : 'Switch to Light Mode'}
        >
          {theme === 'light' ? (
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="theme-toggle-icon">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
            </svg>
          ) : (
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="theme-toggle-icon">
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
          )}
        </button>
      </div>
      <div className="panel-card" style={{ width: '400px', padding: '32px', textAlign: 'center', gap: '24px' }}>
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '8px' }}>
          <div className="pulse-indicator" style={{ width: '12px', height: '12px', background: 'var(--neon-primary)', boxShadow: '0 0 12px var(--neon-primary)' }}></div>
          <h1 className="brand-title" style={{ fontSize: '1.8rem', margin: '8px 0 0 0' }}>Gemini Tasksing Portal</h1>
          <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>Secure Multi-Node SSO Authentication</span>
        </div>

        {/* Premium illustration banner watermark */}
        <div style={{
          width: '100%',
          height: '140px',
          borderRadius: '10px',
          backgroundImage: `url(${groundedBanner})`,
          backgroundSize: 'cover',
          backgroundPosition: 'center',
          border: '1px solid var(--panel-border)'
        }}></div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <p style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', lineHeight: 1.45 }}>
            Authorized retail and logistics personnel only. Clock in using your Google corporate identity.
          </p>
          
          {/* GIS SDK Target Mount Box */}
          <div id="google-signin-btn" style={{ minHeight: '46px', display: 'flex', justifyContent: 'center' }}></div>
        </div>


      </div>
    </div>
  );
};

export default SSOPortal;
