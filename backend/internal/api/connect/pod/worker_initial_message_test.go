package podconnect

import (
	"context"
	"encoding/json"
	"testing"

	sessiondomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	itemdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/conversationitem"
	"github.com/stretchr/testify/require"
)

type recordingConversationItemWriter struct {
	position int64
	item     *itemdomain.Item
}

func (w *recordingConversationItemWriter) NextPosition(context.Context, string) (int64, error) {
	return w.position, nil
}

func (w *recordingConversationItemWriter) Append(
	_ context.Context,
	item *itemdomain.Item,
) error {
	w.item = item
	return nil
}

func TestPrepareWorkerInitialMessagePersistsUserTask(t *testing.T) {
	writer := &recordingConversationItemWriter{position: 1}
	server := NewServer(nil, nil, WithConversationItems(writer))
	prepare := server.prepareWorkerInitialMessage("  Check credentials.  ")

	require.NoError(t, prepare(context.Background(), &sessiondomain.Session{
		ID: "conv_seedance",
	}))
	require.NotNil(t, writer.item)
	require.Equal(t, "conv_seedance", writer.item.SessionID)
	require.Equal(t, "message", writer.item.ItemType)
	require.Equal(t, int64(1), writer.item.Position)
	var payload struct {
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	require.NoError(t, json.Unmarshal(writer.item.Payload, &payload))
	require.Equal(t, "user", payload.Role)
	require.Equal(t, "input_text", payload.Content[0].Type)
	require.Equal(t, "Check credentials.", payload.Content[0].Text)
}

func TestPrepareWorkerInitialMessageRequiresWriter(t *testing.T) {
	prepare := NewServer(nil, nil).prepareWorkerInitialMessage("Check credentials.")

	err := prepare(context.Background(), &sessiondomain.Session{ID: "conv_seedance"})

	require.EqualError(t, err, "conversation item service unavailable")
}

func TestPrepareWorkerInitialMessageSkipsEmptyTask(t *testing.T) {
	require.Nil(t, NewServer(nil, nil).prepareWorkerInitialMessage("  "))
}
