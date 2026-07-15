package main

import (
	"context"
	"log"
	"time"

	marketplacepostgres "github.com/anthropics/agentsmesh/marketplace/internal/infra/postgres"
)

func runExpertCatalogSync(
	ctx context.Context,
	syncer *marketplacepostgres.ExpertCatalogSynchronizer,
) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := syncer.Sync(ctx); err != nil {
				log.Printf("sync published expert catalog: %v", err)
			}
		}
	}
}
