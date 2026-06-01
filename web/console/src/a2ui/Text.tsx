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

interface TextProps {
  node: A2UIComponent;
}

const Text: React.FC<TextProps> = ({ node }: TextProps) => {
  const getStyleStyles = () => {
    if (node.style === 'critical') return { color: 'var(--priority-critical)' };
    if (node.style === 'standard') return { color: 'var(--priority-standard)' };
    if (node.style === 'muted') return { color: 'var(--text-muted)', fontSize: '0.8rem' };
    if (node.style === 'secondary') return { color: 'var(--text-secondary)', fontSize: '0.85rem' };
    return { color: 'var(--text-primary)' };
  };

  return (
    <div style={{ ...getStyleStyles(), lineHeight: 1.45 }}>
      {node.content}
    </div>
  );
};

export default Text;
