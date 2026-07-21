package client

import (
	"testing"
	"time"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendMessage_WorkbenchUsesIndependentReliableQueue(t *testing.T) {
	conn := newTestConnection()
	setFakeStream(conn)

	for i := 0; i < cap(conn.controlCh); i++ {
		conn.controlCh <- &runnerv1.RunnerMessage{}
	}

	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_WorkbenchEvents{},
	}
	require.NoError(t, conn.SendMessage(msg))
	assert.Same(t, msg, <-conn.workbenchCh)
}

func TestSendMessage_WorkbenchBackpressuresUntilQueueDrains(t *testing.T) {
	conn := newTestConnection()
	setFakeStream(conn)

	for i := 0; i < cap(conn.workbenchCh); i++ {
		conn.workbenchCh <- &runnerv1.RunnerMessage{}
	}

	result := make(chan error, 1)
	go func() {
		result <- conn.SendMessage(&runnerv1.RunnerMessage{
			Payload: &runnerv1.RunnerMessage_WorkbenchEvents{},
		})
	}()

	select {
	case err := <-result:
		t.Fatalf("reliable send returned before queue drained: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	<-conn.workbenchCh
	require.NoError(t, <-result)
	assert.Equal(t, cap(conn.workbenchCh), len(conn.workbenchCh))
}
