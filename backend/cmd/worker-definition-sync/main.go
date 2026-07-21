package main

import (
	"context"
	"fmt"
	"log"

	"github.com/l8ai-cn/agentcloud/backend/internal/config"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/database"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}
	catalog, err := workerdefinition.Load(cfg.WorkerDefinitionsDir)
	if err != nil {
		log.Fatalf("load Worker definitions: %v", err)
	}
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Warn)})
	defer func() {
		if closeErr := database.Close(db); closeErr != nil {
			log.Printf("close database: %v", closeErr)
		}
	}()
	count, err := workerdefinition.SyncAgentProjections(context.Background(), db, catalog)
	if err != nil {
		log.Fatalf("sync Worker definition projections: %v", err)
	}
	fmt.Printf("synced %d Worker definition projections\n", count)
}
