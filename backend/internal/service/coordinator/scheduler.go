package coordinator

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const schedulerTick = 30 * time.Second

// Scheduler runs enabled coordinator projects on their per-project scan
// interval. A single base ticker drives all projects; each project's
// scan_interval_seconds gates how often it actually runs. Designed for the
// single-instance backend deployment (matches LoopScheduler).
type Scheduler struct {
	service  *Service
	logger   *slog.Logger
	lastRun  map[int64]time.Time
	mu       sync.Mutex
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

func NewScheduler(service *Service, logger *slog.Logger) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{
		service: service,
		logger:  logger.With("component", "coordinator_scheduler"),
		lastRun: map[int64]time.Time{},
		stopCh:  make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.loop()
	s.logger.Info("coordinator scheduler started", "tick", schedulerTick.String())
}

func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		s.wg.Wait()
		s.logger.Info("coordinator scheduler stopped")
	})
}

func (s *Scheduler) loop() {
	defer s.wg.Done()
	ticker := time.NewTicker(schedulerTick)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.tick(context.Background())
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("panic in coordinator tick", "panic", r)
		}
	}()

	projects, err := s.service.Store().ListEnabledProjects(ctx)
	if err != nil {
		s.logger.Error("list enabled projects failed", "error", err)
		return
	}
	for _, project := range projects {
		if !s.due(project.ID, project.ScanInterval()) {
			continue
		}
		result, err := s.service.RunProject(ctx, project)
		if err != nil {
			s.logger.Error("run project failed", "project_id", project.ID, "error", err)
			continue
		}
		if result.Dispatched > 0 || len(result.Errors) > 0 {
			s.logger.Info("coordinator project run",
				"project_id", project.ID, "scanned", result.Scanned,
				"dispatched", result.Dispatched, "errors", len(result.Errors))
		}
	}
}

func (s *Scheduler) due(projectID int64, interval time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Since(s.lastRun[projectID]) < interval {
		return false
	}
	s.lastRun[projectID] = time.Now()
	return true
}
