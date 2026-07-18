package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	workerconfigmigration "github.com/anthropics/agentsmesh/backend/internal/service/workerconfigmigration"
	"github.com/anthropics/agentsmesh/backend/internal/service/workerdefinition"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	apply := flag.Bool("apply", false, "write validated named config document bindings")
	dsn := flag.String("dsn", "", "Postgres DSN; defaults to $DATABASE_URL or DB_*")
	definitions := flag.String(
		"definitions-dir",
		defaultDefinitionsDir(),
		"Worker definition catalog directory",
	)
	flag.Parse()

	resolvedDSN := resolveDSN(*dsn)
	if resolvedDSN == "" {
		exitf("DATABASE_URL/--dsn or DB_* is required")
	}
	catalog, err := workerdefinition.Load(*definitions)
	if err != nil {
		exitf("load Worker definitions: %v", err)
	}
	db, err := gorm.Open(postgres.Open(resolvedDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		exitf("open database: %v", err)
	}
	migrator, err := workerconfigmigration.New(db, catalog)
	if err != nil {
		exitf("initialize migration: %v", err)
	}
	if *apply {
		report, err := migrator.Run(context.Background())
		writeReport("migration", report)
		if err != nil {
			exitf("migration failed: %v", err)
		}
	}
	report, err := migrator.Check(context.Background())
	if err != nil {
		exitf("migration check failed: %v", err)
	}
	writeReport("check", report)
	if !report.Clean() {
		os.Exit(1)
	}
}

func resolveDSN(value string) string {
	if value != "" {
		return value
	}
	if value = os.Getenv("DATABASE_URL"); value != "" {
		return value
	}
	cfg, err := config.Load()
	if err != nil {
		return ""
	}
	return cfg.Database.DSN()
}

func defaultDefinitionsDir() string {
	if value := os.Getenv("WORKER_DEFINITIONS_DIR"); value != "" {
		return value
	}
	return "config/worker-types"
}

func writeReport(label string, report workerconfigmigration.Report) {
	raw, err := json.MarshalIndent(map[string]any{label: report}, "", "  ")
	if err != nil {
		exitf("encode %s report: %v", label, err)
	}
	fmt.Println(string(raw))
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
