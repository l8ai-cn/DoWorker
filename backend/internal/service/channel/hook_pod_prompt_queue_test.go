package channel

import (
	"context"
	"errors"
	"strings"
	"testing"

	channelDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/channel"
)

type mockPromptQueuer struct {
	queued []promptRecord
	err    error
}

func (m *mockPromptQueuer) QueueChannelPrompt(_ context.Context, podKey, prompt string, _ int64) error {
	if m.err != nil {
		return m.err
	}
	m.queued = append(m.queued, promptRecord{podKey: podKey, prompt: prompt})
	return nil
}

func TestPodPromptHook_OfflinePod_QueuesPrompt(t *testing.T) {
	router := &mockPodRouter{failKeys: map[string]bool{"offline-pod": true}}
	writer := &mockSystemWriter{}
	queuer := &mockPromptQueuer{}
	hook := NewPodPromptHook(router, writer, queuer)

	mc := &MessageContext{
		Channel:  &channelDomain.Channel{ID: 1, Name: "test"},
		Message:  &channelDomain.Message{ID: 42, Body: "hi", ChannelID: 1},
		Mentions: &MentionResult{PodKeys: []string{"offline-pod"}},
	}

	if err := hook(context.Background(), mc); err != nil {
		t.Fatalf("hook failed: %v", err)
	}

	if len(queuer.queued) != 1 || queuer.queued[0].podKey != "offline-pod" {
		t.Fatalf("prompt should be queued for offline pod, got %v", queuer.queued)
	}
	if len(writer.messages) != 1 {
		t.Fatalf("expected 1 system notice, got %d", len(writer.messages))
	}
	if !strings.Contains(writer.messages[0].Body, "queued") {
		t.Errorf("notice should mention queueing, got %q", writer.messages[0].Body)
	}
}

func TestPodPromptHook_QueueFails_FallsBackToOfflineNotice(t *testing.T) {
	router := &mockPodRouter{failKeys: map[string]bool{"offline-pod": true}}
	writer := &mockSystemWriter{}
	queuer := &mockPromptQueuer{err: errors.New("queue full")}
	hook := NewPodPromptHook(router, writer, queuer)

	mc := &MessageContext{
		Channel:  &channelDomain.Channel{ID: 1, Name: "test"},
		Message:  &channelDomain.Message{ID: 43, Body: "hi", ChannelID: 1},
		Mentions: &MentionResult{PodKeys: []string{"offline-pod"}},
	}

	if err := hook(context.Background(), mc); err != nil {
		t.Fatalf("hook failed: %v", err)
	}

	if len(writer.messages) != 1 {
		t.Fatalf("expected 1 offline notice, got %d", len(writer.messages))
	}
	if !strings.Contains(writer.messages[0].Body, "cannot receive") {
		t.Errorf("fallback notice should be the offline message, got %q", writer.messages[0].Body)
	}
}

func TestPodPromptHook_OnlinePod_NotQueued(t *testing.T) {
	router := &mockPodRouter{}
	queuer := &mockPromptQueuer{}
	hook := NewPodPromptHook(router, &mockSystemWriter{}, queuer)

	mc := &MessageContext{
		Channel:  &channelDomain.Channel{ID: 1, Name: "test"},
		Message:  &channelDomain.Message{ID: 44, Body: "hi", ChannelID: 1},
		Mentions: &MentionResult{PodKeys: []string{"online-pod"}},
	}

	if err := hook(context.Background(), mc); err != nil {
		t.Fatalf("hook failed: %v", err)
	}

	if len(router.prompts) != 1 {
		t.Fatalf("online pod should receive the prompt directly, got %v", router.prompts)
	}
	if len(queuer.queued) != 0 {
		t.Fatalf("online pod must not enter the queue, got %v", queuer.queued)
	}
}
