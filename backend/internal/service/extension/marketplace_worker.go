package extension

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// MarketplaceWorker runs background sync for the official MCP Registry.
// (Skill sourcing moved to the unified skill catalog — imports are explicit
// user actions on the skill service, not a background registry poll.)
type MarketplaceWorker struct {
	registrySyncer *McpRegistrySyncer
	syncInterval   time.Duration

	cancel    context.CancelFunc
	wg        sync.WaitGroup
	startOnce sync.Once
}

// NewMarketplaceWorker creates a new MarketplaceWorker.
// registrySyncer may be nil if MCP Registry sync is disabled.
func NewMarketplaceWorker(registrySyncer *McpRegistrySyncer, syncInterval time.Duration) *MarketplaceWorker {
	return &MarketplaceWorker{
		registrySyncer: registrySyncer,
		syncInterval:   syncInterval,
	}
}

// Start begins the background sync loop.
// It performs an initial sync, then repeats at the configured interval.
// Calling Start multiple times is safe; only the first call has any effect.
func (w *MarketplaceWorker) Start(ctx context.Context) {
	w.startOnce.Do(func() {
		if w.registrySyncer == nil {
			return
		}
		ctx, w.cancel = context.WithCancel(ctx)

		slog.InfoContext(ctx, "MarketplaceWorker starting", "interval", w.syncInterval)

		w.wg.Add(1)
		go func() {
			defer w.wg.Done()

			// Initial sync after a short delay to let the system warm up
			timer := time.NewTimer(10 * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}

			w.syncMcpRegistry(ctx)

			ticker := time.NewTicker(w.syncInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					w.syncMcpRegistry(ctx)
				}
			}
		}()
	})
}

// Stop gracefully stops the worker
func (w *MarketplaceWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	slog.Info("MarketplaceWorker stopped")
}

func (w *MarketplaceWorker) syncMcpRegistry(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	slog.InfoContext(ctx, "MarketplaceWorker: starting MCP Registry sync")
	if err := w.registrySyncer.Sync(ctx); err != nil {
		slog.ErrorContext(ctx, "MarketplaceWorker: MCP Registry sync failed", "error", err)
	} else {
		slog.InfoContext(ctx, "MarketplaceWorker: MCP Registry sync completed")
	}
}
