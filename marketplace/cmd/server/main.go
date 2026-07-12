package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anthropics/agentsmesh/marketplace/internal/api"
	"github.com/anthropics/agentsmesh/marketplace/internal/config"
	marketplacepostgres "github.com/anthropics/agentsmesh/marketplace/internal/infra/postgres"
	"github.com/anthropics/agentsmesh/marketplace/internal/integration/identity"
	runtimebridge "github.com/anthropics/agentsmesh/marketplace/internal/integration/runtime"
	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		databaseURL := os.Getenv("MARKETPLACE_MIGRATION_DATABASE_URL")
		if databaseURL == "" {
			log.Fatal("MARKETPLACE_MIGRATION_DATABASE_URL is required")
		}
		if err := migrateUp(databaseURL); err != nil {
			log.Fatalf("migrate marketplace database: %v", err)
		}
		return
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL))
	if err != nil {
		log.Fatalf("connect marketplace database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("open marketplace database pool: %v", err)
	}
	defer sqlDB.Close()
	identityVerifier, err := identity.NewJWKSVerifier(identity.JWKSConfig{
		URL:             cfg.IdentityJWKSURL,
		Issuer:          cfg.IdentityIssuer,
		Audience:        cfg.IdentityAudience,
		RefreshInterval: 5 * time.Minute,
	}, &http.Client{Timeout: 5 * time.Second})
	if err != nil {
		log.Fatalf("configure marketplace identity: %v", err)
	}
	runtimeClient, err := runtimebridge.NewClient(
		cfg.RuntimeBridgeURL,
		cfg.InternalAPISecret,
		&http.Client{Timeout: 30 * time.Second},
	)
	if err != nil {
		log.Fatalf("configure marketplace runtime bridge: %v", err)
	}
	installationRepository := marketplacepostgres.NewInstallationRepository(db)

	server := &http.Server{
		Addr: cfg.HTTPAddress,
		Handler: api.NewRouter(api.Dependencies{
			Ready:      sqlDB.PingContext,
			Storefront: service.NewStorefrontService(marketplacepostgres.NewStorefrontRepository(db)),
			Identity:   identityVerifier,
			Installations: service.NewInstallationOrchestrationService(
				installationRepository,
				runtimeClient,
				time.Now,
			),
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		errs <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("marketplace server: %v", err)
		}
	case <-signals:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("shutdown marketplace server: %v", err)
		}
	}
}
