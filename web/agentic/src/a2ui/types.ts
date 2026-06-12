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

export interface A2UIComponent {
  type: 'card' | 'column' | 'row' | 'text' | 'button' | 'table' | 'select' | 'input' | 'canvas' | 'checkbox';
  id?: string;
  title?: string;
  label?: string;
  content?: string;
  style?: 'primary' | 'secondary' | 'critical' | 'standard' | 'outline' | 'muted';
  align?: 'start' | 'center' | 'end' | 'between';
  gap?: number;
  rows?: Array<{ label: string; value: string; action?: string; actionData?: any }>;
  children?: A2UIComponent[];
  action?: string;
  actionData?: any;
  
  // Dynamic Form bindings parameters
  name?: string;
  placeholder?: string;
  options?: Array<{ label: string; value: string }>;
  value?: string;

  // Custom Canvas parameters
  layout?: 'linear' | 'boutique' | 'racetrack';
  beacon?: { x: number; y: number; name?: string };
}

export type A2UIActionHandler = (actionType: string, actionData: any) => void;

