package sessionmessage

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domainitem "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPersistAndQueueWritesItemAndIdempotentCommand(t *testing.T) {
	db := newPromptOutboxDB(t)
	queue := runnerservice.NewPendingCommandQueue(
		infra.NewPendingCommandRepository(db), nil, 10, 0, true, nil,
	)
	item, err := UserItem("item_1", "session_1", "response_1", []map[string]string{
		{"type": "input_text", "text": "hello"},
	})
	require.NoError(t, err)

	require.NoError(t, NewPromptOutbox(db, queue).PersistAndQueue(context.Background(), PromptInput{
		OrganizationID: 1, RunnerID: 2, PodKey: "pod_1", Item: item, Prompt: "hello",
	}))

	var stored domainitem.Item
	require.NoError(t, db.First(&stored, "id = ?", item.ID).Error)
	require.Equal(t, int64(1), stored.Position)
	var command agentpod.PendingCommand
	require.NoError(t, db.First(&command, "command_id = ?", item.ID).Error)
	var message runnerv1.ServerMessage
	require.NoError(t, proto.Unmarshal(command.Payload, &message))
	require.Equal(t, item.ID, message.GetSendPrompt().GetCommandId())
	require.Equal(t, "hello", message.GetSendPrompt().GetPrompt())
}

func TestPersistAndQueueRollsBackItemWhenCommandConflicts(t *testing.T) {
	db := newPromptOutboxDB(t)
	queue := runnerservice.NewPendingCommandQueue(
		infra.NewPendingCommandRepository(db), nil, 10, 0, true, nil,
	)
	require.NoError(t, db.Create(&agentpod.PendingCommand{
		OrganizationID: 1, RunnerID: 2, PodKey: "pod_1",
		CommandType: agentpod.CommandTypeSendPrompt, CommandID: "item_1",
		Payload: []byte("existing"), ExpiresAt: time.Now().Add(time.Minute),
	}).Error)
	item, err := UserItem("item_1", "session_1", "response_1", json.RawMessage(`[]`))
	require.NoError(t, err)

	err = NewPromptOutbox(db, queue).PersistAndQueue(context.Background(), PromptInput{
		OrganizationID: 1, RunnerID: 2, PodKey: "pod_1", Item: item, Prompt: "hello",
	})
	require.Error(t, err)
	var count int64
	require.NoError(t, db.Model(&domainitem.Item{}).Count(&count).Error)
	require.Zero(t, count)
}

func TestPromptLockKeyIsStableAndNamespaced(t *testing.T) {
	require.Equal(t, promptLockKey("runner", "7"), promptLockKey("runner", "7"))
	require.NotEqual(t, promptLockKey("runner", "7"), promptLockKey("session", "7"))
	require.NotEqual(t, promptLockKey("session", "session_1"), promptLockKey("session", "session_2"))
}

func newPromptOutboxDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/prompt-outbox.db"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domainitem.Item{}))
	require.NoError(t, db.Exec(`
		CREATE TABLE pending_runner_commands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id INTEGER NOT NULL,
			runner_id INTEGER NOT NULL,
			pod_key TEXT NOT NULL,
			command_type TEXT NOT NULL,
			command_id TEXT NOT NULL UNIQUE,
			payload BLOB NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error)
	return db
}
