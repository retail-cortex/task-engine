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

interface ButtonProps {
  node: A2UIComponent;
  onActionTrigger?: A2UIActionHandler;
  formState?: Record<string, any>;
}

const Button: React.FC<ButtonProps> = ({ node, onActionTrigger, formState }: ButtonProps) => {
  const getButtonStyle = () => {
    if (node.style === 'critical') {
      return {
        borderColor: 'var(--priority-critical)',
        color: 'var(--priority-critical)'
      };
    }
    if (node.style === 'outline') {
      return {
        borderColor: 'var(--panel-border-hover)',
        color: 'var(--text-secondary)'
      };
    }
    // Default to premium neon primary outline
    return {};
  };

  const handleClick = () => {
    if (onActionTrigger && node.action) {
      const resolvedData = { ...(node.actionData || {}) };
      // Scan and resolve any path-bound values (e.g. /store_switcher/selected)
      // using the actual selected values stored in formState
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
    <button
      type="button"
      className="a2ui-btn-action"
      style={getButtonStyle()}
      onClick={handleClick}
    >
      {node.label || 'Submit'}
    </button>
  );
};

export default Button;
