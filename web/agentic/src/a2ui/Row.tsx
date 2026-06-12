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

interface RowProps {
  node: A2UIComponent;
  renderChild: (child: A2UIComponent, idx: number) => React.ReactNode;
}

const Row: React.FC<RowProps> = ({ node, renderChild }: RowProps) => {
  const getAlignStyles = () => {
    if (node.align === 'center') return 'center';
    if (node.align === 'end') return 'flex-end';
    if (node.align === 'between') return 'space-between';
    return 'flex-start';
  };

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'row',
      alignItems: 'center',
      justifyContent: getAlignStyles(),
      gap: node.gap !== undefined ? node.gap : 8,
      width: '100%'
    }}>
      {node.children?.map((child, idx) => renderChild(child, idx))}
    </div>
  );
};

export default Row;
