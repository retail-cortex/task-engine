package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
)

// Config holds the server binding configuration.
type Config struct {
	Port    string `toml:"port"`
	Address string `toml:"address"`
}

// Server defines the GIN web server and routing engine.
type Server struct {
	cfg                Config
	engine             *gin.Engine
	adminService       service.AdminService
	taskService        service.TaskService
	shiftService       service.ShiftService
	ragService         service.RAGService
	adminHandler       *AdminHandler
	operationalHandler *OperationalHandler
	mcpHandler         *MCPHandler
}

// NewServer instantiates the Gin engine and mounts all route controllers.
func NewServer(
	cfg Config,
	adminService service.AdminService,
	taskService service.TaskService,
	shiftService service.ShiftService,
	ragService service.RAGService,
) (*Server, error) {
	engine := gin.New()

	// Base middlewares
	engine.Use(gin.Recovery())
	engine.Use(CORSConfigMiddleware())
	engine.Use(UserContextMiddleware())

	adminHandler := NewAdminHandler(adminService)
	operationalHandler := NewOperationalHandler(taskService, shiftService)
	mcpHandler, err := NewMCPHandler(taskService, ragService, shiftService)
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg:                cfg,
		engine:             engine,
		adminService:       adminService,
		taskService:        taskService,
		shiftService:       shiftService,
		ragService:         ragService,
		adminHandler:       adminHandler,
		operationalHandler: operationalHandler,
		mcpHandler:         mcpHandler,
	}

	s.setupRoutes()

	return s, nil
}

func (s *Server) setupRoutes() {
	// Health endpoint
	s.engine.GET("/health/readiness", s.operationalHandler.Readiness)

	// Admin endpoints
	admin := s.engine.Group("/api/v1/admin")
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
	}

	// Operational endpoints grouped by organization context
	orgs := s.engine.Group("/api/v1/organizations/:orgId")
	{
		// Org-level tasks
		orgs.GET("/tasks", s.operationalHandler.GetOrgTasks)

		// Site-level tasks
		orgs.GET("/sites/:siteId/tasks", s.operationalHandler.GetSiteTasks)

		// User-level tasks
		orgs.GET("/sites/:siteId/users/:userId/tasks", s.operationalHandler.GetUserSiteTasks)

		// Task status updates & constraint overrides
		orgs.PATCH("/sites/:siteId/tasks/:id/status", s.operationalHandler.UpdateTaskStatus)
		orgs.POST("/sites/:siteId/tasks/:id/override", s.operationalHandler.OverrideAsset)

		// Scoped trades
		orgs.POST("/sites/:siteId/trades", s.operationalHandler.ProposeTrade)

		// Interactive MCP chat session
		orgs.POST("/sites/:siteId/users/:userId/sessions/shift/:shiftId/chat", s.mcpHandler.Handler())
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
