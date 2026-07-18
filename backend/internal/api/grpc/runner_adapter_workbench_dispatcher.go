package grpc

import (
	"context"
	"sync"

	"github.com/anthropics/agentsmesh/backend/internal/service/runner"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

type runnerWorkbenchDispatcher struct {
	adapter  *GRPCRunnerAdapter
	runnerID int64
	conn     *runner.GRPCConnection
	ctx      context.Context
	cancel   context.CancelFunc
	wake     chan struct{}
	done     chan struct{}
	mu       sync.Mutex
	queue    []*runnerv1.RunnerMessage
}

func newRunnerWorkbenchDispatcher(
	ctx context.Context,
	adapter *GRPCRunnerAdapter,
	runnerID int64,
	conn *runner.GRPCConnection,
) *runnerWorkbenchDispatcher {
	dispatchCtx, cancel := context.WithCancel(ctx)
	dispatcher := &runnerWorkbenchDispatcher{
		adapter: adapter, runnerID: runnerID, conn: conn,
		ctx: dispatchCtx, cancel: cancel,
		wake: make(chan struct{}, 1), done: make(chan struct{}),
	}
	go dispatcher.run()
	return dispatcher
}

func (dispatcher *runnerWorkbenchDispatcher) enqueue(
	message *runnerv1.RunnerMessage,
) {
	dispatcher.mu.Lock()
	dispatcher.queue = append(dispatcher.queue, message)
	dispatcher.mu.Unlock()
	select {
	case dispatcher.wake <- struct{}{}:
	default:
	}
}

func (dispatcher *runnerWorkbenchDispatcher) stop() {
	dispatcher.cancel()
	<-dispatcher.done
}

func (dispatcher *runnerWorkbenchDispatcher) run() {
	defer close(dispatcher.done)
	for {
		if dispatcher.ctx.Err() != nil {
			return
		}
		if message := dispatcher.next(); message != nil {
			dispatcher.adapter.handleProtoMessage(
				dispatcher.ctx,
				dispatcher.runnerID,
				dispatcher.conn,
				message,
			)
			continue
		}
		select {
		case <-dispatcher.ctx.Done():
			return
		case <-dispatcher.wake:
		}
	}
}

func (dispatcher *runnerWorkbenchDispatcher) next() *runnerv1.RunnerMessage {
	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()
	if len(dispatcher.queue) == 0 {
		return nil
	}
	message := dispatcher.queue[0]
	dispatcher.queue[0] = nil
	dispatcher.queue = dispatcher.queue[1:]
	return message
}
