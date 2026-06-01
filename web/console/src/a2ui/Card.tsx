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
import type { A2UIComponent } from './types';

interface CardProps {
  node: A2UIComponent;
  renderChild: (child: A2UIComponent, idx: number) => React.ReactNode;
}

const Card: React.FC<CardProps> = ({ node, renderChild }: CardProps) => {
  const getHeaderBorderColor = () => {
    if (node.style === 'critical') return 'rgba(248, 113, 113, 0.25)'; // Critical prioritised alarms
    if (node.style === 'standard') return 'rgba(52, 211, 153, 0.25)'; // Standard sweeps
    return 'rgba(99, 102, 241, 0.25)'; // Brand primary outline
  };

  const getHeaderGlowStyle = () => {
    if (node.style === 'critical') return 'var(--priority-critical)';
    if (node.style === 'standard') return 'var(--priority-standard)';
    return 'var(--accent-primary)';
  };

  return (
    <div className="a2ui-card-container">
      <div 
        className="a2ui-header"
        style={{ borderColor: getHeaderBorderColor() }}
      >
        <span>{node.title || 'OPERATIONAL DISPATCH WIDGET'}</span>
        <span 
          style={{ 
            color: getHeaderGlowStyle(), 
            fontFamily: 'var(--font-mono)',
            textTransform: 'uppercase',
            fontSize: '0.8rem',
            fontWeight: 600
          }}
        >
          {node.style || 'INFO'}
        </span>
      </div>
      <div className="a2ui-body">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {node.children?.map((child, idx) => renderChild(child, idx))}
        </div>
      </div>
    </div>
  );
};

export default Card;
