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
import type { A2UIComponent } from './types';

interface InputProps {
  node: A2UIComponent;
  value?: string;
  onChange: (val: string) => void;
}

const Input: React.FC<InputProps> = ({ node, value = '', onChange }: InputProps) => {
  // Auto-initialize the form state to default value on mount if not already present
  useEffect(() => {
    if (!value && node.value) {
      onChange(node.value);
    }
  }, [node.value, value, onChange]);

  return (
    <div className="a2ui-form-group" style={{ display: 'flex', flexDirection: 'column', gap: 6, width: '100%' }}>
      {node.label && (
        <label style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 500, letterSpacing: '0.3px' }}>
          {node.label}
        </label>
      )}
      <input
        type="text"
        className="a2ui-form-input"
        placeholder={node.placeholder || ''}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        style={{
          width: '100%',
          background: 'rgba(5, 6, 12, 0.45)',
          border: '1px solid var(--panel-border)',
          borderRadius: '6px',
          padding: '8px 12px',
          color: 'var(--text-primary)',
          fontSize: '0.85rem',
          outline: 'none',
          fontFamily: 'inherit',
          transition: 'border-color 0.2s ease-in-out'
        }}
      />
    </div>
  );
};

export default Input;
