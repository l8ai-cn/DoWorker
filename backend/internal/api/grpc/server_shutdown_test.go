package grpc

import (
	"sync"
	"testing"
	"time"
)

type immediateGRPCStopper struct {
	mu         sync.Mutex
	stopCalled bool
}

func (s *immediateGRPCStopper) GracefulStop() {}

func (s *immediateGRPCStopper) Stop() {
	s.mu.Lock()
	s.stopCalled = true
	s.mu.Unlock()
}

func (s *immediateGRPCStopper) wasForced() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopCalled
}

type blockingGRPCStopper struct {
	gracefulStarted chan struct{}
	stopCalled      chan struct{}
	release         chan struct{}
}

func newBlockingGRPCStopper() *blockingGRPCStopper {
	return &blockingGRPCStopper{
		gracefulStarted: make(chan struct{}),
		stopCalled:      make(chan struct{}),
		release:         make(chan struct{}),
	}
}

func (s *blockingGRPCStopper) GracefulStop() {
	close(s.gracefulStarted)
	<-s.release
}

func (s *blockingGRPCStopper) Stop() {
	close(s.stopCalled)
	close(s.release)
}

func TestStopGRPCServerCompletesGracefully(t *testing.T) {
	server := &immediateGRPCStopper{}

	graceful := stopGRPCServer(server, time.Second)

	if !graceful {
		t.Fatal("expected graceful shutdown")
	}
	if server.wasForced() {
		t.Fatal("did not expect forced shutdown")
	}
}

func TestStopGRPCServerForcesBlockedConnectionsAfterTimeout(t *testing.T) {
	server := newBlockingGRPCStopper()

	graceful := stopGRPCServer(server, 10*time.Millisecond)

	if graceful {
		t.Fatal("expected forced shutdown")
	}
	select {
	case <-server.gracefulStarted:
	default:
		t.Fatal("expected graceful shutdown attempt")
	}
	select {
	case <-server.stopCalled:
	default:
		t.Fatal("expected forced shutdown after timeout")
	}
}
