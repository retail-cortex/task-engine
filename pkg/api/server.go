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

package api

import (
	"context"
	_ "embed"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/agents"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/gorm"
)

//go:embed openapi.json
var openapiJSON []byte

//go:embed swagger.html
var swaggerHTML []byte

type OAuthConfig struct {
	ClientID         string   `toml:"client_id"`
	AllowedClientIDs []string `toml:"allowed_client_ids"`
	Secret           string   `toml:"secret"`
}

type IAPConfig struct {
	Audience string `toml:"audience"`
}

// Config holds the server binding configuration.
type Config struct {
	Port    string      `toml:"port"`
	Address string      `toml:"address"`
	OAuth   OAuthConfig `toml:"oauth"`
	IAP     IAPConfig   `toml:"iap"`
}

// Server defines the GIN web server and routing engine.
type Server struct {
	cfg                Config
	engine             *gin.Engine
	adminService       service.AdminService
	taskService        service.TaskService
	shiftService       service.ShiftService
	ragService         service.RAGService
	automationService  service.AutomationService
	schedulerService   service.SchedulerService
	adminHandler       *AdminHandler
	operationalHandler *OperationalHandler
	MCPHandler         *MCPHandler
	chatHandler        *ChatHandler
	ttsHandler         *TTSHandler
	traceProxyHandler  *TraceProxyHandler
}

// NewServer instantiates the Gin engine and mounts all route controllers.
func NewServer(
	cfg Config,
	adminService service.AdminService,
	taskService service.TaskService,
	shiftService service.ShiftService,
	ragService service.RAGService,
	automationService service.AutomationService,
	schedulerService service.SchedulerService,
	db ...*gorm.DB, // Variadic optional database connection parameters block!
) (*Server, error) {
	engine := gin.New()

	// Base middlewares
	engine.Use(otelgin.Middleware("gemini-task-api"))
	engine.Use(gin.Recovery())
	engine.Use(CORSConfigMiddleware())

	// Serve static workspace file assets relativement
	engine.Static("/static", "./web/static")

	// Instantiate the self-bootstrapping Postgres ADK SessionService
	var agentsSessionService agents.SessionService
	if len(db) > 0 && db[0] != nil {
		var errSess error
		agentsSessionService, errSess = agents.NewSessionService(db[0])
		if errSess != nil {
			log.Printf("WARNING: Failed to bootstrap ADK agent session state postgres service: %v. Falling back to in-memory session service to allow degraded operation.", errSess)
			agentsSessionService = agents.NewInMemorySessionService()
		}
	} else {
		// Mock compatible fallback memory session module (used inside sandboxes and unit tests)
		agentsSessionService = agents.NewInMemorySessionService()
	}

	var activeDB *gorm.DB
	if len(db) > 0 {
		activeDB = db[0]
	}
	adminHandler := NewAdminHandler(adminService, schedulerService, taskService, shiftService, ragService)
	operationalHandler := NewOperationalHandler(taskService, shiftService, automationService, cfg, activeDB)
	mcpHandler, err := NewMCPHandler(taskService, ragService, shiftService, automationService)
	if err != nil {
		return nil, err
	}
	chatHandler := NewChatHandler(taskService, shiftService, ragService, automationService, agentsSessionService)

	traceProxyHandler, err := NewTraceProxyHandler(context.Background())
	if err != nil {
		log.Printf("WARNING: Failed to initialize trace proxy handler: %v. Frontend traces will not be proxied.", err)
	}

	s := &Server{
		cfg:                cfg,
		engine:             engine,
		adminService:       adminService,
		taskService:        taskService,
		shiftService:       shiftService,
		ragService:         ragService,
		automationService:  automationService,
		schedulerService:   schedulerService,
		adminHandler:       adminHandler,
		operationalHandler: operationalHandler,
		MCPHandler:         mcpHandler,
		chatHandler:        chatHandler,
		ttsHandler:         NewTTSHandler(),
		traceProxyHandler:  traceProxyHandler,
	}

	s.setupRoutes()

	return s, nil
}

func (s *Server) setupRoutes() {
	// Swagger documentation endpoints
	s.engine.GET("/swagger", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerHTML)
	})
	s.engine.GET("/swagger/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", openapiJSON)
	})

	// Health and probe endpoints
	s.engine.GET("/health/readiness", s.operationalHandler.Readiness)
	s.engine.GET("/liveness", s.operationalHandler.Liveness)
	s.engine.GET("/startup", s.operationalHandler.Startup)

	// Trace Proxy endpoint (for receiving frontend traces and securely forwarding them to Google Cloud Trace)
	if s.traceProxyHandler != nil {
		s.engine.POST("/api/v1/traces", s.traceProxyHandler.ProxyTraces)
	}

	// Global static/parameterless MCP endpoint (for standard client discovery)
	s.engine.POST("/api/v1/mcp", UserContextMiddleware(s.adminService, s.cfg), s.MCPHandler.Handler())
	s.engine.GET("/api/v1/mcp", UserContextMiddleware(s.adminService, s.cfg), s.MCPHandler.Handler())

	// Secure Text-to-Speech endpoint
	s.engine.POST("/api/v1/tts", UserContextMiddleware(s.adminService, s.cfg), s.ttsHandler.Synthesize)

	// Admin endpoints
	admin := s.engine.Group("/api/v1/admin", UserContextMiddleware(s.adminService, s.cfg))
	{
		// Users
		admin.GET("/users", s.adminHandler.ListUsers)
		admin.GET("/users/:id", s.adminHandler.GetUser)
		admin.POST("/users", s.adminHandler.CreateUser)
		admin.PUT("/users/:id", s.adminHandler.UpdateUser)
		admin.DELETE("/users/:id", s.adminHandler.DeleteUser)
		admin.PUT("/users/:id/roles", s.adminHandler.AssignRole)

		// Roles
		admin.POST("/roles", s.adminHandler.CreateRole)
		admin.GET("/roles/:id", s.adminHandler.GetRole)
		admin.PUT("/roles/:id", s.adminHandler.UpdateRole)
		admin.DELETE("/roles/:id", s.adminHandler.DeleteRole)
		admin.GET("/roles", s.adminHandler.ListRoles)

		// Organizations
		admin.POST("/organizations", s.adminHandler.CreateOrganization)
		admin.GET("/organizations/:id", s.adminHandler.GetOrganization)
		admin.PUT("/organizations/:id", s.adminHandler.UpdateOrganization)
		admin.DELETE("/organizations/:id", s.adminHandler.DeleteOrganization)
		admin.GET("/organizations", s.adminHandler.ListOrganizations)
		admin.PUT("/organizations/:id/users/:userId", s.adminHandler.AssignUserToOrganization)

		// Sites
		admin.POST("/organizations/:id/sites", s.adminHandler.CreateSite)
		admin.GET("/sites/:id", s.adminHandler.GetSite)
		admin.PUT("/sites/:id", s.adminHandler.UpdateSite)
		admin.DELETE("/sites/:id", s.adminHandler.DeleteSite)
		admin.GET("/sites", s.adminHandler.ListSites)

		// Locations
		admin.POST("/organizations/:id/sites/:siteId/locations", s.adminHandler.CreateLocation)
		admin.GET("/locations/:id", s.adminHandler.GetLocation)
		admin.PUT("/locations/:id", s.adminHandler.UpdateLocation)
		admin.DELETE("/locations/:id", s.adminHandler.DeleteLocation)
		admin.GET("/locations", s.adminHandler.ListLocations)

		// Assets
		admin.POST("/organizations/:id/sites/:siteId/locations/:locationId/assets", s.adminHandler.CreateAsset)
		admin.GET("/assets/:id", s.adminHandler.GetAsset)
		admin.PUT("/assets/:id", s.adminHandler.UpdateAsset)
		admin.DELETE("/assets/:id", s.adminHandler.DeleteAsset)
		admin.GET("/assets", s.adminHandler.ListAssets)

		// Tasks
		admin.POST("/tasks/templates", s.adminHandler.CreateTaskTemplate)
		admin.GET("/tasks/templates/:id", s.adminHandler.GetTaskTemplate)
		admin.PUT("/tasks/templates/:id", s.adminHandler.UpdateTaskTemplate)
		admin.DELETE("/tasks/templates/:id", s.adminHandler.DeleteTaskTemplate)
		admin.GET("/tasks/templates", s.adminHandler.ListTaskTemplates)

		// Task Executions CRUD
		admin.GET("/tasks/executions", s.adminHandler.ListTaskExecutions)
		admin.GET("/tasks/executions/:id", s.adminHandler.GetTaskExecution)
		admin.DELETE("/tasks/executions/:id", s.adminHandler.DeleteTaskExecution)

		// Shift Agent Sessions CRUD
		admin.GET("/shifts/sessions", s.adminHandler.ListShiftSessions)
		admin.GET("/shifts/sessions/:id", s.adminHandler.GetShiftSession)
		admin.DELETE("/shifts/sessions/:id", s.adminHandler.DeleteShiftSession)

		// SOP RAG Resources CRUD
		admin.GET("/rag/sops", s.adminHandler.ListSOPs)
		admin.GET("/rag/sops/:id", s.adminHandler.GetSOP)
		admin.DELETE("/rag/sops/:id", s.adminHandler.DeleteSOP)

		// SOP Ingestion Processes CRUD
		admin.GET("/rag/processes", s.adminHandler.ListProcesses)
		admin.GET("/rag/processes/:id", s.adminHandler.GetProcess)
		admin.DELETE("/rag/processes/:id", s.adminHandler.DeleteProcess)

		// Background Job Scheduler Controls
		admin.POST("/scheduler/trigger", s.adminHandler.TriggerSchedulerSweep)
		admin.GET("/scheduler/status", s.adminHandler.GetSchedulerStatus)
	}

	// Operational endpoints grouped by organization context
	orgs := s.engine.Group("/api/v1/organizations/:orgId", UserContextMiddleware(s.adminService, s.cfg))
	{
		// Active ProfileContext (me)
		orgs.GET("/me", s.operationalHandler.GetMe)

		// Site listing
		orgs.GET("/sites", s.operationalHandler.ListSites)

		// Org-level tasks
		orgs.GET("/tasks", s.operationalHandler.GetOrgTasks)

		// Site-level tasks
		orgs.GET("/sites/:siteId/tasks", s.operationalHandler.GetSiteTasks)

		// User-level tasks
		orgs.GET("/sites/:siteId/users/:userId/tasks", s.operationalHandler.GetUserSiteTasks)

		// Task status updates & constraint overrides
		orgs.PATCH("/sites/:siteId/tasks/:id/status", s.operationalHandler.UpdateTaskStatus)
		orgs.POST("/sites/:siteId/tasks/:id/override", s.operationalHandler.OverrideAsset)
		orgs.POST("/sites/:siteId/tasks/:id/claim", s.operationalHandler.ClaimTask)

		// Scoped trades
		orgs.POST("/sites/:siteId/trades", s.operationalHandler.ProposeTrade)
		orgs.GET("/sites/:siteId/trades", s.operationalHandler.ListPendingTrades)
		orgs.POST("/sites/:siteId/trades/:tradeId/accept", s.operationalHandler.AcceptTrade)
		orgs.POST("/sites/:siteId/trades/:tradeId/reject", s.operationalHandler.RejectTrade)

		// Dynamic ad-hoc streaming alerts triggers
		orgs.POST("/sites/:siteId/alerts", s.operationalHandler.TriggerAlert)

		// Interactive MCP chat session
		orgs.POST("/sites/:siteId/users/:userId/sessions/shift/:shiftId/chat", s.MCPHandler.Handler())
		orgs.GET("/sites/:siteId/users/:userId/sessions/shift/:shiftId/chat", s.MCPHandler.Handler())

		// Conversational agent orchestrator chat endpoint
		orgs.POST("/sites/:siteId/users/:userId/sessions/shift/:shiftId/message", s.chatHandler.SendMessage)
	}
}

// Run starts the HTTP server on the configured address and port.
func (s *Server) Run() error {
	addr := s.cfg.Address + ":" + s.cfg.Port
	return s.engine.Run(addr)
}

// Engine exposes the internal Gin engine (specifically for unit testing).
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
