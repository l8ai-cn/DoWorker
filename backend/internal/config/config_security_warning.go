package config

import "log/slog"

func (c *Config) WarnInsecureDefaults() {
	if c.Server.InternalAPISecret == "change-me-internal-secret" {
		slog.Warn("SECURITY: INTERNAL_API_SECRET is using the default value; set a strong random secret via environment variable")
	}
	if c.JWT.Secret != "change-me-in-production" {
		return
	}
	if c.Server.Debug {
		slog.Warn("SECURITY: JWT_SECRET is using the default value; set a strong random secret via environment variable")
		return
	}
	slog.Error("SECURITY: JWT_SECRET is using the default value in non-debug mode; this is a critical security risk — set JWT_SECRET environment variable")
}
