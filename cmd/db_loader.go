package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/rrmcguinness/modenv/pkg/modenv"
	"gorm.io/gorm"
)

// AppConfig maps the persistence settings required for DB connection.
type AppConfig struct {
	Persistence persistence.DBConfig `toml:"persistence"`
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", []string(*i))
}

func (i *arrayFlags) Set(value string) error {
	files := strings.Split(value, ",")
	for _, file := range files {
		trimmed := strings.TrimSpace(file)
		if trimmed != "" {
			*i = append(*i, trimmed)
		}
	}
	return nil
}

func main() {
	migrateFlag := flag.Bool("migrate", false, "Run database auto-migration and exit")
	var sqlFiles arrayFlags
	flag.Var(&sqlFiles, "file", "Path to the SQL script file(s) to run (can be comma-separated or specified multiple times)")
	flag.Parse()

	if !*migrateFlag && len(sqlFiles) == 0 {
		log.Fatalf("Error: at least one flag (-migrate or -file) must be provided")
	}

	log.Printf("Loading environment configurations using modenv...")
	var cfg AppConfig
	cloneCfg, err := modenv.Load(&cfg)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	appConfig := cloneCfg.(*AppConfig)

	// Support DB_CONNECTION_STRING override
	if envConn := os.Getenv("DB_CONNECTION_STRING"); envConn != "" {
		appConfig.Persistence.ConnectionString = envConn
	}

	log.Printf("Initializing AlloyDB/PostgreSQL connection...")
	db, err := persistence.InitDB(appConfig.Persistence)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if *migrateFlag {
		log.Printf("Running schema migration...")
		if err := MigrateSchema(db); err != nil {
			log.Fatalf("Database migration failed: %v", err)
		}
		log.Printf("Database migration completed successfully!")
	}

	if len(sqlFiles) > 0 {
		for _, file := range sqlFiles {
			filePath := file
			if !filepath.IsAbs(filePath) {
				// 1. Resolve relative to BUILD_WORKING_DIRECTORY (if run under bazel run)
				if bwd := os.Getenv("BUILD_WORKING_DIRECTORY"); bwd != "" {
					testPath := filepath.Join(bwd, filePath)
					if _, err := os.Stat(testPath); err == nil {
						filePath = testPath
					}
				}
				// 2. Resolve relative to MODENV_PREFIX as fallback
				if !filepath.IsAbs(filePath) {
					if prefix := os.Getenv("MODENV_PREFIX"); prefix != "" {
						testPath := filepath.Join(prefix, filePath)
						if _, err := os.Stat(testPath); err == nil {
							filePath = testPath
						}
					}
				}
			}

			log.Printf("Executing SQL script file: %s...", filePath)
			if err := LoadDevEnv(db, filePath); err != nil {
				log.Fatalf("Failed to run seed SQL script %s: %v", filePath, err)
			}
			log.Printf("Database seeded successfully from %s!", filePath)
		}
	}
}

// MigrateSchema runs GORM AutoMigrate for all defined database structures.
func MigrateSchema(db *gorm.DB) error {
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		return err
	}
	return db.AutoMigrate(
		&model.Organization{},
		&model.TimeZone{},
		&model.ICAOCode{},
		&model.Site{},
		&model.Location{},
		&model.Asset{},
		&model.User{},
		&model.Role{},
		&model.UserRole{},
		&model.UserSite{},
		&model.UserCertification{},
		&model.Task{},
		&model.TaskAsset{},
		&model.TaskApprovalRule{},
		&model.TaskSOP{},
		&model.TaskExecution{},
		&model.TaskExecutionAudit{},
		&model.TaskTrade{},
		&model.ShiftAgentSession{},
		&model.SOP{},
		&model.SOPProcess{},
		&model.SOPChunk{},
		&model.Event{},
		&model.UserEventSchedule{},
		&model.UserEventInstance{},
		&model.UserAvailability{},
	)
}

// LoadDevEnv reads the seed SQL file and executes it within a database transaction.
func LoadDevEnv(db *gorm.DB, sqlFilePath string) error {
	content, err := os.ReadFile(sqlFilePath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file %s: %w", sqlFilePath, err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("failed to execute SQL statements: %w", err)
		}
		return nil
	})
}
