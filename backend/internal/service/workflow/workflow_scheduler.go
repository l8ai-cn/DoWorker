package workflow

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/service/instance"
	"github.com/robfig/cron/v3"
)

type WorkflowScheduler struct {
	workflowService *WorkflowService
	orchestrator    *WorkflowOrchestrator
	orgProvider     instance.LocalOrgProvider
	logger          *slog.Logger
	cronParser      cron.Parser
	stopCh          chan struct{}
	stopOnce        sync.Once
	wg              sync.WaitGroup
}

func NewWorkflowScheduler(
	workflowService *WorkflowService,
	orchestrator *WorkflowOrchestrator,
	orgProvider instance.LocalOrgProvider,
	logger *slog.Logger,
) *WorkflowScheduler {
	return &WorkflowScheduler{
		workflowService: workflowService,
		orchestrator:    orchestrator,
		orgProvider:     orgProvider,
		logger:          logger.With("component", "loop_scheduler"),
		cronParser:      cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		stopCh:          make(chan struct{}),
	}
}

func (s *WorkflowScheduler) getOrgIDs() []int64 {
	if s.orgProvider == nil {
		return nil
	}
	return s.orgProvider.GetLocalOrgIDs()
}

func (s *WorkflowScheduler) Start() {
	if err := s.InitializeNextRunTimes(context.Background()); err != nil {
		s.logger.Error("failed to initialize next run times", "error", err)
	}

	s.wg.Add(2)

	go s.safeLoop("cron_trigger", s.runCronLoop)
	go s.safeLoop("timeout_detection", s.runTimeoutLoop)

	s.logger.Info("workflow scheduler started (cron check: 30s, timeout check: 60s)")
}

func (s *WorkflowScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		s.wg.Wait()
		s.logger.Info("workflow scheduler stopped")
	})
}

func (s *WorkflowScheduler) safeLoop(name string, fn func()) {
	defer s.wg.Done()
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("panic in scheduler goroutine, restarting after cooldown",
						"goroutine", name, "panic", r)
				}
			}()
			fn()
		}()
		select {
		case <-s.stopCh:
			return
		default:
			time.Sleep(5 * time.Second)
		}
	}
}

func (s *WorkflowScheduler) runCronLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.CheckAndTriggerCronLoops(context.Background()); err != nil {
				s.logger.Error("cron workflow check failed", "error", err)
			}
		}
	}
}

func (s *WorkflowScheduler) runTimeoutLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.orchestrator.CheckTimeoutRuns(context.Background(), s.getOrgIDs()); err != nil {
				s.logger.Error("timeout check failed", "error", err)
			}
			if err := s.orchestrator.CheckApprovalTimeouts(context.Background(), s.getOrgIDs()); err != nil {
				s.logger.Error("approval timeout check failed", "error", err)
			}
			if err := s.orchestrator.CheckIdleLoopPods(context.Background(), s.getOrgIDs()); err != nil {
				s.logger.Error("idle workflow pod check failed", "error", err)
			}
			if err := s.orchestrator.CleanupOrphanPendingRuns(context.Background(), s.getOrgIDs()); err != nil {
				s.logger.Error("orphan cleanup failed", "error", err)
			}
		}
	}
}

func (s *WorkflowScheduler) CalculateNextRun(cronExpr string) (*time.Time, error) {
	schedule, err := s.cronParser.Parse(cronExpr)
	if err != nil {
		return nil, err
	}
	next := schedule.Next(time.Now())
	return &next, nil
}

func (s *WorkflowScheduler) InitializeNextRunTimes(ctx context.Context) error {
	orgIDs := s.getOrgIDs()

	workflows, err := s.workflowService.FindWorkflowsNeedingNextRun(ctx, orgIDs)
	if err != nil {
		return err
	}

	for _, workflow := range workflows {
		if workflow.CronExpression != nil {
			nextRunAt, err := s.CalculateNextRun(*workflow.CronExpression)
			if err != nil {
				s.logger.Error("invalid cron expression", "workflow_id", workflow.ID, "error", err)
				continue
			}
			if err := s.workflowService.UpdateNextRunAt(ctx, workflow.ID, nextRunAt); err != nil {
				s.logger.Error("failed to set initial next_run_at", "error", err, "workflow_id", workflow.ID)
			}
		}
	}

	return nil
}
