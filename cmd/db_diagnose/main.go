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
	"fmt"
	"log"

	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/rrmcguinness/modenv/pkg/modenv"
)

type Config struct {
	Persistence persistence.DBConfig `toml:"persistence"`
}

func main() {
	var cfg Config
	cloneCfg, err := modenv.Load(&cfg)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	appConfig := cloneCfg.(*Config)

	db, err := persistence.InitDB(appConfig.Persistence)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("=== GTE DATABASE DIAGNOSTICS ===")

	// 1. Sites
	var siteCount int64
	db.Table("sites").Count(&siteCount)
	fmt.Printf("Total Sites: %d\n", siteCount)

	type SiteInfo struct {
		ID   string
		Name string
	}
	var sites []SiteInfo
	db.Table("sites").Select("id, name").Find(&sites)
	for _, s := range sites {
		fmt.Printf("  - Site [%s]: %s\n", s.ID, s.Name)
	}

	// 2. Task Blueprints
	var taskCount int64
	db.Table("tasks").Count(&taskCount)
	fmt.Printf("Total Task Blueprints: %d\n", taskCount)

	// 3. Event Instances
	var instanceCount int64
	db.Table("user_event_instances").Count(&instanceCount)
	fmt.Printf("Total Event Instances: %d\n", instanceCount)

	type InstanceStatusCount struct {
		EventStatus string
		Count       int64
	}
	var statusCounts []InstanceStatusCount
	db.Table("user_event_instances").Select("event_status, count(*) as count").Group("event_status").Find(&statusCounts)
	for _, sc := range statusCounts {
		fmt.Printf("  - Status [%s]: %d\n", sc.EventStatus, sc.Count)
	}

	// 4. Task Executions Grouped by Site
	type SiteExecCount struct {
		SiteID string
		Status string
		Count  int64
	}
	var execCounts []SiteExecCount
	err = db.Table("task_executions").
		Select("events.site_id, task_executions.status, count(*) as count").
		Joins("JOIN user_event_instances ON user_event_instances.id = task_executions.event_instance_id").
		Joins("JOIN user_event_schedules ON user_event_schedules.id = user_event_instances.schedule_id").
		Joins("JOIN events ON events.id = user_event_schedules.event_id").
		Group("events.site_id, task_executions.status").
		Find(&execCounts).Error

	if err != nil {
		log.Fatalf("Failed to query execution counts: %v", err)
	}

	fmt.Println("Task Executions by Site and Status:")
	if len(execCounts) == 0 {
		fmt.Println("  (No task executions found in the database)")
	}
	for _, ec := range execCounts {
		siteName := ec.SiteID
		for _, s := range sites {
			if s.ID == ec.SiteID {
				siteName = s.Name
				break
			}
		}
		fmt.Printf("  - Site [%s] Status [%s]: %d\n", siteName, ec.Status, ec.Count)
	}

	fmt.Println("=================================")
}
