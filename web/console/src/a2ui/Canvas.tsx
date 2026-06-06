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
import type { A2UIComponent } from './types';

interface CanvasProps {
  node: A2UIComponent;
}

const Canvas: React.FC<CanvasProps> = ({ node }: CanvasProps) => {
  const layout = node.layout || 'linear';
  const focal = node.beacon;

  const [zoom, setZoom] = useState<number>(1.25);
  const [pan, setPan] = useState<{ x: number; y: number }>({ x: -10, y: -10 });
  const [isDragging, setIsDragging] = useState<boolean>(false);
  const [dragStart, setDragStart] = useState<{ x: number; y: number }>({ x: 0, y: 0 });

  const handleMapMouseDown = (e: React.MouseEvent<SVGSVGElement>) => {
    setIsDragging(true);
    setDragStart({ x: e.clientX - pan.x, y: e.clientY - pan.y });
  };

  const handleMapMouseMove = (e: React.MouseEvent<SVGSVGElement>) => {
    if (!isDragging) return;
    setPan({
      x: e.clientX - dragStart.x,
      y: e.clientY - dragStart.y
    });
  };

  const handleMapMouseUpOrLeave = () => {
    setIsDragging(false);
  };

  const handleMapWheel = (e: React.WheelEvent<SVGSVGElement>) => {
    e.preventDefault();
    const scaleFactor = 0.05;
    const nextZoom = e.deltaY < 0 ? zoom + scaleFactor : zoom - scaleFactor;
    setZoom(Math.max(0.7, Math.min(nextZoom, 3.5)));
  };

  const resetMapTransform = () => {
    setZoom(1.25);
    setPan({ x: -10, y: -10 });
  };

  return (
    <div className="blueprint-wrapper" style={{ position: 'relative', width: '100%', margin: '8px 0' }}>
      <h4 className="blueprint-title" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span>Store Digital Twin (Aisle Layout)</span>
        <span style={{ fontSize: '0.8rem', color: 'var(--priority-high)', fontWeight: 600 }}>
          Spatial Focus: {focal ? (focal.name || 'System beacon') : 'System Center'}
        </span>
      </h4>
      <div className="blueprint-container" style={{ position: 'relative', overflow: 'hidden', borderRadius: 8, border: '1px solid var(--panel-border)', background: 'var(--blueprint-bg)', height: 260 }}>
        <svg 
          className="blueprint-grid-floor" 
          viewBox="0 0 200 150"
          onMouseDown={handleMapMouseDown}
          onMouseMove={handleMapMouseMove}
          onMouseUp={handleMapMouseUpOrLeave}
          onMouseLeave={handleMapMouseUpOrLeave}
          onWheel={handleMapWheel}
          style={{ 
            width: '100%', 
            height: '100%', 
            cursor: isDragging ? 'grabbing' : 'grab',
            userSelect: 'none'
          }}
        >
          {/* Transformation grouping wrapper */}
          <g transform={`translate(${pan.x}, ${pan.y}) scale(${zoom})`}>
            {/* Outer perimeter outline bounds */}
            <rect x="2" y="2" width="196" height="146" rx="3" fill="none" stroke="var(--blueprint-grid-stroke)" strokeWidth="1" strokeDasharray="4,2" />

            {layout === 'boutique' && (
              <>
                {/* A. Backrooms & Cargo Receiving (Top Left / Bottom Left) */}
                <rect className="floor-shelf-fixture" x="5" y="5" width="22" height="40" rx="1" />
                <text x="16" y="27" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">LOADING BAY</text>

                <rect className="floor-shelf-fixture" x="5" y="65" width="22" height="78" rx="1" />
                <text x="16" y="105" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">STOCK CAGE</text>

                {/* B. Organic Wet Wall (Top Center) */}
                <rect className="floor-shelf-fixture" x="35" y="5" width="60" height="10" rx="1" />
                <text x="65" y="11" className="blueprint-text-secondary" fontSize="3" textAnchor="middle">ORGANIC MICRO-GREENS WALL</text>

                {/* C. Secure Vault (Top Right) */}
                <rect className="floor-shelf-fixture" x="155" y="5" width="40" height="35" rx="1" />
                <text x="175" y="23" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">SECURE VAULT</text>

                {/* D. Showcase Rings (Center Floor) */}
                {/* Ring A (Appliance Ring) */}
                <circle cx="100" cy="75" r="20" className="floor-shelf-fixture" fill="none" strokeWidth="3" stroke="var(--blueprint-grid-stroke)" />
                <text x="100" y="76" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">SHOWCASE A</text>

                {/* Ring B (Home Demo) */}
                <circle cx="60" cy="75" r="12" className="floor-shelf-fixture" fill="none" strokeWidth="2" stroke="var(--blueprint-grid-stroke)" />
                <text x="60" y="76" className="blueprint-text-muted" fontSize="2.5" textAnchor="middle">SHOWCASE B</text>

                {/* Ring C (Digital Demo) */}
                <circle cx="140" cy="75" r="12" className="floor-shelf-fixture" fill="none" strokeWidth="2" stroke="var(--blueprint-grid-stroke)" />
                <text x="140" y="76" className="blueprint-text-muted" fontSize="2.5" textAnchor="middle">SHOWCASE C</text>

                {/* E. Lounge & Coffee (Bottom Right) */}
                <rect className="floor-shelf-fixture" x="145" y="110" width="50" height="33" rx="2" />
                <text x="170" y="128" className="blueprint-text-secondary" fontSize="3" textAnchor="middle">COFFEE BAR & LOUNGE</text>

                {/* F. Checkout Front (Bottom Center) */}
                <rect className="floor-shelf-fixture" x="90" y="120" width="30" height="12" rx="1" />
                <text x="105" y="127" className="blueprint-text-primary" fontSize="3" textAnchor="middle">CHECKOUT COUNTER</text>
              </>
            )}

            {layout === 'racetrack' && (
              <>
                {/* A. Central Atrium Experience Center */}
                <rect className="floor-shelf-fixture" x="50" y="40" width="100" height="65" rx="2" fill="none" strokeWidth="1.5" strokeDasharray="3,3" />
                <text x="100" y="74" className="blueprint-text-secondary" fontSize="3.8" textAnchor="middle">ATRIUM EXPERIENCE CENTER</text>

                {/* B. Cargo Bay (Top Right) */}
                <rect className="floor-shelf-fixture" x="165" y="5" width="30" height="25" rx="1" />
                <text x="180" y="19" className="blueprint-text-secondary" fontSize="3" textAnchor="middle">INTAKE BAY</text>

                {/* C. Warehouse Staging C (Right Center) */}
                <rect className="floor-shelf-fixture" x="165" y="35" width="30" height="55" rx="1" />
                <text x="180" y="65" className="blueprint-text-secondary" fontSize="3" textAnchor="middle" transform="rotate(-90 180 65)">STAGING AREA C</text>

                {/* D. Secure Sub Vault (Bottom Left) */}
                <rect className="floor-shelf-fixture" x="5" y="110" width="40" height="35" rx="1" />
                <text x="25" y="129" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">SUB-LEVEL VAULT</text>

                {/* E. Register Gallery (Bottom Right) */}
                <rect className="floor-shelf-fixture" x="120" y="115" width="40" height="28" rx="1" />
                <text x="140" y="131" className="blueprint-text-primary" fontSize="3" textAnchor="middle">REGISTER GALLERY</text>

                {/* F. Fresh Market (Top Left / Top Center) */}
                <rect className="floor-shelf-fixture" x="5" y="5" width="55" height="15" rx="1" />
                <text x="32.5" y="14" className="blueprint-text-secondary" fontSize="3" textAnchor="middle">FRESH CANOPY</text>

                <rect className="floor-shelf-fixture" x="65" y="5" width="90" height="15" rx="1" />
                <text x="110" y="14" className="blueprint-text-secondary" fontSize="3" textAnchor="middle">PERISHABLES MARKET</text>

                {/* G. Flanking Aisles */}
                <rect className="floor-shelf-fixture" x="20" y="30" width="10" height="70" rx="1" />
                <text x="25" y="65" className="blueprint-text-muted" fontSize="3" textAnchor="middle" transform="rotate(-90 25 65)">AISLE A</text>

                <rect className="floor-shelf-fixture" x="140" y="30" width="10" height="70" rx="1" />
                <text x="145" y="65" className="blueprint-text-muted" fontSize="3" textAnchor="middle" transform="rotate(-90 145 65)">AISLE B</text>

                {/* H. Promenade Tracks (Racetrack) */}
                <rect x="10" y="24" width="150" height="2" fill="var(--blueprint-grid-stroke)" opacity="0.4" />
                <rect x="40" y="107" width="115" height="2" fill="var(--blueprint-grid-stroke)" opacity="0.4" />
              </>
            )}

            {layout === 'linear' && (
              <>
                {/* A. Backrooms & Cargo Receiving (Left Edge) */}
                <rect className="floor-shelf-fixture" x="5" y="5" width="22" height="15" rx="1" />
                <text x="16" y="14" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">DOCK A</text>

                <rect className="floor-shelf-fixture" x="5" y="23" width="22" height="15" rx="1" />
                <text x="16" y="32" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">DOCK B</text>

                <rect className="floor-shelf-fixture" x="5" y="41" width="22" height="24" rx="1" />
                <text x="16" y="54" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">STORAGE CAGE B</text>

                <rect className="floor-shelf-fixture" x="5" y="68" width="22" height="28" rx="1" />
                <text x="16" y="83" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">WALK-IN COOLER</text>

                <rect className="floor-shelf-fixture" x="5" y="99" width="22" height="46" rx="1" />
                <text x="16" y="123" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">PHARMACY WING</text>

                {/* B. Wet Wall / Prepared Food Perimeter (Top Edge) */}
                <rect className="floor-shelf-fixture" x="30" y="5" width="87" height="8" rx="1" />
                <text x="73.5" y="10.5" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">PRODUCE PERIMETER WET WALL</text>

                <rect className="floor-shelf-fixture" x="120" y="5" width="50" height="8" rx="1" />
                <text x="145" y="10.5" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">HOT FOOD DELI DEPOT</text>

                <rect className="floor-shelf-fixture" x="173" y="5" width="22" height="15" rx="1" />
                <text x="184" y="14" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">BAKERY OVENS</text>

                {/* C. Vertical Aisles Blocks (Center Floor) */}
                {/* Aisle 7 */}
                <rect className="floor-shelf-fixture" x="30" y="22" width="12" height="40" rx="1" />
                <text x="36" y="44" className="blueprint-text-muted" fontSize="3.5" textAnchor="middle" transform="rotate(-90 36 44)">AISLE 7A</text>
                
                <rect className="floor-shelf-fixture" x="30" y="82" width="12" height="40" rx="1" />
                <text x="36" y="104" className="blueprint-text-muted" fontSize="3.5" textAnchor="middle" transform="rotate(-90 36 104)">AISLE 7B</text>

                <rect className="floor-shelf-fixture" x="30" y="16" width="12" height="4" rx="0.5" fill="var(--priority-high)" opacity="0.25" />
                <text x="36" y="19.2" className="blueprint-text-primary" fontSize="2.5" textAnchor="middle">E1</text>
                <rect className="floor-shelf-fixture" x="30" y="124" width="12" height="4" rx="0.5" fill="var(--priority-high)" opacity="0.25" />
                <text x="36" y="127.2" className="blueprint-text-primary" fontSize="2.5" textAnchor="middle">E2</text>

                {/* Aisle 8 */}
                <rect className="floor-shelf-fixture" x="55" y="22" width="12" height="40" rx="1" />
                <text x="61" y="44" className="blueprint-text-muted" fontSize="3.5" textAnchor="middle" transform="rotate(-90 61 44)">AISLE 8A</text>

                <rect className="floor-shelf-fixture" x="55" y="82" width="12" height="40" rx="1" />
                <text x="61" y="104" className="blueprint-text-muted" fontSize="3.5" textAnchor="middle" transform="rotate(-90 61 104)">AISLE 8B</text>

                <rect className="floor-shelf-fixture" x="55" y="16" width="12" height="4" rx="0.5" fill="var(--priority-high)" opacity="0.25" />
                <text x="61" y="19.2" className="blueprint-text-primary" fontSize="2.5" textAnchor="middle">E3</text>
                <rect className="floor-shelf-fixture" x="55" y="124" width="12" height="4" rx="0.5" fill="var(--priority-high)" opacity="0.25" />
                <text x="61" y="127.2" className="blueprint-text-primary" fontSize="2.5" textAnchor="middle">E4</text>

                {/* Aisle 9 */}
                <rect className="floor-shelf-fixture" x="80" y="22" width="12" height="40" rx="1" />
                <text x="86" y="44" className="blueprint-text-muted" fontSize="3.5" textAnchor="middle" transform="rotate(-90 86 44)">AISLE 9A</text>

                <rect className="floor-shelf-fixture" x="80" y="82" width="12" height="40" rx="1" />
                <text x="86" y="104" className="blueprint-text-muted" fontSize="3.5" textAnchor="middle" transform="rotate(-90 86 104)">AISLE 9B</text>

                <rect className="floor-shelf-fixture" x="80" y="16" width="12" height="4" rx="0.5" fill="var(--priority-high)" opacity="0.25" />
                <text x="86" y="19.2" className="blueprint-text-primary" fontSize="2.5" textAnchor="middle">E5</text>
                <rect className="floor-shelf-fixture" x="80" y="124" width="12" height="4" rx="0.5" fill="var(--priority-high)" opacity="0.25" />
                <text x="86" y="127.2" className="blueprint-text-primary" fontSize="2.5" textAnchor="middle">E6</text>

                {/* Aisle 10 */}
                <rect className="floor-shelf-fixture" x="105" y="22" width="12" height="40" rx="1" />
                <text x="111" y="44" className="blueprint-text-muted" fontSize="3.5" textAnchor="middle" transform="rotate(-90 111 44)">AISLE 10A</text>

                <rect className="floor-shelf-fixture" x="105" y="82" width="12" height="40" rx="1" />
                <text x="111" y="104" className="blueprint-text-muted" fontSize="3.5" textAnchor="middle" transform="rotate(-90 111 104)">AISLE 10B</text>

                <rect className="floor-shelf-fixture" x="105" y="16" width="12" height="4" rx="0.5" fill="var(--priority-high)" opacity="0.25" />
                <text x="111" y="19.2" className="blueprint-text-primary" fontSize="2.5" textAnchor="middle">E7</text>
                <rect className="floor-shelf-fixture" x="105" y="124" width="12" height="4" rx="0.5" fill="var(--priority-high)" opacity="0.25" />
                <text x="111" y="127.2" className="blueprint-text-primary" fontSize="2.5" textAnchor="middle">E8</text>

                {/* D. Registers (Right Center) */}
                <rect className="floor-shelf-fixture" x="148" y="25" width="4" height="12" rx="0.5" />
                <text x="150" y="32.5" className="blueprint-text-primary" fontSize="2.8" textAnchor="middle">R1</text>

                <rect className="floor-shelf-fixture" x="148" y="45" width="4" height="12" rx="0.5" />
                <text x="150" y="52.5" className="blueprint-text-primary" fontSize="2.8" textAnchor="middle">R2</text>

                <rect className="floor-shelf-fixture" x="148" y="65" width="4" height="12" rx="0.5" />
                <text x="150" y="72.5" className="blueprint-text-primary" fontSize="2.8" textAnchor="middle">R3</text>

                <rect className="floor-shelf-fixture" x="148" y="85" width="4" height="12" rx="0.5" />
                <text x="150" y="92.5" className="blueprint-text-primary" fontSize="2.8" textAnchor="middle">R4</text>

                {/* Help Desk & Cash Vault */}
                <rect className="floor-shelf-fixture" x="135" y="112" width="30" height="10" rx="1" />
                <text x="150" y="118.5" className="blueprint-text-secondary" fontSize="3" textAnchor="middle">HELP DESK</text>

                <rect className="floor-shelf-fixture" x="173" y="112" width="22" height="33" rx="1" />
                <text x="184" y="130" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle">CASH VAULT</text>
              </>
            )}

            {/* Spatial Focus Beacon Pin */}
            {focal && (
              <circle className="target-focus-beacon" cx={focal.x} cy={focal.y} r="1.6" fill="var(--priority-critical)" />
            )}
          </g>
        </svg>

        {/* HUD controls */}
        <div className="hud-overlay">
          <button 
            onClick={() => setZoom(z => Math.min(z + 0.15, 3.5))} 
            className="hud-btn"
            title="Zoom In"
          >
            +
          </button>
          <button 
            onClick={() => setZoom(z => Math.max(z - 0.15, 0.7))} 
            className="hud-btn"
            title="Zoom Out"
          >
            -
          </button>
          <button 
            onClick={resetMapTransform} 
            className="hud-btn-reset"
            title="Recenter Map"
          >
            Reset
          </button>
          <span style={{ fontSize: '0.65rem', color: 'var(--text-muted)', marginLeft: 5, letterSpacing: '0.01em' }}>
            Drag to Pan | Scroll to Zoom
          </span>
        </div>
      </div>
    </div>
  );
};

export default Canvas;
