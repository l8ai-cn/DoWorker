package infra

import (
	"testing"
	"time"

	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func encryptedWorkerLaunchPayload(t *testing.T, podKey string) []byte {
	t.Helper()
	plaintext, err := proto.Marshal(&runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_CreatePod{
			CreatePod: &runnerv1.CreatePodCommand{PodKey: podKey},
		},
	})
	require.NoError(t, err)
	queue := runnerservice.NewPendingCommandQueue(
		nil,
		nil,
		1,
		time.Minute,
		true,
		crypto.NewEncryptor("worker-launch-payload-test"),
		nil,
	)
	envelope, err := queue.SealPayload(plaintext)
	require.NoError(t, err)
	return envelope
}
