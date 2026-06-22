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
from typing import Dict, Any, List

def synthesize_a2ui_error_card(exc: Exception) -> List[Dict[str, Any]]:
    exc_type = type(exc).__name__
    exc_msg = str(exc)
    
    friendly_title = "Application Error"
    friendly_msg = "An unexpected error occurred while processing your request. Please try again."
    
    # Customize message based on known error signatures
    if "429" in exc_msg or "RESOURCE_EXHAUSTED" in exc_msg:
        friendly_title = "Rate Limit Exceeded (429)"
        friendly_msg = "The task engine is experiencing heavy traffic and has temporarily rate-limited your request. Please wait a moment and click retry."
    elif "connection" in exc_msg.lower() or "dial tcp" in exc_msg.lower() or "reset by peer" in exc_msg.lower():
        friendly_title = "Database Connection Failed"
        friendly_msg = "The task engine was unable to establish a secure database connection. The database may be restarting or undergoing maintenance."
    elif "unauthorized" in exc_msg.lower() or "401" in exc_msg or "token" in exc_msg.lower():
        friendly_title = "Authentication Required (401)"
        friendly_msg = "Your user session has expired or is invalid. Please sign out and sign back in to refresh your credentials."
        
    error_card = {
        "type": "card",
        "title": f"⚠️ {friendly_title}",
        "style": "standard",
        "children": [
            {
                "type": "text",
                "content": friendly_msg,
                "style": "primary"
            },
            {
                "type": "divider"
            },
            {
                "type": "text",
                "content": f"Technical Details ({exc_type}):",
                "style": "secondary"
            },
            {
                "type": "text",
                "content": exc_msg,
                "style": "secondary"
            }
        ]
    }
    
    from a2ui_transpiler import normalize_card_to_a2ui_messages
    parts_data = normalize_card_to_a2ui_messages(error_card)
    
    a2a_parts = []
    for part in parts_data:
        a2a_parts.append({
            "data": {
                "data": part,
                "metadata": {
                    "mimeType": "application/json+a2ui"
                }
            }
        })
        
    return a2a_parts

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
    except Exception as exc:
        if request.url.path == "/a2a/task_agent" and request.method == "POST":
            import sys
            import traceback
            print(f"[A2A Exception Handler] Caught unhandled exception: {exc}", file=sys.stderr)
            traceback.print_exc(file=sys.stderr)
            
            try:
                a2a_parts = synthesize_a2ui_error_card(exc)
                request_id = 1
                try:
                    body_bytes = await request.body()
                    if body_bytes:
                        body_json = json.loads(body_bytes)
                        request_id = body_json.get("id", 1)
                except Exception:
                    pass
                    
                from fastapi.responses import JSONResponse
                return JSONResponse(
                    status_code=200,
                    content={
                        "jsonrpc": "2.0",
                        "result": {
                            "message": {
                                "role": "agent",
                                "parts": a2a_parts
                            }
                        },
                        "id": request_id
                    }
                )
            except Exception as handler_err:
                print(f"[A2A Exception Handler] Fatal: Failed to synthesize error card: {handler_err}", file=sys.stderr)
                traceback.print_exc(file=sys.stderr)
        raise exc
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
    from a2ui_transpiler import generate_blueprint_svg
    
    svg_content = generate_blueprint_svg(layout, x, y)
    return Response(content=svg_content, media_type="image/svg+xml")

if __name__ == "__main__":
    # Expose agent on port 8081 by default to prevent collision with 8080 Go API server
    server_cfg = settings.get("server", {})
    host: str = server_cfg.get("address", "0.0.0.0")
    port: int = int(server_cfg.get("port", 8081))
    
    print(f"Starting ADK Agent Service on {host}:{port}...")
    uvicorn.run(app, host=host, port=port)

# Trigger reload: Telemetry deferred to startup event to prevent deadlocks.

