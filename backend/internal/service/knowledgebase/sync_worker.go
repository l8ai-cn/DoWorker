package knowledgebase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/knowledgebase"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/knowledgebase/connector"
)

// SyncWorker periodically materializes external sources (feishu / dingtalk /
// google) into each KB's raw/ directory — same lifecycle contract as
// MarketplaceWorker (idempotent Start, Stop cancels the loop).
type SyncWorker struct {
	svc        *Service
	connectors map[string]connector.Connector
	interval   time.Duration

	startOnce sync.Once
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewSyncWorker returns nil when the KB service is disabled.
func NewSyncWorker(svc *Service, interval time.Duration) *SyncWorker {
	if svc == nil {
		return nil
	}
	if interval <= 0 {
		interval = time.Hour
	}
	return &SyncWorker{
		svc:        svc,
		connectors: connector.NewRegistry(),
		interval:   interval,
	}
}

func (w *SyncWorker) Start(ctx context.Context) {
	w.startOnce.Do(func() {
		ctx, w.cancel = context.WithCancel(ctx)
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			ticker := time.NewTicker(w.interval)
			defer ticker.Stop()
			w.syncAll(ctx)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					w.syncAll(ctx)
				}
			}
		}()
	})
}

func (w *SyncWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
}

func (w *SyncWorker) SyncSingle(ctx context.Context, orgID int64, slug string) error {
	kb, err := w.svc.GetBySlug(ctx, orgID, slug)
	if err != nil {
		return err
	}
	if kb.SourceType == knowledgebase.SourceTypeGit {
		return fmt.Errorf("%w: git knowledge bases are not synced from external sources", ErrInvalidInput)
	}
	conn, ok := w.connectors[kb.SourceType]
	if !ok {
		return fmt.Errorf("%w: no connector for source_type %q", ErrInvalidInput, kb.SourceType)
	}
	return w.svc.SyncFromConnector(ctx, kb, conn)
}

func (w *SyncWorker) syncAll(ctx context.Context) {
	kbs, err := w.svc.repo.ListExternal(ctx)
	if err != nil {
		w.svc.log.Warn("kb sync: list external failed", "error", err)
		return
	}
	for _, kb := range kbs {
		if ctx.Err() != nil {
			return
		}
		conn, ok := w.connectors[kb.SourceType]
		if !ok {
			w.svc.log.Warn("kb sync: no connector", "slug", kb.Slug, "source_type", kb.SourceType)
			continue
		}
		if err := w.svc.SyncFromConnector(ctx, kb, conn); err != nil {
			w.svc.log.Warn("kb sync failed", "slug", kb.Slug, "error", err)
		}
	}
}
