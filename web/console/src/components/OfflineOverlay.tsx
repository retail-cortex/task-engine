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

import React from 'react';
import groundedBanner from '../assets/grounded_intelligence_banner.png';

interface OfflineOverlayProps {
  backendActive: boolean;
}

const OfflineOverlay: React.FC<OfflineOverlayProps> = ({ backendActive }) => {
  if (backendActive) return null;

  return (
    <>
      <div style={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: '100vw',
        height: '100vh',
        background: 'rgba(5, 6, 12, 0.85)',
        backdropFilter: 'blur(20px)',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        zIndex: 9999
      }}>
        <div className="panel-card" style={{ width: '450px', padding: '32px', textAlign: 'center', gap: '24px', border: '1px solid var(--priority-critical-glow)' }}>
          <div className="pulse-indicator" style={{ width: '12px', height: '12px', background: 'var(--priority-critical)', boxShadow: '0 0 12px var(--priority-critical)' }}></div>
          <h2 className="brand-title" style={{ color: 'var(--priority-critical)', fontSize: '1.4rem', margin: '8px 0 0 0' }}>NEXUS OPERATIONS OFFLINE</h2>
          <p style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', lineHeight: 1.5 }}>
            The backing multi-node task API server on port <strong>8080</strong> is currently unreachable.
            Start your local cluster runtime servers before proceeding:
          </p>
          <code style={{ background: 'var(--input-bg)', padding: '10px 14px', borderRadius: '6px', fontSize: '0.8rem', color: 'var(--accent-primary)', border: '1px solid var(--panel-border)', fontFamily: 'var(--font-mono)' }}>
            bazel run //:dev_server
          </code>
          <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            AlloyDB persistent ledgers, pgvector RAGs, and Maker/Checker task overrides are locked.
          </p>
        </div>
      </div>

      {/* Decorative radial vector backdrop asset */}
      <div style={{
        position: 'absolute',
        bottom: '2%',
        right: '2%',
        width: '80px',
        height: '80px',
        opacity: 0.15,
        backgroundImage: `url(${groundedBanner})`,
        backgroundSize: 'contain',
        backgroundRepeat: 'no-repeat',
        pointerEvents: 'none',
        zIndex: -1
      }}></div>
    </>
  );
};

export default OfflineOverlay;
