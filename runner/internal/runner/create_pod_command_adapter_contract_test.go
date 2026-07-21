package runner

import (
	"testing"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestCreatePodCommandRequiresExplicitAdapterID(t *testing.T) {
	field := (&runnerv1.CreatePodCommand{}).ProtoReflect().
		Descriptor().Fields().ByName("adapter_id")

	require.NotNil(t, field, "CreatePodCommand.adapter_id is required")
	require.Equal(t, protoreflect.StringKind, field.Kind())
	require.Equal(t, protoreflect.FieldNumber(26), field.Number())
}
