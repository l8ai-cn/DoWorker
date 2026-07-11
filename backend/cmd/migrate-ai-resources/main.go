package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/anthropics/agentsmesh/backend/internal/config"
	airesourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	apply := flag.Bool("apply", false, "migrate legacy ai_models and credential EnvBundles")
	dsn := flag.String("dsn", defaultDSN(), "Postgres DSN; defaults to $DATABASE_URL or DB_*")
	cipherKey := flag.String("cipher-key", defaultCipherKey(), "credential cipher key; defaults to $AI_RESOURCE_MIGRATION_CIPHER_KEY or $JWT_SECRET")
	createdBy := flag.Int64("created-by", defaultCreatedBy(), "existing users.id recorded as provider_connections.created_by")
	flag.Parse()

	if *dsn == "" || *cipherKey == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL/--dsn and JWT_SECRET/--cipher-key are required")
		os.Exit(2)
	}
	db, err := gorm.Open(postgres.Open(*dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		fmt.Fprintln(os.Stderr, "db open:", err)
		os.Exit(1)
	}

	migrator := airesourcesvc.NewLegacyMigrator(db, crypto.NewEncryptor(*cipherKey), *createdBy)
	ctx := context.Background()
	if *apply {
		if *createdBy <= 0 {
			fmt.Fprintln(os.Stderr, "--created-by or AI_RESOURCE_MIGRATION_CREATED_BY must be a positive users.id")
			os.Exit(2)
		}
		report, err := migrator.Run(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "migration failed:", err)
			os.Exit(1)
		}
		writeJSON("migration", report)
	}

	check, err := migrator.Check(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "migration check failed:", err)
		os.Exit(1)
	}
	writeJSON("check", check)
	if !check.Clean() {
		os.Exit(1)
	}
}

func defaultCipherKey() string {
	if key := os.Getenv("AI_RESOURCE_MIGRATION_CIPHER_KEY"); key != "" {
		return key
	}
	return os.Getenv("JWT_SECRET")
}

func defaultDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	cfg, err := config.Load()
	if err != nil {
		return ""
	}
	return cfg.Database.DSN()
}

func defaultCreatedBy() int64 {
	raw := os.Getenv("AI_RESOURCE_MIGRATION_CREATED_BY")
	if raw == "" {
		return 0
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func writeJSON(label string, value any) {
	payload, err := json.MarshalIndent(map[string]any{label: value}, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "json encode:", err)
		os.Exit(1)
	}
	fmt.Println(string(payload))
}
