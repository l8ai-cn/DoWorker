package runner

import (
	"context"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/pkg/crypto"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestDrain_RejectsUnreadablePayloadWithoutSendingOrDeleting(t *testing.T) {
	validPlaintext := pendingCreateMessage(t, "pd-secure")
	validEnvelope := encryptPendingPayload(t, testPendingEncryptor(), validPlaintext)
	tamperedEnvelope := append([]byte(nil), validEnvelope...)
	tamperedEnvelope[len(tamperedEnvelope)-1] ^= 1

	tests := map[string][]byte{
		"plaintext": []byte(agentpod.PendingPayloadPrefix + "not-an-envelope"),
		"tampered":  tamperedEnvelope,
		"wrong key": encryptPendingPayload(t, crypto.NewEncryptor("wrong-key"), validPlaintext),
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			assertRejectedPendingPayload(t, payload, "pd-secure")
		})
	}
}

func TestDrain_RejectsPayloadWhoseMetadataDoesNotMatchRow(t *testing.T) {
	payload := encryptPendingPayload(
		t,
		testPendingEncryptor(),
		pendingCreateMessage(t, "pd-other"),
	)
	assertRejectedPendingPayload(t, payload, "pd-secure")
}

func assertRejectedPendingPayload(t *testing.T, payload []byte, podKey string) {
	t.Helper()
	const runnerID = int64(7)
	repo := &memPendingRepo{rows: []*agentpod.PendingCommand{{
		ID: 1, RunnerID: runnerID, PodKey: podKey,
		CommandType: agentpod.CommandTypeCreatePod, CommandID: podKey,
		Payload: payload, ExpiresAt: time.Now().Add(time.Hour),
	}}}
	sender := &recordingMsgSender{}
	drainer := NewPendingCommandDrainer(
		repo,
		nil,
		nil,
		sender,
		&stubConnChecker{connected: true},
		nil,
		nil,
		nil,
		time.Minute,
		testPendingEncryptor(),
		newTestLogger(),
	)

	drainer.drainRunner(context.Background(), runnerID)

	assert.Zero(t, sender.calls.Load())
	assert.Len(t, mustListPending(repo, runnerID), 1)
}

func pendingCreateMessage(t *testing.T, podKey string) []byte {
	t.Helper()
	payload, err := proto.Marshal(&runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_CreatePod{
			CreatePod: &runnerv1.CreatePodCommand{PodKey: podKey},
		},
	})
	require.NoError(t, err)
	return payload
}

func encryptPendingPayload(t *testing.T, encryptor *crypto.Encryptor, payload []byte) []byte {
	t.Helper()
	envelope, err := newPendingPayloadCipher(encryptor).encrypt(payload)
	require.NoError(t, err)
	return envelope
}
