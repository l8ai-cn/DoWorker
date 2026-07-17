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
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPersistAndQueueWritesItemAndIdempotentCommand(t *testing.T) {
	db := newPromptOutboxDB(t)
	queue := newPromptTestQueue(db)
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
	require.NotContains(t, string(command.Payload), "hello")
	plaintext, err := promptCommandPayload("pod_1", item.ID, "hello")
	require.NoError(t, err)
	matches, err := queue.PayloadMatches(command.Payload, plaintext)
	require.NoError(t, err)
	require.True(t, matches)

	retry, err := UserItem("item_1", "session_1", "response_1", []map[string]string{
		{"type": "input_text", "text": "hello"},
	})
	require.NoError(t, err)
	require.NoError(t, NewPromptOutbox(db, queue).PersistAndQueue(
		context.Background(),
		PromptInput{
			OrganizationID: 1, RunnerID: 2, PodKey: "pod_1",
			Item: retry, Prompt: "hello",
		},
	))
	var itemCount, commandCount int64
	require.NoError(t, db.Model(&domainitem.Item{}).Count(&itemCount).Error)
	require.NoError(t, db.Model(&agentpod.PendingCommand{}).Count(&commandCount).Error)
	require.Equal(t, int64(1), itemCount)
	require.Equal(t, int64(1), commandCount)
}

func TestPersistAndQueueRollsBackItemWhenCommandConflicts(t *testing.T) {
	db := newPromptOutboxDB(t)
	queue := newPromptTestQueue(db)
	conflictPayload, err := queue.SealPayload([]byte("existing"))
	require.NoError(t, err)
	require.NoError(t, db.Create(&agentpod.PendingCommand{
		OrganizationID: 1, RunnerID: 2, PodKey: "pod_1",
		CommandType: agentpod.CommandTypeSendPrompt, CommandID: "item_1",
		Payload: conflictPayload, ExpiresAt: time.Now().Add(time.Minute),
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

func TestPersistAndQueueRejectsChangedReplay(t *testing.T) {
	db := newPromptOutboxDB(t)
	queue := newPromptTestQueue(db)
	first, err := UserItem("item_1", "session_1", "response_1", json.RawMessage(`[]`))
	require.NoError(t, err)
	outbox := NewPromptOutbox(db, queue)
	require.NoError(t, outbox.PersistAndQueue(context.Background(), PromptInput{
		OrganizationID: 1, RunnerID: 2, PodKey: "pod_1",
		Item: first, Prompt: "hello",
	}))
	changed, err := UserItem("item_1", "session_1", "response_1", json.RawMessage(`[]`))
	require.NoError(t, err)

	err = outbox.PersistAndQueue(context.Background(), PromptInput{
		OrganizationID: 1, RunnerID: 2, PodKey: "pod_1",
		Item: changed, Prompt: "different",
	})
	require.Error(t, err)
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

func newPromptTestQueue(db *gorm.DB) *runnerservice.PendingCommandQueue {
	return runnerservice.NewPendingCommandQueue(
		infra.NewPendingCommandRepository(db),
		nil,
		10,
		0,
		true,
		crypto.NewEncryptor("prompt-outbox-test-key"),
		nil,
	)
}
