package config

import "time"

type PendingQueueConfig struct {
	Enabled        bool
	MaxPerRunner   int
	DefaultTTL     time.Duration
	SweepInterval  time.Duration
}

func loadPendingQueueConfig() PendingQueueConfig {
	return PendingQueueConfig{
		Enabled:       getEnvBool("PENDING_QUEUE_ENABLED", true),
		MaxPerRunner:  getEnvInt("PENDING_QUEUE_MAX_PER_RUNNER", 20),
		DefaultTTL:    getEnvDuration("PENDING_QUEUE_DEFAULT_TTL", 30*time.Minute),
		SweepInterval: getEnvDuration("PENDING_QUEUE_SWEEP_INTERVAL", 60*time.Second),
	}
}
