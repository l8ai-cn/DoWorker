package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

// applyEnvCompat copies DO_WORKER_* values into viper when set, and mirrors
// AGENTCLOUD_* into DO_WORKER_* when the newer name is unset.
func applyEnvCompat(v *viper.Viper) {
	for _, env := range os.Environ() {
		key, val, ok := strings.Cut(env, "=")
		if !ok || val == "" {
			continue
		}
		switch {
		case strings.HasPrefix(key, "DO_WORKER_"):
			v.Set(strings.ToLower(strings.TrimPrefix(key, "DO_WORKER_")), val)
		case strings.HasPrefix(key, "AGENTCLOUD_"):
			newKey := strings.ToLower(strings.TrimPrefix(key, "AGENTCLOUD_"))
			if os.Getenv("DO_WORKER_"+strings.TrimPrefix(key, "AGENTCLOUD_")) == "" {
				v.Set(newKey, val)
			}
		}
	}
}
