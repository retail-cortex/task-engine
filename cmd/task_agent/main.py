import sys
# Self-healing namespace merge to bypass setuptools nspkg.pth hijacking
if 'google' in sys.modules:
    m = sys.modules['google']
    adk_path = '/Users/rmcguinness/Projects/google/adk-python/src/google'
    if hasattr(m, '__path__') and adk_path not in m.__path__:
        m.__path__.append(adk_path)

import os
# Force immediate local credential resolution and native DNS resolution
# to prevent 30-second cold-start metadata server timeouts in local offline contexts.
# In deployed Cloud Run environments (dev/prod), we must use the real metadata server.
if os.environ.get("MODENV_RUNTIME", "") in ["", "local"]:
    os.environ["GCE_METADATA_HOST"] = "127.0.0.1:9999"
os.environ["GRPC_DNS_RESOLVER"] = "native"

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

# A2UI cards are fully cached using builtins and tokens, leaving the history naturally clean and token-efficient.


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

# Determine Session Service URI (Local In-Memory for extreme speed, or PostgreSQL/AlloyDB if explicitly set)
session_service_uri = os.environ.get("SESSION_SERVICE_URI")
session_db_kwargs = {
    "pool_pre_ping": True,
    "pool_recycle": 1800,
    "pool_size": 10,
    "max_overflow": 20
}
db_cfg = settings.get("persistence", {})
if not session_service_uri and os.environ.get("FORCE_DB_SESSION") == "true" and db_cfg.get("host"):
    session_service_uri = f"postgresql+pg8000://{db_cfg.get('user')}:{db_cfg.get('password')}@{db_cfg.get('host')}:{db_cfg.get('port', '5432')}/{db_cfg.get('dbname')}"
    
    # Configure SSL for pg8000 via connect_args
    sslmode = db_cfg.get("sslmode", "disable")
    connect_args = {}
    if sslmode != "disable":
        import ssl
        ssl_context = ssl.create_default_context()
        if sslmode in ("required", "require"):
            ssl_context.check_hostname = False
            ssl_context.verify_mode = ssl.CERT_NONE
        connect_args["ssl_context"] = ssl_context
    session_db_kwargs["connect_args"] = connect_args

# Determine Artifact Service URI (Local In-Memory for extreme speed, or GCS if explicitly set)
artifact_service_uri = os.environ.get("ARTIFACT_SERVICE_URI")
if not artifact_service_uri:
    storage_cfg = settings.get("storage", {})
    artifact_bucket = storage_cfg.get("artifact_bucket")
    # Only use GCS if FORCE_GCS_ARTIFACTS is explicitly enabled
    if os.environ.get("FORCE_GCS_ARTIFACTS") == "true" and artifact_bucket:
        artifact_service_uri = f"gs://{artifact_bucket}"
    else:
        # Default to None (In-Memory) for local development speed
        artifact_service_uri = None

# Initialize OpenTelemetry Tracing targeting Google Cloud Trace
def init_telemetry():
    import os
    import sys
    # Aggressive flush config for ADK's Cloud Trace processor
    os.environ["OTEL_BSP_SCHEDULE_DELAY"] = "1000"       # Delay in milliseconds
    os.environ["OTEL_BSP_MAX_EXPORT_BATCH_SIZE"] = "5"   # Batch size

    from opentelemetry import trace
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor
    from opentelemetry.exporter.cloud_trace import CloudTraceSpanExporter
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
    from opentelemetry.propagate import set_global_textmap

    # Ensure W3C Trace Context is the global propagator
    set_global_textmap(TraceContextTextMapPropagator())

    # Create resource definition
    resource = Resource.create({
        "service.name": "gemini-task-agent",
        "service.version": "1.0.0"
    })

    # Initialize Google Cloud Trace exporter
    # It automatically resolves credentials and project ID using ADC
    try:
        exporter = CloudTraceSpanExporter()
        span_processor = BatchSpanProcessor(exporter)
        
        provider = TracerProvider(resource=resource)
        provider.add_span_processor(span_processor)
        trace.set_tracer_provider(provider)
        print("OpenTelemetry Tracing successfully initialized with Google Cloud Trace exporter.")
    except Exception as e:
        print(f"[Warning] Failed to initialize Google Cloud Trace exporter: {e}. Tracing will be disabled.", file=sys.stderr)

# Initialize ADK FastAPI application
print(f"Resolving AGENT_DIR: {AGENT_DIR} (exists: {os.path.exists(AGENT_DIR)})")
print(f"Configuring Session Store: {session_service_uri is not None} (SSL: {db_cfg.get('sslmode')})")
print(f"Configuring Artifact Store: {artifact_service_uri}")

try:
    app = get_fast_api_app(
        agents_dir=AGENT_DIR,
        web=True,
        a2a=True,
        session_service_uri=session_service_uri,
        session_db_kwargs=session_db_kwargs,
        artifact_service_uri=artifact_service_uri,
        allow_origins=["*"],
    )
except Exception as e:
    import sys
    print(f"[Warning] Database connectivity failed for ADK Session Store: {e}", file=sys.stderr)
    print("[Warning] Bootstrapping Python Agent in degraded offline mode (falling back to in-memory session store)...", file=sys.stderr)
    app = get_fast_api_app(
        agents_dir=AGENT_DIR,
        web=True,
        a2a=True,
        session_service_uri=None, # forces in-memory store!
        artifact_service_uri=artifact_service_uri,
        allow_origins=["*"],
    )

# Thread-local user ID context propagator middleware
from shared_context import active_user_id_var
from fastapi import Request

@app.middleware("http")
async def extract_user_id_middleware(request: Request, call_next):
    # Exempt health checks, documentation, and ADK internal session endpoints
    path = request.url.path
    if path in ["/healthz", "/readiness", "/docs", "/openapi.json"] or path.startswith("/apps/"):
        return await call_next(request)

    user_id = request.headers.get("X-User-ID")

    # Try to extract and validate OIDC identity from Authorization header
    auth_header = request.headers.get("Authorization")
    if not user_id and auth_header and auth_header.startswith("Bearer "):
        token = auth_header.removeprefix("Bearer ")
        import sys
        if token.startswith("ya29."):
            # Google OAuth2 Access Token: validate using Google's tokeninfo endpoint
            import urllib.request
            tokeninfo_url = f"https://oauth2.googleapis.com/tokeninfo?access_token={token}"
            try:
                req = urllib.request.Request(tokeninfo_url)
                with urllib.request.urlopen(req, timeout=3) as response:
                    payload = json.loads(response.read().decode("utf-8"))
                    email = payload.get("email")
                    if email:
                        user_id = email
                        print(f"[Middleware] Authenticated A2A Access Token User: {user_id}", file=sys.stderr)
            except Exception as token_err:
                print(f"[Middleware] Failed validating A2A Access Token via tokeninfo: {token_err}", file=sys.stderr)
        else:
            # Google OIDC ID Token (JWT): validate cryptographically using public certificates
            try:
                from google.oauth2 import id_token
                from google.auth.transport import requests as auth_requests
                payload = id_token.verify_oauth2_token(token, auth_requests.Request(), audience=None)
                email = payload.get("email")
                if email:
                    user_id = email
                    print(f"[Middleware] Authenticated A2A OIDC User: {user_id}", file=sys.stderr)
            except Exception as e:
                print(f"[Middleware] Failed validating A2A OIDC token: {e}", file=sys.stderr)

    # If header is absent and this is an A2A POST request, extract from body
    if not user_id and request.url.path == "/a2a/task_agent" and request.method == "POST":
        try:
            body_bytes = await request.body()
            # Restore request body so that subsequent route handlers can read it
            async def receive():
                return {"type": "http.request", "body": body_bytes}
            request._receive = receive

            if body_bytes:
                body_json = json.loads(body_bytes)
                import sys
                print(f"[Middleware] A2A body JSON: {body_json}", file=sys.stderr)
                if isinstance(body_json, dict):
                    params = body_json.get("params", {})
                    if isinstance(params, dict):
                        context = params.get("context", {})
                        if isinstance(context, dict):
                            # Try userEmail first, then userId, then contextId
                            user_id = context.get("userEmail") or context.get("userId") or context.get("contextId")
        except Exception as e:
            import sys
            print(f"[Middleware] Error parsing A2A body: {e}", file=sys.stderr)

    if not user_id or user_id == "00000000-0000-0000-0000-000000000000":
        from fastapi.responses import JSONResponse
        return JSONResponse(
            status_code=401,
            content={"detail": "Unauthorized: A valid, authenticated user context is required"}
        )
    token = active_user_id_var.set(user_id)
    
    # Safely extract context_id from POST request body if present to populate mapping
    if request.method == "POST" and user_id:
        try:
            body_bytes = await request.body()
            # Restore request body so that subsequent route handlers can read it
            async def receive():
                return {"type": "http.request", "body": body_bytes}
            request._receive = receive
            
            if body_bytes:
                body_json = json.loads(body_bytes)
                if isinstance(body_json, dict):
                    params = body_json.get("params", {})
                    if isinstance(params, dict):
                        message = params.get("message", {})
                        if isinstance(message, dict):
                            context_id = message.get("context_id")
                            if context_id:
                                from shared_context import session_user_map
                                session_user_map[context_id] = user_id
                                import sys
                                print(f"[Identity Map] Mapped session {context_id} -> user {user_id}", file=sys.stderr)
        except Exception as e:
            import sys
            print(f"[Identity Map] Error parsing session ID: {e}", file=sys.stderr)

    try:
        response = await call_next(request)
        return response
    finally:
        active_user_id_var.reset(token)


# Telemetry and instrumentation will be initialized in the startup event to prevent gRPC fork deadlocks

async def warm_up_vertex_ai():
    try:
        print("[Warmup] Eagerly warming up Vertex AI connection pool in the background...")
        from google import genai
        client = genai.Client()
        # Non-blocking, simple ping to resolve ADC credentials and establish TLS/gRPC pool
        await client.models.generate_content_async(
            model='gemini-2.5-flash',
            contents='ping'
        )
        print("[Warmup] Vertex AI connection pool is warm and ready!")
    except Exception as e:
        print(f"[Warmup] Vertex AI warm-up failed: {e}. This is normal if offline or during local test runs.")

@app.on_event("startup")
def startup_event():
    print("➜ FastAPI startup event: Initializing Google Cloud Trace telemetry and instrumenting app...")
    init_telemetry()
    from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
    FastAPIInstrumentor.instrument_app(app)
    
    # Eagerly warm up Vertex AI in the background
    import asyncio
    asyncio.create_task(warm_up_vertex_ai())

@app.on_event("shutdown")
def shutdown_event():
    from opentelemetry import trace
    print("Shutting down Python Agent... Flushing remaining OpenTelemetry spans.")
    provider = trace.get_tracer_provider()
    if hasattr(provider, "force_flush"):
        provider.force_flush(2000)
    if hasattr(provider, "shutdown"):
        provider.shutdown()

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

# Trigger reload: Telemetry deferred to startup event to prevent deadlocks.

