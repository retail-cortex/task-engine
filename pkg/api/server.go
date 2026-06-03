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
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/agents"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"gorm.io/gorm"
)

type OAuthConfig struct {
	ClientID string `toml:"client_id"`
	Secret   string `toml:"secret"`
}

// Config holds the server binding configuration.
type Config struct {
	Port    string      `toml:"port"`
	Address string      `toml:"address"`
	OAuth   OAuthConfig `toml:"oauth"`
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
			return nil, fmt.Errorf("failed to bootstrap ADK agent session state postgres service: %w", errSess)
		}
	} else {
		// Mock compatible fallback memory session module (used inside sandboxes and unit tests)
		agentsSessionService = agents.NewInMemorySessionService()
	}

	adminHandler := NewAdminHandler(adminService, schedulerService)
	operationalHandler := NewOperationalHandler(taskService, shiftService, automationService, cfg)
	mcpHandler, err := NewMCPHandler(taskService, ragService, shiftService, automationService)
	if err != nil {
		return nil, err
	}
	chatHandler := NewChatHandler(taskService, shiftService, ragService, automationService, agentsSessionService)

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
	}

	s.setupRoutes()

	return s, nil
}

func (s *Server) setupRoutes() {
	// Health endpoint
	s.engine.GET("/health/readiness", s.operationalHandler.Readiness)

	// Admin endpoints
	admin := s.engine.Group("/api/v1/admin", UserContextMiddleware(s.adminService, s.cfg.OAuth.ClientID))
	{
		admin.GET("/users", s.adminHandler.ListUsers)
		admin.POST("/roles", s.adminHandler.CreateRole)
		admin.PUT("/users/:id/roles", s.adminHandler.AssignRole)
		admin.POST("/organizations", s.adminHandler.CreateOrganization)
		admin.GET("/organizations", s.adminHandler.ListOrganizations)
		admin.PUT("/organizations/:orgId/users/:userId", s.adminHandler.AssignUserToOrganization)
		admin.POST("/organizations/:orgId/sites", s.adminHandler.CreateSite)
		admin.POST("/organizations/:orgId/sites/:siteId/locations", s.adminHandler.CreateLocation)
		admin.POST("/organizations/:orgId/sites/:siteId/locations/:locationId/assets", s.adminHandler.CreateAsset)
		admin.POST("/tasks/templates", s.adminHandler.CreateTaskTemplate)

		// Background Job Scheduler Controls
		admin.POST("/scheduler/trigger", s.adminHandler.TriggerSchedulerSweep)
		admin.GET("/scheduler/status", s.adminHandler.GetSchedulerStatus)
	}

	// Operational endpoints grouped by organization context
	orgs := s.engine.Group("/api/v1/organizations/:orgId", UserContextMiddleware(s.adminService, s.cfg.OAuth.ClientID))
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
