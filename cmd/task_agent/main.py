import os
import uvicorn
import re
import json
from typing import Optional

# Dynamic A2UI Card converter monkeypatch
import google.adk.a2a.converters.part_converter as pc
from google.genai import types as genai_types
from a2a import types as a2a_types

from a2ui_transpiler import register_monkeypatch
register_monkeypatch()


from google.adk.cli.fast_api import get_fast_api_app
from config import settings

# Resolve relative agents directory
AGENT_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "agents")

# Dynamically update agent.json URL if PUBLIC_URL is set in environment
public_url = os.environ.get("PUBLIC_URL")
if public_url:
    agent_json_path = os.path.join(AGENT_DIR, "task_agent", "agent.json")
    if os.path.exists(agent_json_path):
        try:
            with open(agent_json_path, "r", encoding="utf-8") as f:
                card_data = json.load(f)
            card_data["url"] = f"{public_url.rstrip('/')}/a2a/task_agent"
            with open(agent_json_path, "w", encoding="utf-8") as f:
                json.dump(card_data, f, indent=2)
            print(f"Dynamically updated agent.json URL to: {card_data['url']}")
        except Exception as e:
            print(f"Failed to update agent.json URL: {e}")

# Determine Session Service URI (PostgreSQL/AlloyDB)
session_service_uri = None
session_db_kwargs = {}
db_cfg = settings.get("persistence", {})
if db_cfg.get("host"):
    session_service_uri = f"postgresql+pg8000://{db_cfg.get('user')}:{db_cfg.get('password')}@{db_cfg.get('host')}:{db_cfg.get('port', '5432')}/{db_cfg.get('dbname')}"
    
    # Configure SSL for pg8000 via connect_args
    sslmode = db_cfg.get("sslmode", "disable")
    if sslmode != "disable":
        import ssl
        ssl_context = ssl.create_default_context()
        if sslmode in ("required", "require"):
            ssl_context.check_hostname = False
            ssl_context.verify_mode = ssl.CERT_NONE
        session_db_kwargs["connect_args"] = {"ssl_context": ssl_context}


# Determine Artifact Service URI (GCS Bucket)
artifact_service_uri = None
storage_cfg = settings.get("storage", {})
artifact_bucket = storage_cfg.get("artifact_bucket")
if artifact_bucket:
    artifact_service_uri = f"gs://{artifact_bucket}"

# Initialize ADK FastAPI application
print(f"Resolving AGENT_DIR: {AGENT_DIR} (exists: {os.path.exists(AGENT_DIR)})")
print(f"Configuring Session Store: {session_service_uri is not None} (SSL: {db_cfg.get('sslmode')})")
print(f"Configuring Artifact Store: {artifact_service_uri}")

app = get_fast_api_app(
    agents_dir=AGENT_DIR,
    web=True,
    a2a=True,
    session_service_uri=session_service_uri,
    session_db_kwargs=session_db_kwargs,
    artifact_service_uri=artifact_service_uri,
)

@app.get("/health")
@app.get("/readiness")
@app.get("/liveness")
@app.get("/startup")
async def health_check():
    return {"status": "healthy"}

@app.get("/api/v1/blueprint")
async def get_blueprint(layout: str = "linear", x: Optional[float] = None, y: Optional[float] = None):
    from fastapi.responses import Response
    
    svg_styles = """
    rect.fixture {
        fill: rgba(99, 102, 241, 0.04);
        stroke: rgba(99, 102, 241, 0.18);
        stroke-width: 1;
    }
    .text-primary {
        fill: rgba(255, 255, 255, 0.65);
        font-family: Outfit, sans-serif;
        font-weight: 600;
    }
    .text-secondary {
        fill: rgba(255, 255, 255, 0.45);
        font-family: Outfit, sans-serif;
        font-weight: 500;
    }
    .text-muted {
        fill: rgba(255, 255, 255, 0.25);
        font-family: Outfit, sans-serif;
    }
    .beacon {
        fill: #f87171;
    }
    """
    
    svg_content = f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 150" style="background: #0b0d19; width: 100%; height: 100%;">
      <style>{svg_styles}</style>
      <rect x="2" y="2" width="196" height="146" rx="3" fill="none" stroke="rgba(255, 255, 255, 0.03)" stroke-width="1" stroke-dasharray="4,2" />
    """
    
    if layout == "boutique":
        svg_content += """
        <rect class="fixture" x="5" y="5" width="22" height="40" rx="1" />
        <text x="16" y="27" class="text-secondary" font-size="3.2" text-anchor="middle">LOADING BAY</text>

        <rect class="fixture" x="5" y="65" width="22" height="78" rx="1" />
        <text x="16" y="105" class="text-secondary" font-size="3.2" text-anchor="middle">STOCK CAGE</text>

        <rect class="fixture" x="35" y="5" width="60" height="10" rx="1" />
        <text x="65" y="11" class="text-secondary" font-size="3" text-anchor="middle">ORGANIC MICRO-GREENS WALL</text>

        <rect class="fixture" x="155" y="5" width="40" height="35" rx="1" />
        <text x="175" y="23" class="text-secondary" font-size="3.2" text-anchor="middle">SECURE VAULT</text>

        <circle cx="100" cy="75" r="20" class="fixture" fill="none" stroke-width="3" />
        <text x="100" y="76" class="text-secondary" font-size="3.2" text-anchor="middle">SHOWCASE A</text>

        <circle cx="60" cy="75" r="12" class="fixture" fill="none" stroke-width="2" />
        <text x="60" y="76" class="text-muted" font-size="2.5" text-anchor="middle">SHOWCASE B</text>

        <circle cx="140" cy="75" r="12" class="fixture" fill="none" stroke-width="2" />
        <text x="140" y="76" class="text-muted" font-size="2.5" text-anchor="middle">SHOWCASE C</text>

        <rect class="fixture" x="145" y="110" width="50" height="33" rx="2" />
        <text x="170" y="128" class="text-secondary" font-size="3" text-anchor="middle">COFFEE BAR & LOUNGE</text>

        <rect class="fixture" x="90" y="120" width="30" height="12" rx="1" />
        <text x="105" y="127" class="text-primary" font-size="3" text-anchor="middle">CHECKOUT COUNTER</text>
        """
    elif layout == "racetrack":
        svg_content += """
        <rect class="fixture" x="50" y="40" width="100" height="65" rx="2" fill="none" stroke-width="1.5" stroke-dasharray="3,3" />
        <text x="100" y="74" class="text-secondary" font-size="3.8" text-anchor="middle">ATRIUM EXPERIENCE CENTER</text>

        <rect class="fixture" x="165" y="5" width="30" height="25" rx="1" />
        <text x="180" y="19" class="text-secondary" font-size="3" text-anchor="middle">INTAKE BAY</text>

        <rect class="fixture" x="165" y="35" width="30" height="55" rx="1" />
        <text x="180" y="65" class="text-secondary" font-size="3" text-anchor="middle" transform="rotate(-90 180 65)">STAGING AREA C</text>

        <rect class="fixture" x="5" y="110" width="40" height="35" rx="1" />
        <text x="25" y="129" class="text-secondary" font-size="3.2" text-anchor="middle">SUB-LEVEL VAULT</text>

        <rect class="fixture" x="120" y="115" width="40" height="28" rx="1" />
        <text x="140" y="131" class="text-primary" font-size="3" text-anchor="middle">REGISTER GALLERY</text>

        <rect class="fixture" x="5" y="5" width="55" height="15" rx="1" />
        <text x="32.5" y="14" class="text-secondary" font-size="3" text-anchor="middle">FRESH CANOPY</text>

        <rect class="fixture" x="65" y="5" width="90" height="15" rx="1" />
        <text x="110" y="14" class="text-secondary" font-size="3" text-anchor="middle">PERISHABLES MARKET</text>

        <rect class="fixture" x="20" y="30" width="10" height="70" rx="1" />
        <text x="25" y="65" class="text-muted" font-size="3" text-anchor="middle" transform="rotate(-90 25 65)">AISLE A</text>

        <rect class="fixture" x="140" y="30" width="10" height="70" rx="1" />
        <text x="145" y="65" class="text-muted" font-size="3" text-anchor="middle" transform="rotate(-90 145 65)">AISLE B</text>

        <rect x="10" y="24" width="150" height="2" fill="rgba(255, 255, 255, 0.03)" opacity="0.4" />
        <rect x="40" y="107" width="115" height="2" fill="rgba(255, 255, 255, 0.03)" opacity="0.4" />
        """
    else: # linear
        svg_content += """
        <rect class="fixture" x="5" y="5" width="22" height="15" rx="1" />
        <text x="16" y="14" class="text-secondary" font-size="3.2" text-anchor="middle">DOCK A</text>

        <rect class="fixture" x="5" y="23" width="22" height="15" rx="1" />
        <text x="16" y="32" class="text-secondary" font-size="3.2" text-anchor="middle">DOCK B</text>

        <rect class="fixture" x="5" y="41" width="22" height="24" rx="1" />
        <text x="16" y="54" class="text-secondary" font-size="3.2" text-anchor="middle">STORAGE CAGE B</text>

        <rect class="fixture" x="5" y="68" width="22" height="28" rx="1" />
        <text x="16" y="83" class="text-secondary" font-size="3.2" text-anchor="middle">WALK-IN COOLER</text>

        <rect class="fixture" x="5" y="99" width="22" height="46" rx="1" />
        <text x="16" y="123" class="text-secondary" font-size="3.2" text-anchor="middle">PHARMACY WING</text>

        <rect class="fixture" x="30" y="5" width="87" height="8" rx="1" />
        <text x="73.5" y="10.5" class="text-secondary" font-size="3.2" text-anchor="middle">PRODUCE PERIMETER WET WALL</text>

        <rect class="fixture" x="120" y="5" width="50" height="8" rx="1" />
        <text x="145" y="10.5" class="text-secondary" font-size="3.2" text-anchor="middle">HOT FOOD DELI DEPOT</text>

        <rect class="fixture" x="173" y="5" width="22" height="15" rx="1" />
        <text x="184" y="14" class="text-secondary" font-size="3.2" text-anchor="middle">BAKERY OVENS</text>

        <rect class="fixture" x="30" y="22" width="12" height="40" rx="1" />
        <text x="36" y="44" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 36 44)">AISLE 7A</text>
        
        <rect class="fixture" x="30" y="82" width="12" height="40" rx="1" />
        <text x="36" y="104" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 36 104)">AISLE 7B</text>

        <rect class="fixture" x="30" y="16" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="36" y="19.2" class="text-primary" font-size="2.5" text-anchor="middle">E1</text>
        <rect class="fixture" x="30" y="124" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="36" y="127.2" class="text-primary" font-size="2.5" text-anchor="middle">E2</text>

        <rect class="fixture" x="55" y="22" width="12" height="40" rx="1" />
        <text x="61" y="44" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 61 44)">AISLE 8A</text>

        <rect class="fixture" x="55" y="82" width="12" height="40" rx="1" />
        <text x="61" y="104" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 61 104)">AISLE 8B</text>

        <rect class="fixture" x="55" y="16" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="61" y="19.2" class="text-primary" font-size="2.5" text-anchor="middle">E3</text>
        <rect class="fixture" x="55" y="124" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="61" y="127.2" class="text-primary" font-size="2.5" text-anchor="middle">E4</text>

        <rect class="fixture" x="80" y="22" width="12" height="40" rx="1" />
        <text x="86" y="44" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 86 44)">AISLE 9A</text>

        <rect class="fixture" x="80" y="82" width="12" height="40" rx="1" />
        <text x="86" y="104" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 86 104)">AISLE 9B</text>

        <rect class="fixture" x="80" y="16" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="86" y="19.2" class="text-primary" font-size="2.5" text-anchor="middle">E5</text>
        <rect class="fixture" x="80" y="124" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="86" y="127.2" class="text-primary" font-size="2.5" text-anchor="middle">E6</text>

        <rect class="fixture" x="105" y="22" width="12" height="40" rx="1" />
        <text x="111" y="44" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 111 44)">AISLE 10A</text>

        <rect class="fixture" x="105" y="82" width="12" height="40" rx="1" />
        <text x="111" y="104" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 111 104)">AISLE 10B</text>

        <rect class="fixture" x="105" y="16" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="111" y="19.2" class="text-primary" font-size="2.5" text-anchor="middle">E7</text>
        <rect class="fixture" x="105" y="124" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="111" y="127.2" class="text-primary" font-size="2.5" text-anchor="middle">E8</text>

        <rect class="fixture" x="148" y="25" width="4" height="12" rx="0.5" />
        <text x="150" y="32.5" class="text-primary" font-size="2.8" text-anchor="middle">R1</text>

        <rect class="fixture" x="148" y="45" width="4" height="12" rx="0.5" />
        <text x="150" y="52.5" class="text-primary" font-size="2.8" text-anchor="middle">R2</text>

        <rect class="fixture" x="148" y="65" width="4" height="12" rx="0.5" />
        <text x="150" y="72.5" class="text-primary" font-size="2.8" text-anchor="middle">R3</text>

        <rect class="fixture" x="148" y="85" width="4" height="12" rx="0.5" />
        <text x="150" y="92.5" class="text-primary" font-size="2.8" text-anchor="middle">R4</text>

        <rect class="fixture" x="135" y="112" width="30" height="10" rx="1" />
        <text x="150" y="118.5" class="text-secondary" font-size="3" text-anchor="middle">HELP DESK</text>

        <rect class="fixture" x="173" y="112" width="22" height="33" rx="1" />
        <text x="184" y="130" class="text-secondary" font-size="3.2" text-anchor="middle">CASH VAULT</text>
        """
        
    if x is not None and y is not None:
        svg_content += f"""
        <circle class="beacon" cx="{x}" cy="{y}" r="2" />
        <circle cx="{x}" cy="{y}" r="4" fill="none" stroke="#f87171" stroke-width="0.5" opacity="0.8">
          <animate attributeName="r" values="2;7;2" dur="2s" repeatCount="indefinite" />
          <animate attributeName="opacity" values="0.8;0;0.8" dur="2s" repeatCount="indefinite" />
        </circle>
        """
        
    svg_content += "</svg>"
    return Response(content=svg_content, media_type="image/svg+xml")

if __name__ == "__main__":
    # Expose agent on port 8081 by default to prevent collision with 8080 Go API server
    server_cfg = settings.get("server", {})
    host: str = server_cfg.get("address", "0.0.0.0")
    port: int = int(server_cfg.get("port", 8081))
    
    print(f"Starting ADK Agent Service on {host}:{port}...")
    uvicorn.run(app, host=host, port=port)

