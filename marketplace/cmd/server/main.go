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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
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

	server := &http.Server{
		Addr: cfg.HTTPAddress,
		Handler: api.NewRouter(api.Dependencies{
			Ready: sqlDB.PingContext,
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
