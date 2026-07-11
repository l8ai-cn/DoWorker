package goalloop

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const timeoutCheckInterval = time.Minute

type TimeoutMonitor struct {
	service *Service
	logger  *slog.Logger
	stop    chan struct{}
	once    sync.Once
}

func NewTimeoutMonitor(service *Service, logger *slog.Logger) *TimeoutMonitor {
	return &TimeoutMonitor{
		service: service,
		logger:  logger,
		stop:    make(chan struct{}),
	}
}

func (m *TimeoutMonitor) Start() {
	go func() {
		m.check()
		ticker := time.NewTicker(timeoutCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.check()
			case <-m.stop:
				return
			}
		}
	}()
}

func (m *TimeoutMonitor) Stop() {
	m.once.Do(func() { close(m.stop) })
}

func (m *TimeoutMonitor) check() {
	if err := m.service.ExpireTimedOut(context.Background(), time.Now()); err != nil {
		m.logger.Error("failed to expire timed out goal loops", "error", err)
	}
}
