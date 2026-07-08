package runner

import (
	"context"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
)

func (d *PendingCommandDrainer) StartExpirySweeper(ctx context.Context) {
	ticker := time.NewTicker(d.sweepInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				d.sweepExpired(ctx)
				d.drainOnlineRunnersWithBacklog(ctx)
			}
		}
	}()
}

func (d *PendingCommandDrainer) sweepExpired(ctx context.Context) {
	rows, err := d.repo.ListExpired(ctx, time.Now(), 100)
	if err != nil {
		d.logger.Error("failed to list expired pending commands", "error", err)
		return
	}
	for _, row := range rows {
		d.handleExpired(ctx, row)
	}
}

func (d *PendingCommandDrainer) handleExpired(ctx context.Context, row *agentpod.PendingCommand) {
	_ = d.repo.Delete(ctx, row.ID)
	if row.CommandType != agentpod.CommandTypeCreatePod {
		d.logger.Debug("expired send_prompt discarded", "pod_key", row.PodKey)
		return
	}
	if d.expiryMarker != nil {
		msg := "Task expired after waiting for runner to come online"
		_ = d.expiryMarker.MarkQueueExpired(ctx, row.PodKey, agentpod.ErrCodeQueueExpired, msg)
	}
	if d.onQueueExpired != nil {
		d.onQueueExpired(ctx, row.PodKey)
	}
	publishQueueEvent(d.eventBus, d.logger, eventbus.EventPodQueueExpired, row.OrganizationID, row.PodKey, map[string]interface{}{
		"pod_key":   row.PodKey,
		"runner_id": row.RunnerID,
	})
}

// Safety net for lost online/terminated events (callback panic, event landing
// on another backend replica): every sweep, re-trigger drain for connected
// runners that still have backlog. Single-flight in DrainRunner dedupes.
func (d *PendingCommandDrainer) drainOnlineRunnersWithBacklog(ctx context.Context) {
	if d.connChecker == nil {
		return
	}
	ids, err := d.repo.ListRunnerIDsWithPending(ctx, 200)
	if err != nil {
		d.logger.Error("failed to list runners with backlog", "error", err)
		return
	}
	for _, runnerID := range ids {
		if d.connChecker.IsConnected(runnerID) {
			d.DrainRunner(runnerID)
		}
	}
}
