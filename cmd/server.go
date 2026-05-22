package main

import (
	"log"
	"os"

	"github.com/rmcguinness/gemini_task_engine/pkg/api"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"github.com/rrmcguinness/modenv/pkg/modenv"
)

// AppConfig represents the root hierarchical configuration.
type AppConfig struct {
	Persistence persistence.DBConfig `toml:"persistence"`
	Server      api.Config           `toml:"server"`
}

func main() {
	log.Printf("Loading environment configurations using modenv...")
	var cfg AppConfig
	cloneCfg, err := modenv.Load(&cfg)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	appConfig := cloneCfg.(*AppConfig)

	// Keep backward compatibility for DB_CONNECTION_STRING environment override
	if envConn := os.Getenv("DB_CONNECTION_STRING"); envConn != "" {
		appConfig.Persistence.ConnectionString = envConn
	}

	log.Printf("Initializing AlloyDB/PostgreSQL connection...")
	db, err := persistence.InitDB(appConfig.Persistence)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Printf("Database successfully connected. Setting up GORM repositories...")
	userRepo := persistence.NewUserRepository(db)
	orgRepo := persistence.NewOrganizationRepository(db)
	siteRepo := persistence.NewSiteRepository(db)
	taskRepo := persistence.NewTaskRepository(db)
	execRepo := persistence.NewTaskExecutionRepository(db)
	sessionRepo := persistence.NewShiftAgentSessionRepository(db)
	sopRepo := persistence.NewSOPRepository(db)

	log.Printf("GORM repositories successfully instantiated. Setting up Core services...")
	adminService := service.NewAdminService(userRepo, orgRepo, siteRepo, taskRepo)
	taskService := service.NewTaskService(execRepo, siteRepo)
	shiftService := service.NewShiftService(sessionRepo, userRepo)
	ragService := service.NewRAGService(sopRepo)

	log.Printf("Core services instantiated. Initializing Gin base server with MCP capabilities...")
	srv, err := api.NewServer(appConfig.Server, adminService, taskService, shiftService, ragService)
	if err != nil {
		log.Fatalf("Failed to initialize Gin server: %v", err)
	}

	log.Printf("Server listening on %s:%s...", appConfig.Server.Address, appConfig.Server.Port)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
