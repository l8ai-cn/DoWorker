package sessionapi

import (
	"context"
	"strings"
	"testing"

	domainitem "github.com/l8ai-cn/agentcloud/backend/internal/domain/conversationitem"
	itemsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/conversationitem"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHandleAcpLogPublishesCanonicalErrorItemOnce(t *testing.T) {
	publisher, hub, db := newErrorStreamPublisher(t)
	frames := hub.Subscribe("conv-error")
	defer hub.Unsubscribe("conv-error", frames)

	publisher.handleAcpLog(
		context.Background(),
		"conv-error",
		`{"level":"error","message":"provider failed"}`,
	)

	published := drainSessionFrames(frames)
	require.Equal(t, 1, countFrames(published, "event: turn.item.done"))
	require.Equal(t, 0, countFrames(published, "event: response.error"))

	var count int64
	require.NoError(t, db.Model(&domainitem.Item{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestHandleAcpLogProjectsReconnectAsTransientRetry(t *testing.T) {
	publisher, hub, db := newErrorStreamPublisher(t)
	frames := hub.Subscribe("conv-retry")
	defer hub.Unsubscribe("conv-retry", frames)

	publisher.handleAcpLog(
		context.Background(),
		"conv-retry",
		`{"level":"error","message":"Reconnecting... 2/5"}`,
	)

	published := drainSessionFrames(frames)
	require.Equal(t, 1, countFrames(published, "event: response.retry"))
	require.Equal(t, 0, countFrames(published, "event: turn.item.done"))

	var count int64
	require.NoError(t, db.Model(&domainitem.Item{}).Count(&count).Error)
	require.Zero(t, count)
}

func newErrorStreamPublisher(
	t *testing.T,
) (*SessionStreamPublisher, *SessionHub, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/errors.db"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domainitem.Item{}))
	hub := NewSessionHub()
	return NewSessionStreamPublisher(hub, itemsvc.NewService(db), nil, nil), hub, db
}

func drainSessionFrames(ch chan string) []string {
	var frames []string
	for {
		select {
		case frame := <-ch:
			frames = append(frames, frame)
		default:
			return frames
		}
	}
}

func countFrames(frames []string, event string) int {
	count := 0
	for _, frame := range frames {
		if strings.Contains(frame, event) {
			count++
		}
	}
	return count
}
