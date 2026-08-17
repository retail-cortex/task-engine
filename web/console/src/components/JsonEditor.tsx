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

import React, { useState, useEffect } from 'react';

interface JsonEditorProps {
  value: string;
  onChange: (val: string) => void;
  placeholder?: string;
  onValidate?: (isValid: boolean) => void;
}

export const JsonEditor: React.FC<JsonEditorProps> = ({ value, onChange, placeholder, onValidate }) => {
  const [activeTab, setActiveTab] = useState<'edit' | 'tree'>('edit');
  const [isValid, setIsValid] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [parsedData, setParsedData] = useState<any>(null);

  // Real-time validation
  useEffect(() => {
    if (!value || value.trim() === '') {
      setIsValid(true);
      setErrorMsg(null);
      setParsedData({});
      if (onValidate) onValidate(true);
      return;
    }
    try {
      const parsed = JSON.parse(value);
      setIsValid(true);
      setErrorMsg(null);
      setParsedData(parsed);
      if (onValidate) onValidate(true);
    } catch (err: any) {
      setIsValid(false);
      setErrorMsg(err.message);
      if (onValidate) onValidate(false);
    }
  }, [value, onValidate]);

  const handleBeautify = () => {
    try {
      const parsed = JSON.parse(value);
      onChange(JSON.stringify(parsed, null, 2));
    } catch (err: any) {
      alert(`Cannot format: ${err.message}`);
    }
  };

  const handleMinify = () => {
    try {
      const parsed = JSON.parse(value);
      onChange(JSON.stringify(parsed));
    } catch (err: any) {
      alert(`Cannot minify: ${err.message}`);
    }
  };

  const handleClear = () => {
    if (window.confirm('Are you sure you want to clear the metadata?')) {
      onChange('{}');
    }
  };

  // --- RECURSIVE TREE VIEW RENDERER ---
  const TreeNode: React.FC<{ name: string | number; value: any; isLast?: boolean }> = ({ name, value: nodeVal, isLast }) => {
    const [isExpanded, setIsExpanded] = useState(true);

    const toggleExpand = (e: React.MouseEvent) => {
      e.stopPropagation();
      setIsExpanded(!isExpanded);
    };

    const renderPrimitiveValue = (val: any) => {
      if (val === null) return <span className="json-null">null</span>;
      if (typeof val === 'boolean') return <span className="json-boolean">{val ? 'true' : 'false'}</span>;
      if (typeof val === 'number') return <span className="json-number">{val}</span>;
      return <span className="json-string">"{String(val)}"</span>;
    };

    const label = typeof name === 'number' ? name : `"${name}"`;

    if (nodeVal && typeof nodeVal === 'object') {
      const isArray = Array.isArray(nodeVal);
      const keys = Object.keys(nodeVal);
      const isEmpty = keys.length === 0;

      return (
        <div style={{ marginLeft: '16px', fontFamily: 'var(--font-mono)', fontSize: '0.85rem' }}>
          <div 
            onClick={toggleExpand}
            style={{ 
              display: 'flex', 
              alignItems: 'center', 
              cursor: 'pointer', 
              padding: '2px 4px', 
              borderRadius: '4px',
              userSelect: 'none'
            }}
            className="hover:bg-white/5"
          >
            <span style={{ 
              display: 'inline-block', 
              width: '12px', 
              marginRight: '6px', 
              transition: 'transform 0.15s ease',
              transform: isExpanded ? 'rotate(90deg)' : 'rotate(0deg)',
              color: 'var(--text-secondary)'
            }}>
              ▶
            </span>
            <span className="json-key" style={{ marginRight: '6px' }}>{label}:</span>
            <span className="json-label">{isArray ? `Array[${keys.length}]` : `Object {${keys.length}}`}</span>
          </div>

          {isExpanded && !isEmpty && (
            <div style={{ borderLeft: '1px dashed var(--panel-border)', marginLeft: '10px', paddingLeft: '8px' }}>
              {keys.map((key, index) => (
                <TreeNode 
                  key={key} 
                  name={isArray ? Number(key) : key} 
                  value={nodeVal[key]} 
                  isLast={index === keys.length - 1} 
                />
              ))}
            </div>
          )}
        </div>
      );
    }

    // Primitive Node
    return (
      <div style={{ marginLeft: '16px', padding: '2px 4px', fontFamily: 'var(--font-mono)', fontSize: '0.85rem', display: 'flex', alignItems: 'baseline' }}>
        <span className="json-quote" style={{ marginRight: '6px' }}>{label}:</span>
        <span style={{ marginRight: '4px' }}>{renderPrimitiveValue(nodeVal)}</span>
        {!isLast && <span style={{ color: 'var(--text-muted)' }}>,</span>}
      </div>
    );
  };

  return (
    <div 
      className="flex flex-col rounded-xl border overflow-hidden" 
      style={{ 
        borderColor: 'var(--panel-border)', 
        background: 'var(--input-bg)',
        boxShadow: '0 4px 12px rgba(0,0,0,0.03)'
      }}
    >
      <style>{`
        .json-tab {
          padding: 8px 16px;
          font-size: 0.8rem;
          font-weight: 600;
          cursor: pointer;
          border-bottom: 2px solid transparent;
          transition: all 0.2s ease;
          color: var(--text-secondary);
        }
        .json-tab.active {
          color: var(--accent-primary);
          border-bottom-color: var(--accent-primary);
          background: rgba(99, 102, 241, 0.05);
        }
        .json-btn {
          padding: 4px 10px;
          font-size: 0.75rem;
          font-weight: 600;
          border-radius: 6px;
          border: 1px solid var(--panel-border);
          background: var(--panel-glass);
          color: var(--text-secondary);
          cursor: pointer;
          transition: all 0.2s ease;
        }
        .json-btn:hover:not(:disabled) {
          color: var(--text-primary);
          border-color: var(--text-secondary);
          background: rgba(255,255,255,0.03);
        }
        .json-btn:disabled {
          opacity: 0.4;
          cursor: not-allowed;
        }
        .json-tree-container {
          padding: 16px;
          overflow-y: auto;
          max-height: 320px;
          background: #1e1e2e; /* Slate dark */
          border-radius: 0 0 12px 12px;
        }
        html[data-theme="light"] .json-tree-container {
          background: #f8fafc; /* Cool white in light mode */
          border-top: 1px solid var(--panel-border);
        }
        
        /* Syntax highlighting colors - Dark Mode Default */
        .json-key { color: #8be9fd; }
        .json-string { color: #50fa7b; }
        .json-number { color: #ffb86c; }
        .json-boolean { color: #bd93f9; }
        .json-null { color: #ff79c6; }
        .json-quote { color: #f1fa8c; }
        .json-label { color: var(--text-muted); }

        /* Syntax highlighting colors - Light Mode Override */
        html[data-theme="light"] .json-key { color: #0284c7; }
        html[data-theme="light"] .json-string { color: #16a34a; }
        html[data-theme="light"] .json-number { color: #ea580c; }
        html[data-theme="light"] .json-boolean { color: #7c3aed; }
        html[data-theme="light"] .json-null { color: #db2777; }
        html[data-theme="light"] .json-quote { color: #475569; }
        html[data-theme="light"] .json-label { color: var(--text-secondary); }
      `}</style>

      {/* --- HEADER CONTROLS --- */}
      <div className="flex justify-between items-center px-4 border-b" style={{ borderColor: 'var(--panel-border)', background: 'rgba(255,255,255,0.01)' }}>
        {/* Navigation Tabs */}
        <div className="flex">
          <div 
            className={`json-tab ${activeTab === 'edit' ? 'active' : ''}`}
            onClick={() => setActiveTab('edit')}
          >
            Raw Editor
          </div>
          <div 
            className={`json-tab ${activeTab === 'tree' ? 'active' : ''}`}
            onClick={() => setActiveTab('tree')}
          >
            Interactive Tree
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-2 py-2">
          <button type="button" className="json-btn" onClick={handleBeautify} disabled={!isValid}>Format</button>
          <button type="button" className="json-btn" onClick={handleMinify} disabled={!isValid}>Minify</button>
          <button type="button" className="json-btn" style={{ borderColor: 'rgba(239, 68, 68, 0.3)', color: 'rgb(239, 68, 68)' }} onClick={handleClear}>Clear</button>
        </div>
      </div>

      {/* --- BODY AREA --- */}
      <div className="relative" style={{ minHeight: '160px' }}>
        
        {/* TAB 1: RAW EDITOR */}
        {activeTab === 'edit' && (
          <div className="flex flex-col h-full">
            <textarea
              className="w-full p-4 font-mono text-sm"
              style={{ 
                minHeight: '160px', 
                resize: 'vertical',
                color: 'var(--text-primary)', // Guaranteed theme-aware primary text color!
                background: 'transparent', // Inherit wrapper background
                border: 'none',
                outline: 'none',
                lineHeight: '1.5'
              }}
              placeholder={placeholder || '{}'}
              value={value}
              onChange={(e) => onChange(e.target.value)}
            />
          </div>
        )}

        {/* TAB 2: INTERACTIVE TREE VIEW */}
        {activeTab === 'tree' && (
          <div className="json-tree-container">
            {isValid ? (
              parsedData && Object.keys(parsedData).length > 0 ? (
                <div style={{ paddingLeft: '8px' }}>
                  <span className="json-label" style={{ fontFamily: 'var(--font-mono)', fontSize: '0.8rem' }}>Root</span>
                  {Object.keys(parsedData).map((key, index) => (
                    <TreeNode 
                      key={key} 
                      name={key} 
                      value={parsedData[key]} 
                      isLast={index === Object.keys(parsedData).length - 1} 
                    />
                  ))}
                </div>
              ) : (
                <div className="text-center py-8 text-sm text-muted font-mono" style={{ color: 'var(--text-muted)' }}>Empty Object {"{}"}</div>
              )
            ) : (
              <div className="flex flex-col items-center justify-center py-8 text-center px-4">
                <span className="text-xs font-mono" style={{ color: 'var(--priority-critical)', marginBottom: 8 }}>
                  ⚠️ JSON Parsing Error
                </span>
                <span className="text-xs text-muted font-mono" style={{ maxWidth: '80%', color: 'var(--text-muted)' }}>
                  Please fix the syntax errors in the Raw Editor to enable the interactive tree view.
                </span>
              </div>
            )}
          </div>
        )}
      </div>

      {/* --- FOOTER STATUS BAR --- */}
      <div 
        className="flex justify-between items-center px-4 py-2 border-t text-xs font-mono" 
        style={{ 
          borderColor: 'var(--panel-border)', 
          background: 'rgba(0,0,0,0.02)' // Extremely light background overlay
        }}
      >
        <div>
          <span 
            className="flex items-center gap-1.5" 
            style={{ 
              color: isValid ? 'var(--priority-standard)' : 'var(--priority-critical)', // Beautiful native theme-aware green/red
              fontWeight: 600 
            }}
          >
            <span 
              style={{ 
                display: 'inline-block', 
                width: 6, 
                height: 6, 
                borderRadius: '50%',
                background: isValid ? 'var(--priority-standard)' : 'var(--priority-critical)', // Glowing dot matching the status
                animation: isValid ? 'none' : 'pulse 1.5s infinite'
              }}
            ></span>
            {isValid ? 'Valid JSONB Metadata' : 'Invalid Syntax'}
          </span>
        </div>
        <div style={{ color: 'var(--text-muted)', maxWidth: '60%', textAlign: 'right', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {errorMsg ? errorMsg : `Size: ${value.length} bytes`}
        </div>
      </div>
    </div>
  );
};
