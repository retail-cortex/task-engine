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

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rmcguinness/gemini_task_engine/pkg/api"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"github.com/rmcguinness/gemini_task_engine/pkg/telemetry"
	"github.com/rrmcguinness/modenv/pkg/modenv"
)

// ServerConfig represents the root hierarchical configuration.
type ServerConfig struct {
	Persistence persistence.DBConfig    `toml:"persistence"`
	Server      api.Config              `toml:"server"`
	Scheduler   service.SchedulerConfig `toml:"scheduler"`
}

func main() {
	// Force immediate local credential resolution and native DNS resolution
	// to prevent 30-second cold-start metadata server timeouts in local offline contexts.
	// In deployed Cloud Run environments (dev/prod), we must use the real metadata server.
	if os.Getenv("MODENV_RUNTIME") == "" || os.Getenv("MODENV_RUNTIME") == "local" {
		os.Setenv("GCE_METADATA_HOST", "127.0.0.1:9999")
	}
	os.Setenv("GRPC_DNS_RESOLVER", "native")

	// Initialize OpenTelemetry Tracing
	otelShutdown, otelErr := telemetry.InitTelemetry(context.Background(), "gemini-task-api")
	if otelErr != nil {
		log.Printf("WARNING: Failed to initialize OpenTelemetry: %v. Traces will not be exported.", otelErr)
	} else {
		defer otelShutdown()

		// Setup signal handling to ensure traces are flushed on sudden SIGINT/SIGTERM
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			sig := <-sigChan
			log.Printf("System received shutdown signal: %v. Triggering OpenTelemetry flush...", sig)
			otelShutdown()
			log.Println("Telemetry successfully flushed. Exiting.")
			os.Exit(0)
		}()
	}

	log.Printf("Loading environment configurations using modenv...")
	var cfg ServerConfig
	cloneCfg, err := modenv.Load(&cfg)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	appConfig := cloneCfg.(*ServerConfig)

	// Keep backward compatibility for DB_CONNECTION_STRING environment override
	if envConn := os.Getenv("DB_CONNECTION_STRING"); envConn != "" {
		appConfig.Persistence.ConnectionString = envConn
	}

	// Support PORT environment variable override for Cloud Run deployments
	if envPort := os.Getenv("PORT"); envPort != "" {
		appConfig.Server.Port = envPort
	}

	log.Printf("Initializing AlloyDB/PostgreSQL connection...")
	db, dbErr := persistence.InitDB(appConfig.Persistence)
	if dbErr != nil {
		log.Printf("WARNING: Database connectivity failed: %v. Bootstrapping server in degraded offline mode.", dbErr)
		var errInit error
		db, errInit = persistence.NewOfflineDB(dbErr)
		if errInit != nil {
			log.Fatalf("Failed to initialize offline fallback database: %v", errInit)
		}
	} else {
		log.Printf("Database successfully connected. Setting up GORM repositories...")
	}
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
	embeddingGen := service.NewDefaultEmbeddingGenerator()
	ragService := service.NewRAGService(sopRepo, embeddingGen)
	automationService := service.NewAutomationService(execRepo, taskRepo, siteRepo, userRepo, sopRepo, embeddingGen)
	schedulerService := service.NewSchedulerServiceWithConfig(db, ragService, taskService, automationService, appConfig.Scheduler)

	// Start distributed scheduler daemon loops before booting HTTP server
	if err := schedulerService.Start(context.Background()); err != nil {
		log.Fatalf("Failed to bootstrap distributed scheduler daemon: %v", err)
	}
	defer schedulerService.Stop()

	log.Printf("Core services instantiated. Initializing Gin base server with MCP capabilities...")
	srv, err := api.NewServer(appConfig.Server, adminService, taskService, shiftService, ragService, automationService, schedulerService, db)
	if err != nil {
		log.Fatalf("Failed to initialize Gin server: %v", err)
	}

	log.Printf("Server listening on %s:%s...", appConfig.Server.Address, appConfig.Server.Port)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
