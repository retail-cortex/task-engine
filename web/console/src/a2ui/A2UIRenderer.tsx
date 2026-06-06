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

import React, { useState } from 'react';
import type { A2UIComponent, A2UIActionHandler } from './types';
import Card from './Card';
import Column from './Column';
import Row from './Row';
import Table from './Table';
import Button from './Button';
import Text from './Text';
import Input from './Input';
import Select from './Select';
import Canvas from './Canvas';

interface A2UIRendererProps {
  component: A2UIComponent;
  onActionTrigger?: A2UIActionHandler;
}

const A2UIRenderer: React.FC<A2UIRendererProps> = ({ component, onActionTrigger }: A2UIRendererProps) => {
  // Centralised Dynamic Form State map (recursively shared down component hierarchies via closures!)
  const [formState, setFormState] = useState<Record<string, any>>({});

  const renderChild = (child: A2UIComponent, idx: number): React.ReactNode => {
    switch (child.type) {
      case 'card':
        return (
          <Card 
            key={child.id || `card-${idx}`} 
            node={child} 
            renderChild={renderChild} 
          />
        );
      case 'column':
        return (
          <Column 
            key={child.id || `col-${idx}`} 
            node={child} 
            renderChild={renderChild} 
          />
        );
      case 'row':
        return (
          <Row 
            key={child.id || `row-${idx}`} 
            node={child} 
            renderChild={renderChild} 
          />
        );
      case 'table':
        return <Table key={child.id || `tbl-${idx}`} node={child} />;
      case 'button':
        return (
          <Button 
            key={child.id || `btn-${idx}`} 
            node={child} 
            onActionTrigger={onActionTrigger} 
            formState={formState}
          />
        );
      case 'text':
        return <Text key={child.id || `txt-${idx}`} node={child} />;
      case 'input':
        return (
          <Input 
            key={child.id || `input-${idx}`}
            node={child}
            value={formState[child.name || '']}
            onChange={(val) => {
              if (child.name) {
                setFormState(prev => ({ ...prev, [child.name!]: val }));
              }
            }}
          />
        );
      case 'select':
        return (
          <Select 
            key={child.id || `select-${idx}`}
            node={child}
            value={formState[child.name || '']}
            onChange={(val) => {
              if (child.name) {
                setFormState(prev => ({ ...prev, [child.name!]: val }));
              }
            }}
          />
        );
      case 'canvas':
        return <Canvas key={child.id || `canvas-${idx}`} node={child} />;
      default:
        return null;
    }
  };

  return <>{renderChild(component, 0)}</>;
};

export default A2UIRenderer;
export type { A2UIComponent, A2UIActionHandler };
