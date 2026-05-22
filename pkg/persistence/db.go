package persistence

import (
	"context"
	"fmt"
	"net"
	"strings"

	"cloud.google.com/go/alloydbconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBConfig holds the database connection configuration.
type DBConfig struct {
	Host             string `toml:"host"`
	Port             string `toml:"port"`
	User             string `toml:"user"`
	Password         string `toml:"password"`
	DBName           string `toml:"dbname"`
	SSLMode          string `toml:"sslmode"`
	ConnectionString string `toml:"connection_string"`
}

// InitDB initializes a new gorm.DB connection using the provided DBConfig. It automatically registers and utilizes
// the Google AlloyDB Dialer if the connection string is formatted as an AlloyDB Instance URI
// (starts with "projects/"), falling back to standard pgx TCP connection otherwise.
func InitDB(cfg DBConfig) (*gorm.DB, error) {
	if cfg.ConnectionString != "" {
		connString := cfg.ConnectionString
		if strings.HasPrefix(connString, "projects/") {
			dialer, err := alloydbconn.NewDialer(context.Background())
			if err != nil {
				return nil, fmt.Errorf("failed to create alloydb dialer: %w", err)
			}

			// Format connection parameters. Since IAM auth handles identity, we default to sslmode=disable.
			config, err := pgx.ParseConfig(fmt.Sprintf("host=%s user=postgres sslmode=disable", connString))
			if err != nil {
				// Try parsing connection string directly if it contains custom parameters
				config, err = pgx.ParseConfig(connString)
				if err != nil {
					return nil, fmt.Errorf("failed to parse pgx config: %w", err)
				}
			}

			config.DialFunc = func(ctx context.Context, _, _ string) (net.Conn, error) {
				return dialer.Dial(ctx, connString)
			}

			dbURI := stdlib.RegisterConnConfig(config)
			return gorm.Open(postgres.Open(dbURI), &gorm.Config{})
		}

		// Fallback to standard GORM postgres initialization
		return gorm.Open(postgres.Open(connString), &gorm.Config{})
	}

	// Dynamic construction from elements
	if strings.HasPrefix(cfg.Host, "projects/") {
		dialer, err := alloydbconn.NewDialer(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to create alloydb dialer: %w", err)
		}

		user := cfg.User
		if user == "" {
			user = "postgres"
		}
		connStr := fmt.Sprintf("host=%s user=%s sslmode=disable", cfg.Host, user)
		if cfg.DBName != "" {
			connStr += fmt.Sprintf(" dbname=%s", cfg.DBName)
		}
		if cfg.Password != "" {
			connStr += fmt.Sprintf(" password=%s", cfg.Password)
		}

		config, err := pgx.ParseConfig(connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pgx config: %w", err)
		}

		config.DialFunc = func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.Dial(ctx, cfg.Host)
		}

		dbURI := stdlib.RegisterConnConfig(config)
		return gorm.Open(postgres.Open(dbURI), &gorm.Config{})
	}

	// Standard PostgreSQL GORM initialization
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
	return gorm.Open(postgres.Open(connString), &gorm.Config{})
}
