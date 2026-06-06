import React, { useState, useEffect, useRef } from 'react';
import { ApiClient } from '../../api/client';
import { useAppContext } from '../../contexts/AppContext';

interface LocationMainProps {
  onEdit: (loc: any) => void;
  onCreate: () => void;
  onError: (err: any) => void;
}

type LayoutType = 'linear' | 'boutique' | 'racetrack';

const getLayoutType = (siteId: string): LayoutType => {
  if (siteId === '44444444-4444-4444-4444-444444440001') {
    return 'boutique'; // San Francisco
  }
  if (siteId === '44444444-4444-4444-4444-444444440002') {
    return 'racetrack'; // Los Angeles
  }
  return 'linear'; // Seattle & default fallback
};

export const LocationMain: React.FC<LocationMainProps> = ({ onEdit, onCreate, onError }) => {
  const { activeSiteID } = useAppContext();
  
  // Locations & Sites Data States
  const [locations, setLocations] = useState<any[]>([]);
  const [sites, setSites] = useState<any[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedSite, setSelectedSite] = useState(activeSiteID || 'ALL');
  const [loading, setLoading] = useState(true);
  
  // Interactive Map States
  const [selectedLocationID, setSelectedLocationID] = useState<string | null>(null);
  const [zoom, setZoom] = useState<number>(1.25);
  const [pan, setPan] = useState<{ x: number; y: number }>({ x: -10, y: -10 });
  const [isDragging, setIsDragging] = useState<boolean>(false);
  const [dragStart, setDragStart] = useState<{ x: number; y: number }>({ x: 0, y: 0 });

  const layout = getLayoutType(selectedSite);

  const loadData = async () => {
    try {
      setLoading(true);
      const [fetchedLocations, fetchedSites] = await Promise.all([
        ApiClient.fetchLocations(),
        ApiClient.fetchSitesAdmin()
      ]);
      setLocations(fetchedLocations || []);
      setSites(fetchedSites || []);
      if (fetchedSites && fetchedSites.length > 0) {
        const hasActiveSite = fetchedSites.some((s) => s.ID === activeSiteID);
        setSelectedSite(hasActiveSite ? activeSiteID : fetchedSites[0].ID);
      }
    } catch (err) {
      onError(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  // Sync with top nav store context
  useEffect(() => {
    if (activeSiteID && activeSiteID !== 'ALL') {
      setSelectedSite(activeSiteID);
      setSelectedLocationID(null); // Clear highlight on site swap
    }
  }, [activeSiteID]);

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this location? All assets located here will lose their spatial grounding.')) return;
    try {
      await ApiClient.deleteLocation(id);
      if (selectedLocationID === id) setSelectedLocationID(null);
      loadData();
    } catch (err) {
      onError(err);
    }
  };

  const filteredLocations = locations.filter((l) => {
    const matchesSearch = l.Name && l.Name.toLowerCase().includes(searchTerm.toLowerCase());
    if (selectedSite === 'ALL') return matchesSearch;
    return matchesSearch && l.SiteID === selectedSite;
  });

  const getSiteName = (siteId: string) => {
    const s = sites.find((x) => x.ID === siteId);
    return s ? s.Name : siteId;
  };

  const activeLoc = locations.find(x => x.ID === selectedLocationID);

  // Resolve focal coordinates mapped to the SVG blueprint dimensions
  const getFocalCoordinates = (loc: any) => {
    if (!loc || !loc.Name) return null;
    const name = loc.Name.toLowerCase();
    
    if (layout === 'boutique') {
      if (name.includes('loading')) return { x: 16, y: 25 };
      if (name.includes('stock')) return { x: 16, y: 105 };
      if (name.includes('greens') || name.includes('wall')) return { x: 65, y: 10 };
      if (name.includes('vault')) return { x: 175, y: 23 };
      if (name.includes('showcase a')) return { x: 100, y: 75 };
      if (name.includes('showcase b')) return { x: 60, y: 75 };
      if (name.includes('showcase c')) return { x: 140, y: 75 };
      if (name.includes('lounge') || name.includes('coffee')) return { x: 170, y: 126 };
      if (name.includes('checkout') || name.includes('counter') || name.includes('register')) return { x: 105, y: 126 };
    }
    
    if (layout === 'racetrack') {
      if (name.includes('atrium')) return { x: 100, y: 72 };
      if (name.includes('intake') || name.includes('dock')) return { x: 180, y: 17 };
      if (name.includes('staging')) return { x: 180, y: 62 };
      if (name.includes('vault')) return { x: 25, y: 127 };
      if (name.includes('register') || name.includes('gallery')) return { x: 140, y: 129 };
      if (name.includes('canopy') || name.includes('fresh')) return { x: 32.5, y: 12 };
      if (name.includes('perishables') || name.includes('market')) return { x: 110, y: 12 };
      if (name.includes('aisle a') || name.includes('aisle 1')) return { x: 25, y: 65 };
      if (name.includes('aisle b') || name.includes('aisle 2')) return { x: 145, y: 65 };
    }
    
    if (layout === 'linear') {
      if (name.includes('dock a')) return { x: 16, y: 12 };
      if (name.includes('dock b')) return { x: 16, y: 30 };
      if (name.includes('stock room') || name.includes('backroom')) return { x: 16, y: 98 };
      if (name.includes('vault')) return { x: 180, y: 22 };
      if (name.includes('produce') || name.includes('wet wall')) return { x: 95, y: 12 };
      if (name.includes('deli') || name.includes('hot food')) return { x: 95, y: 34 };
      if (name.includes('bakery')) return { x: 180, y: 55 };
      if (name.includes('cooler')) return { x: 180, y: 108 };
      if (name.includes('pharmacy')) return { x: 146, y: 82.5 };
      
      if (name.includes('aisle 7')) {
        if (name.includes('front') || name.includes('endcap')) return { x: 46, y: 35 };
        if (name.includes('back') || name.includes('endcap')) return { x: 46, y: 122 };
        return { x: 46, y: 82.5 };
      }
      if (name.includes('aisle 8')) {
        if (name.includes('front') || name.includes('endcap')) return { x: 62, y: 35 };
        if (name.includes('back') || name.includes('endcap')) return { x: 62, y: 122 };
        return { x: 62, y: 82.5 };
      }
      if (name.includes('aisle 9')) {
        if (name.includes('front') || name.includes('endcap')) return { x: 78, y: 35 };
        if (name.includes('back') || name.includes('endcap')) return { x: 78, y: 122 };
        return { x: 78, y: 82.5 };
      }
      if (name.includes('aisle 10')) {
        if (name.includes('front') || name.includes('endcap')) return { x: 94, y: 35 };
        if (name.includes('back') || name.includes('endcap')) return { x: 94, y: 122 };
        return { x: 94, y: 82.5 };
      }
      if (name.includes('register 1')) return { x: 114, y: 62.5 };
      if (name.includes('register 2')) return { x: 114, y: 94.5 };
      if (name.includes('register 3')) return { x: 130, y: 62.5 };
      if (name.includes('register 4')) return { x: 130, y: 94.5 };
    }
    
    if (typeof loc.X === 'number' && typeof loc.Y === 'number') {
      return { x: loc.X, y: loc.Y };
    }
    return null;
  };

  const focalCoords = getFocalCoordinates(activeLoc);

  // SVG Interactive Map Handlers (Zoom & Drag)
  // SVG Interactive Map Handlers (Zoom & Drag)
  const selectLocationByName = (nameQuery: string) => {
    const found = locations.find(l => 
      l.SiteID === selectedSite && 
      l.Name && 
      l.Name.toLowerCase().includes(nameQuery.toLowerCase())
    );
    if (found) {
      setSelectedLocationID(found.ID);
      const rowEl = document.getElementById(`loc-row-${found.ID}`);
      if (rowEl) {
        rowEl.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
      }
    }
  };

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
    const zoomFactor = 1.1;
    const nextZoom = e.deltaY < 0 ? zoom * zoomFactor : zoom / zoomFactor;
    setZoom(Math.max(0.5, Math.min(4, nextZoom)));
  };

  const handleZoomIn = () => setZoom(prev => Math.min(4, prev * 1.2));
  const handleZoomOut = () => setZoom(prev => Math.max(0.5, prev / 1.2));
  const handleReset = () => {
    setZoom(1.25);
    setPan({ x: -10, y: -10 });
  };

  return (
    <div className="flex gap-4 flex-1 min-h-0 w-full" style={{ height: '100%' }}>
      <style>{`
        .interactive-fixture {
          cursor: pointer;
          transition: all 0.2s ease-in-out;
        }
        .interactive-fixture:hover {
          stroke: var(--accent-primary) !important;
          fill: rgba(99, 102, 241, 0.15) !important;
        }
        .active-row {
          background: rgba(99, 102, 241, 0.08) !important;
          border-left: 3px solid var(--accent-primary) !important;
        }
      `}</style>
      
      {/* --- COLUMN 1: STORE DIGITAL TWIN BLUEPRINT --- */}
      <div className="panel-card flex flex-col min-h-0" style={{ width: '40%', flexShrink: 0 }}>
        <div className="panel-header" style={{ borderBottom: '1px solid var(--panel-border)', paddingBottom: '16px' }}>
          <div className="flex justify-between items-center w-full">
            <div>
              <h2 className="panel-title">Store Digital Twin</h2>
              <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                Interactive spatial coordinates and layout.
              </p>
            </div>
            {activeLoc && (
              <span className="site-meta-pill text-xs font-bold" style={{ background: 'var(--accent-primary-glow)', color: 'var(--accent-primary)' }}>
                Focus Locked
              </span>
            )}
          </div>
        </div>

        {/* Spatial Blueprint Container */}
        <div className="flex-1 min-h-0 p-4 flex flex-col gap-3" style={{ background: 'rgba(0,0,0,0.1)' }}>
          <div className="blueprint-container" style={{ position: 'relative', overflow: 'hidden', borderRadius: 12, border: '1px solid var(--panel-border)', background: 'var(--blueprint-bg)', flex: 1, minHeight: '320px' }}>
            
            {/* SVG Canvas */}
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
              {/* Transform Group */}
              <g transform={`translate(${pan.x}, ${pan.y}) scale(${zoom})`}>
                
                {/* Outer bounds grid layout */}
                <rect x="2" y="2" width="196" height="146" rx="3" fill="none" stroke="var(--blueprint-grid-stroke)" strokeWidth="1" strokeDasharray="4,2" />

                {/* --- 1. SF BOUTIQUE LAYOUT --- */}
                {layout === 'boutique' && (
                  <>
                    <rect className="floor-shelf-fixture interactive-fixture" x="5" y="5" width="22" height="40" rx="1" onClick={() => selectLocationByName('Loading Bay')} />
                    <text x="16" y="27" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle" style={{ pointerEvents: 'none' }}>LOADING BAY</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="5" y="65" width="22" height="78" rx="1" onClick={() => selectLocationByName('Stock Cage')} />
                    <text x="16" y="105" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle" style={{ pointerEvents: 'none' }}>STOCK CAGE</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="35" y="5" width="60" height="10" rx="1" onClick={() => selectLocationByName('Greens')} />
                    <text x="65" y="11" className="blueprint-text-secondary" fontSize="3" textAnchor="middle" style={{ pointerEvents: 'none' }}>ORGANIC MICRO-GREENS WALL</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="155" y="5" width="40" height="35" rx="1" onClick={() => selectLocationByName('Vault')} />
                    <text x="175" y="23" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle" style={{ pointerEvents: 'none' }}>SECURE VAULT</text>

                    <circle cx="100" cy="75" r="20" className="floor-shelf-fixture interactive-fixture" fill="none" strokeWidth="3" stroke="var(--blueprint-grid-stroke)" onClick={() => selectLocationByName('Showcase A')} />
                    <text x="100" y="76" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle" style={{ pointerEvents: 'none' }}>SHOWCASE A</text>

                    <circle cx="60" cy="75" r="12" className="floor-shelf-fixture interactive-fixture" fill="none" strokeWidth="2" stroke="var(--blueprint-grid-stroke)" onClick={() => selectLocationByName('Showcase B')} />
                    <text x="60" y="76" className="blueprint-text-muted" fontSize="2.5" textAnchor="middle" style={{ pointerEvents: 'none' }}>SHOWCASE B</text>

                    <circle cx="140" cy="75" r="12" className="floor-shelf-fixture interactive-fixture" fill="none" strokeWidth="2" stroke="var(--blueprint-grid-stroke)" onClick={() => selectLocationByName('Showcase C')} />
                    <text x="140" y="76" className="blueprint-text-muted" fontSize="2.5" textAnchor="middle" style={{ pointerEvents: 'none' }}>SHOWCASE C</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="145" y="110" width="50" height="33" rx="2" onClick={() => selectLocationByName('Lounge')} />
                    <text x="170" y="128" className="blueprint-text-secondary" fontSize="3" textAnchor="middle" style={{ pointerEvents: 'none' }}>COFFEE BAR & LOUNGE</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="90" y="120" width="30" height="12" rx="1" onClick={() => selectLocationByName('Checkout')} />
                    <text x="105" y="127" className="blueprint-text-primary" fontSize="3" textAnchor="middle" style={{ pointerEvents: 'none' }}>CHECKOUT COUNTER</text>
                  </>
                )}

                {/* --- 2. LA RACETRACK LAYOUT --- */}
                {layout === 'racetrack' && (
                  <>
                    <rect className="floor-shelf-fixture interactive-fixture" x="50" y="40" width="100" height="65" rx="2" fill="none" strokeWidth="1.5" strokeDasharray="3,3" onClick={() => selectLocationByName('Atrium')} />
                    <text x="100" y="74" className="blueprint-text-secondary" fontSize="3.8" textAnchor="middle" style={{ pointerEvents: 'none' }}>ATRIUM EXPERIENCE CENTER</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="165" y="5" width="30" height="25" rx="1" onClick={() => selectLocationByName('Intake')} />
                    <text x="180" y="19" className="blueprint-text-secondary" fontSize="3" textAnchor="middle" style={{ pointerEvents: 'none' }}>INTAKE BAY</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="165" y="35" width="30" height="55" rx="1" onClick={() => selectLocationByName('Staging')} />
                    <text x="180" y="65" className="blueprint-text-secondary" fontSize="3" textAnchor="middle" transform="rotate(-90 180 65)" style={{ pointerEvents: 'none' }}>STAGING AREA C</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="5" y="110" width="40" height="35" rx="1" onClick={() => selectLocationByName('Vault')} />
                    <text x="25" y="129" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle" style={{ pointerEvents: 'none' }}>SUB-LEVEL VAULT</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="120" y="115" width="40" height="28" rx="1" onClick={() => selectLocationByName('Register')} />
                    <text x="140" y="131" className="blueprint-text-primary" fontSize="3" textAnchor="middle" style={{ pointerEvents: 'none' }}>REGISTER GALLERY</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="5" y="5" width="55" height="15" rx="1" onClick={() => selectLocationByName('Canopy')} />
                    <text x="32.5" y="14" className="blueprint-text-secondary" fontSize="3" textAnchor="middle" style={{ pointerEvents: 'none' }}>FRESH CANOPY</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="65" y="5" width="90" height="15" rx="1" onClick={() => selectLocationByName('Perishables')} />
                    <text x="110" y="14" className="blueprint-text-secondary" fontSize="3" textAnchor="middle" style={{ pointerEvents: 'none' }}>PERISHABLES MARKET</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="20" y="30" width="10" height="70" rx="1" onClick={() => selectLocationByName('Aisle A')} />
                    <text x="25" y="65" className="blueprint-text-muted" fontSize="3" textAnchor="middle" transform="rotate(-90 25 65)" style={{ pointerEvents: 'none' }}>AISLE A</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="140" y="30" width="10" height="70" rx="1" onClick={() => selectLocationByName('Aisle B')} />
                    <text x="145" y="65" className="blueprint-text-muted" fontSize="3" textAnchor="middle" transform="rotate(-90 145 65)" style={{ pointerEvents: 'none' }}>AISLE B</text>
                  </>
                )}

                {/* --- 3. SEATTLE/DALLAS LINEAR LAYOUT --- */}
                {layout === 'linear' && (
                  <>
                    <rect className="floor-shelf-fixture interactive-fixture" x="5" y="5" width="22" height="15" rx="1" onClick={() => selectLocationByName('Dock A')} />
                    <text x="16" y="14" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle" style={{ pointerEvents: 'none' }}>DOCK A</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="5" y="23" width="22" height="15" rx="1" onClick={() => selectLocationByName('Dock B')} />
                    <text x="16" y="32" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle" style={{ pointerEvents: 'none' }}>DOCK B</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="5" y="50" width="22" height="95" rx="1" onClick={() => selectLocationByName('Stock Room')} />
                    <text x="16" y="98" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle" style={{ pointerEvents: 'none' }}>STOCK ROOM</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="165" y="5" width="30" height="35" rx="1" onClick={() => selectLocationByName('Vault')} />
                    <text x="180" y="23" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle" style={{ pointerEvents: 'none' }}>VAULT</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="40" y="5" width="110" height="15" rx="1" onClick={() => selectLocationByName('Produce')} />
                    <text x="95" y="14" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle" style={{ pointerEvents: 'none' }}>PRODUCE PERIMETER WET WALL</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="40" y="28" width="110" height="12" rx="1" onClick={() => selectLocationByName('Deli')} />
                    <text x="95" y="35" className="blueprint-text-primary" fontSize="3" textAnchor="middle" style={{ pointerEvents: 'none' }}>HOT FOOD DELI DEPOT</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="165" y="48" width="30" height="15" rx="1" onClick={() => selectLocationByName('Bakery')} />
                    <text x="180" y="57" className="blueprint-text-secondary" fontSize="3" textAnchor="middle" style={{ pointerEvents: 'none' }}>BAKERY OVENS</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="42" y="50" width="8" height="65" rx="1" onClick={() => selectLocationByName('Aisle 7')} />
                    <text x="46" y="82.5" className="blueprint-text-muted" fontSize="2.8" textAnchor="middle" transform="rotate(-90 46 82.5)" style={{ pointerEvents: 'none' }}>AISLE 7A</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="58" y="50" width="8" height="65" rx="1" onClick={() => selectLocationByName('Aisle 8')} />
                    <text x="62" y="82.5" className="blueprint-text-muted" fontSize="2.8" textAnchor="middle" transform="rotate(-90 62 82.5)" style={{ pointerEvents: 'none' }}>AISLE 8A</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="74" y="50" width="8" height="65" rx="1" onClick={() => selectLocationByName('Aisle 9')} />
                    <text x="78" y="82.5" className="blueprint-text-muted" fontSize="2.8" textAnchor="middle" transform="rotate(-90 78 82.5)" style={{ pointerEvents: 'none' }}>AISLE 9A</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="90" y="50" width="8" height="65" rx="1" onClick={() => selectLocationByName('Aisle 10')} />
                    <text x="94" y="82.5" className="blueprint-text-muted" fontSize="2.8" textAnchor="middle" transform="rotate(-90 94 82.5)" style={{ pointerEvents: 'none' }}>AISLE 10A</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="110" y="50" width="8" height="25" rx="1" onClick={() => selectLocationByName('Register 1')} />
                    <text x="114" y="62.5" className="blueprint-text-muted" fontSize="2" textAnchor="middle" transform="rotate(-90 114 62.5)" style={{ pointerEvents: 'none' }}>R1</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="110" y="82" width="8" height="25" rx="1" onClick={() => selectLocationByName('Register 2')} />
                    <text x="114" y="94.5" className="blueprint-text-muted" fontSize="2" textAnchor="middle" transform="rotate(-90 114 94.5)" style={{ pointerEvents: 'none' }}>R2</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="126" y="50" width="8" height="25" rx="1" onClick={() => selectLocationByName('Register 3')} />
                    <text x="130" y="62.5" className="blueprint-text-muted" fontSize="2" textAnchor="middle" transform="rotate(-90 130 62.5)" style={{ pointerEvents: 'none' }}>R3</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="126" y="82" width="8" height="25" rx="1" onClick={() => selectLocationByName('Register 4')} />
                    <text x="130" y="94.5" className="blueprint-text-muted" fontSize="2" textAnchor="middle" transform="rotate(-90 130 94.5)" style={{ pointerEvents: 'none' }}>R4</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="142" y="50" width="8" height="65" rx="1" onClick={() => selectLocationByName('Pharmacy')} />
                    <text x="146" y="82.5" className="blueprint-text-muted" fontSize="2.8" textAnchor="middle" transform="rotate(-90 146 82.5)" style={{ pointerEvents: 'none' }}>PHARMACY WING</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="165" y="70" width="30" height="75" rx="1" onClick={() => selectLocationByName('Cooler')} />
                    <text x="180" y="108" className="blueprint-text-secondary" fontSize="3.2" textAnchor="middle" transform="rotate(-90 180 108)" style={{ pointerEvents: 'none' }}>WALK-IN COOLER</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="42" y="122" width="8" height="23" rx="1" onClick={() => selectLocationByName('Aisle 7')} />
                    <text x="46" y="133.5" className="blueprint-text-muted" fontSize="2" textAnchor="middle" transform="rotate(-90 46 133.5)" style={{ pointerEvents: 'none' }}>AISLE 7B</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="58" y="122" width="8" height="23" rx="1" onClick={() => selectLocationByName('Aisle 8')} />
                    <text x="62" y="133.5" className="blueprint-text-muted" fontSize="2" textAnchor="middle" transform="rotate(-90 62 133.5)" style={{ pointerEvents: 'none' }}>AISLE 8B</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="74" y="122" width="8" height="23" rx="1" onClick={() => selectLocationByName('Aisle 9')} />
                    <text x="78" y="133.5" className="blueprint-text-muted" fontSize="2" textAnchor="middle" transform="rotate(-90 78 133.5)" style={{ pointerEvents: 'none' }}>AISLE 9B</text>

                    <rect className="floor-shelf-fixture interactive-fixture" x="90" y="122" width="8" height="23" rx="1" onClick={() => selectLocationByName('Aisle 10')} />
                    <text x="94" y="133.5" className="blueprint-text-muted" fontSize="2" textAnchor="middle" transform="rotate(-90 94 133.5)" style={{ pointerEvents: 'none' }}>AISLE 10B</text>
                  </>
                )}

                {/* --- 4. DYNAMIC PULSING BEACON (Highlights selected location) --- */}
                {focalCoords && (
                  <g className="spatial-beacon" style={{ pointerEvents: 'none' }}>
                    <circle cx={focalCoords.x} cy={focalCoords.y} r="7" fill="var(--priority-high)" opacity="0.4">
                      <animate attributeName="r" values="3;9;3" dur="2s" repeatCount="indefinite" />
                      <animate attributeName="opacity" values="0.6;0;0.6" dur="2s" repeatCount="indefinite" />
                    </circle>
                    <circle cx={focalCoords.x} cy={focalCoords.y} r="2.5" fill="var(--priority-high)" />
                  </g>
                )}
              </g>
            </svg>

            {/* Map HUD Controls Overlay */}
            <div style={{ position: 'absolute', bottom: 12, right: 12, display: 'flex', gap: 6 }}>
              <button className="a2ui-btn-action text-xs" style={{ padding: '6px 10px', background: 'var(--panel-glass)' }} onClick={handleZoomIn}>+</button>
              <button className="a2ui-btn-action text-xs" style={{ padding: '6px 10px', background: 'var(--panel-glass)' }} onClick={handleZoomOut}>-</button>
              <button className="a2ui-btn-action text-xs" style={{ padding: '6px 12px', background: 'var(--panel-glass)' }} onClick={handleReset}>Reset</button>
            </div>
          </div>

          {/* Focal Details HUD */}
          <div className="site-meta-pill" style={{ background: 'var(--panel-glass)', border: '1px solid var(--panel-border)', padding: '12px', borderRadius: 8 }}>
            <h5 style={{ margin: '0 0 4px 0', fontSize: '0.82rem', color: 'var(--text-primary)' }}>Spatial Metadata</h5>
            {activeLoc ? (
              <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                <div style={{ marginBottom: 4 }}><strong>Focal Zone:</strong> {activeLoc.Name}</div>
                <div style={{ marginBottom: 4 }}><strong>Type:</strong> {activeLoc.LocationType} | {activeLoc.LocationFunctionType}</div>
                <div><strong>Grounding Grid:</strong> X: {activeLoc.X?.toFixed(2)} | Y: {activeLoc.Y?.toFixed(2)} | Z: {activeLoc.Z?.toFixed(2)}</div>
              </div>
            ) : (
              <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                Click on any sub-location row in the table or click on a physical fixture in the Store Digital Twin to lock the spatial focal beacon.
              </div>
            )}
          </div>
        </div>
      </div>

      {/* --- COLUMN 2: FACILITY SUB-LOCATIONS LIST --- */}
      <div className="panel-card flex flex-col min-h-0 flex-1">
        <div className="panel-header" style={{ borderBottom: '1px solid var(--panel-border)', paddingBottom: '16px' }}>
          <div className="flex justify-between items-center w-full">
            <div>
              <h2 className="panel-title">Facility Sub-Locations</h2>
              <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                Map aisles, storage shelves, cash registers, or custom spatial polygons inside retail structures.
              </p>
            </div>
            <button className="btn-primary" onClick={onCreate}>
              + Map Location
            </button>
          </div>
        </div>

        {/* Filters HUD */}
        <div className="p-4 flex gap-4 items-center border-b" style={{ borderColor: 'var(--panel-border)', background: 'rgba(255,255,255,0.01)' }}>
          <div className="flex-1 relative">
            <input
              type="text"
              className="site-meta-pill w-full"
              style={{ borderRadius: '8px', padding: '8px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)' }}
              placeholder="Search sub-locations by name..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          
          {/* Site Filter Dropdown */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)', fontWeight: 600 }}>Site Context:</span>
            <select
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '6px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              value={selectedSite}
              onChange={(e) => {
                setSelectedSite(e.target.value);
                setSelectedLocationID(null); // Clear focal beacon
              }}
            >
              <option value="ALL" style={{ background: 'var(--bg-main)' }}>All Sites</option>
              {sites.map((s) => (
                <option key={s.ID} value={s.ID} style={{ background: 'var(--bg-main)' }}>{s.Name}</option>
              ))}
            </select>
          </div>
        </div>

        {/* Table Area */}
        <div className="panel-body-scrollable flex-1">
          {loading ? (
            <div className="flex justify-center items-center h-32 text-muted">Loading locations database...</div>
          ) : filteredLocations.length === 0 ? (
            <div className="flex justify-center items-center h-32 text-muted">No sub-locations found.</div>
          ) : (
            <table className="a2ui-table">
              <thead>
                <tr style={{ borderBottom: '2px solid var(--panel-border)' }}>
                  <th className="a2ui-label" style={{ padding: '12px' }}>Location Name</th>
                  <th className="a2ui-label" style={{ padding: '12px' }}>Site Context</th>
                  <th className="a2ui-label" style={{ padding: '12px' }}>Spatial Vectors (X, Y, Z)</th>
                  <th className="a2ui-label" style={{ padding: '12px' }}>Class / Function</th>
                  <th className="a2ui-label" style={{ padding: '12px', textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredLocations.map((l) => (
                  <tr 
                    id={`loc-row-${l.ID}`}
                    key={l.ID} 
                    className={`hover:bg-white/5 transition-colors cursor-pointer ${selectedLocationID === l.ID ? 'active-row' : ''}`}
                    onClick={() => setSelectedLocationID(l.ID)}
                    style={selectedLocationID === l.ID ? { background: 'rgba(99, 102, 241, 0.08)' } : {}}
                  >
                    <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px' }}>
                      <span className="font-semibold" style={{ color: selectedLocationID === l.ID ? 'var(--accent-primary)' : 'var(--text-primary)' }}>
                        {l.Name}
                      </span>
                    </td>
                    <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px', color: 'var(--text-secondary)' }}>
                      {getSiteName(l.SiteID)}
                    </td>
                    <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px', fontFamily: 'var(--font-mono)', fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                      X: {l.X?.toFixed(2)}, Y: {l.Y?.toFixed(2)}, Z: {l.Z?.toFixed(2)}
                    </td>
                    <td className="a2ui-value" style={{ textAlign: 'left', padding: '12px' }}>
                      <div className="flex gap-1 text-xs">
                        <span className="site-meta-pill font-semibold uppercase" style={{ background: 'rgba(255,255,255,0.03)' }}>
                          {l.LocationType}
                        </span>
                        <span className="site-meta-pill font-semibold uppercase" style={{ background: 'var(--accent-primary-glow)', borderColor: 'var(--panel-border)' }}>
                          {l.LocationFunctionType}
                        </span>
                      </div>
                    </td>
                    <td className="a2ui-value" style={{ padding: '12px' }}>
                      <div className="flex justify-end items-center gap-2" onClick={(e) => e.stopPropagation()}>
                        <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px' }} onClick={() => onEdit(l)}>
                          Edit
                        </button>
                        <button className="a2ui-btn-action text-xs" style={{ padding: '4px 8px', borderColor: 'var(--priority-critical)', color: 'var(--priority-critical)' }} onClick={() => handleDelete(l.ID)}>
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

    </div>
  );
};

