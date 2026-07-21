package runner

import (
	"testing"
	"time"

	runnerDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func storeActiveRunner(
	t *testing.T,
	db *gorm.DB,
	service *Service,
	r *runnerDomain.Runner,
	lastPing time.Time,
	podCount int,
) {
	t.Helper()
	enabled := r.IsEnabled
	r.CurrentPods = podCount
	require.NoError(t, db.Save(r).Error)
	if !enabled {
		require.NoError(t, db.Model(r).UpdateColumn("is_enabled", false).Error)
		r.IsEnabled = false
	}
	service.activeRunners.Store(r.ID, &ActiveRunner{
		Runner:   r,
		LastPing: lastPing,
		PodCount: podCount,
	})
}
