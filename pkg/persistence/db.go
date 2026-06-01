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

package persistence

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

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
	var db *gorm.DB
	var err error

	if cfg.ConnectionString != "" {
		connString := cfg.ConnectionString
		if strings.HasPrefix(connString, "projects/") {
			dialer, errDialer := alloydbconn.NewDialer(context.Background())
			if errDialer != nil {
				return nil, fmt.Errorf("failed to create alloydb dialer: %w", errDialer)
			}

			config, errCfg := pgx.ParseConfig(fmt.Sprintf("host=%s user=postgres sslmode=disable", connString))
			if errCfg != nil {
				config, errCfg = pgx.ParseConfig(connString)
				if errCfg != nil {
					return nil, fmt.Errorf("failed to parse pgx config: %w", errCfg)
				}
			}

			config.DialFunc = func(ctx context.Context, _, _ string) (net.Conn, error) {
				return dialer.Dial(ctx, connString)
			}

			dbURI := stdlib.RegisterConnConfig(config)
			db, err = gorm.Open(postgres.Open(dbURI), &gorm.Config{})
		} else {
			db, err = gorm.Open(postgres.Open(connString), &gorm.Config{})
		}
	} else if strings.HasPrefix(cfg.Host, "projects/") {
		dialer, errDialer := alloydbconn.NewDialer(context.Background())
		if errDialer != nil {
			return nil, fmt.Errorf("failed to create alloydb dialer: %w", errDialer)
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

		config, errCfg := pgx.ParseConfig(connStr)
		if errCfg != nil {
			return nil, fmt.Errorf("failed to parse pgx config: %w", errCfg)
		}

		config.DialFunc = func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.Dial(ctx, cfg.Host)
		}

		dbURI := stdlib.RegisterConnConfig(config)
		db, err = gorm.Open(postgres.Open(dbURI), &gorm.Config{})
	} else {
		connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
		db, err = gorm.Open(postgres.Open(connString), &gorm.Config{})
	}

	if err != nil {
		return nil, err
	}

	// Optimize the GORM database standard connection pool parameters to minimize connection overhead latencies
	sqlDB, errPool := db.DB()
	if errPool == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	return db, nil
}
