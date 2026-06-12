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
import type { A2UIComponent, A2UIActionHandler } from './types';

interface CheckBoxProps {
  node: A2UIComponent;
  value?: boolean | string;
  onChange: (val: boolean) => void;
  onActionTrigger?: A2UIActionHandler;
  formState?: Record<string, any>;
}

const CheckBox = ({ node, value = false, onChange, onActionTrigger, formState }: CheckBoxProps) => {
  // Coerce value to boolean
  const isChecked = typeof value === 'string' ? value === 'true' : !!value;

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    const nextChecked = !isChecked;
    onChange(nextChecked);

    if (onActionTrigger && node.action) {
      const resolvedData = { ...(node.actionData || {}) };
      // Inject the current checked state
      resolvedData['checked'] = nextChecked;
      resolvedData['value'] = nextChecked;

      // Scan and resolve path-bound values from formState
      Object.keys(resolvedData).forEach(key => {
        const val = resolvedData[key];
        if (typeof val === 'string' && val.startsWith('/') && formState && val in formState) {
          resolvedData[key] = formState[val];
        }
      });
      onActionTrigger(node.action, resolvedData);
    }
  };

  return (
    <div 
      className="a2ui-form-group a2ui-checkbox-container" 
      style={{ 
        display: 'flex', 
        alignItems: 'center', 
        gap: 10, 
        width: '100%',
        padding: '6px 4px',
        cursor: 'pointer',
        userSelect: 'none'
      }}
      onClick={handleClick}
    >
      <div
        style={{
          width: '20px',
          height: '20px',
          borderRadius: '4px',
          border: isChecked ? '2px solid #3870FF' : '2px solid var(--panel-border)',
          background: isChecked ? '#3870FF' : 'rgba(5, 6, 12, 0.45)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
          boxShadow: isChecked ? '0 0 8px rgba(56, 112, 255, 0.4)' : 'none'
        }}
      >
        {isChecked && (
          <svg 
            width="12" 
            height="10" 
            viewBox="0 0 12 10" 
            fill="none" 
            xmlns="http://www.w3.org/2000/svg"
            style={{ stroke: '#ffffff', strokeWidth: 2, strokeLinecap: 'round', strokeLinejoin: 'round' }}
          >
            <path d="M1.5 5L4.5 8L10.5 1.5" />
          </svg>
        )}
      </div>
      {node.label && (
        <span 
          style={{ 
            fontSize: '0.875rem', 
            color: isChecked ? 'var(--text-primary)' : 'var(--text-secondary)',
            fontWeight: 500, 
            transition: 'color 0.2s ease',
            lineHeight: '1.2'
          }}
        >
          {node.label}
        </span>
      )}
    </div>
  );
};

export default CheckBox;
