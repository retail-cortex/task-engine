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

interface TableProps {
  node: A2UIComponent;
}

const Table: React.FC<TableProps> = ({ node }: TableProps) => {
  if (!node.rows || node.rows.length === 0) return null;

  return (
    <table className="a2ui-table">
      <tbody>
        {node.rows.map((row, idx) => {
          const isAction = row.label.toLowerCase() === 'actions' || row.label.toLowerCase() === 'action';
          
          return (
            <tr key={idx}>
              <td className="a2ui-label">{row.label}</td>
              <td className="a2ui-value">
                {isAction ? (
                  // Align actions right using a flex wrapper container with justify-end!
                  <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
                    {row.value}
                  </div>
                ) : (
                  <span>{row.value}</span>
                )}
              </td>
            </tr>
          );
        })}
      </tbody>
    </table>
  );
};

export default Table;
