package goalloopconnect

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	goalloopv1 "github.com/anthropics/agentsmesh/proto/gen/go/goalloop/v1"
)

func TestCreateGoalLoopRequiresResourceApply(t *testing.T) {
	server := NewServer(nil, loopOrgService{})
	_, err := server.CreateGoalLoop(
		loopContext(),
		connect.NewRequest(&goalloopv1.CreateGoalLoopRequest{
			OrgSlug: "acme",
			Name:    "legacy",
		}),
	)

	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Contains(t, err.Error(), "validate-plan-apply")
}
