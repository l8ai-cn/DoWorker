package tunnel

import (
	"context"
	"sync"

	"github.com/anthropics/agentsmesh/relay/internal/protocol/tunnelframe"
)

// creditWindow implements per-stream, per-direction flow control. A sender
// blocks in acquire when the window is exhausted; the receiver replenishes it
// via add after data has been flushed to its destination. This bounds in-flight
// memory to the window size regardless of the total transferred size.
type creditWindow struct {
	mu     sync.Mutex
	cond   *sync.Cond
	avail  int
	closed bool
}

func newCreditWindow(initial int) *creditWindow {
	w := &creditWindow{avail: initial}
	w.cond = sync.NewCond(&w.mu)
	return w
}

// acquire blocks until n credits are available or ctx is cancelled / window closed.
func (w *creditWindow) acquire(ctx context.Context, n int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for w.avail < n && !w.closed {
		if err := ctx.Err(); err != nil {
			return err
		}
		// Bridge ctx cancellation into the cond so a cancelled waiter wakes up.
		done := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				w.cond.Broadcast()
			case <-done:
			}
		}()
		w.cond.Wait()
		close(done)
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if w.closed {
		return context.Canceled
	}
	w.avail -= n
	return nil
}

// add replenishes the window by n credits and wakes any waiters.
func (w *creditWindow) add(n int) {
	w.mu.Lock()
	w.avail += n
	w.mu.Unlock()
	w.cond.Broadcast()
}

// close permanently unblocks all waiters (used on stream teardown).
func (w *creditWindow) close() {
	w.mu.Lock()
	w.closed = true
	w.mu.Unlock()
	w.cond.Broadcast()
}

// Stream is a single multiplexed logical stream over a tunnel connection.
// respCh carries control/response frames (RESP_START/BODY/END/ERROR/CREDIT and
// WS_DATA/WS_CLOSE) destined for the proxy layer that owns this stream.
type Stream struct {
	ID      uint32
	sendWin *creditWindow // credits for our outbound (REQ) body
	recvWin *creditWindow // credits granted to peer for inbound (RESP) body
	respCh  chan tunnelframe.Frame
	cancel  func()

	closeOnce sync.Once
}

func newStream(id uint32, window int) *Stream {
	return &Stream{
		ID:      id,
		sendWin: newCreditWindow(window),
		recvWin: newCreditWindow(window),
		respCh:  make(chan tunnelframe.Frame, 64),
	}
}

// NewStream creates a standalone stream. Used by proxy/test code that manages
// its own stream table (the Tunnel uses the internal newStream).
func NewStream(id uint32, window int) *Stream { return newStream(id, window) }

// RespChan exposes the inbound frame channel for the proxy layer to consume
// RESP_*/WS_*/CREDIT frames destined for this stream.
func (s *Stream) RespChan() <-chan tunnelframe.Frame { return s.respCh }

// Deliver pushes an inbound frame onto this stream. Used by the tunnel read
// loop and by tests injecting simulated peer frames.
func (s *Stream) Deliver(f tunnelframe.Frame) { s.respCh <- f }

// AcquireSend blocks until n outbound (REQ/WS) body credits are available or
// the context is cancelled / the stream closed.
func (s *Stream) AcquireSend(ctx context.Context, n int) error { return s.sendWin.acquire(ctx, n) }

// AddSendCredit replenishes n outbound body credits (on receiving a CREDIT frame).
func (s *Stream) AddSendCredit(n int) { s.sendWin.add(n) }

// SetCancel registers a cancel callback fired when the stream is torn down.
func (s *Stream) SetCancel(fn func()) { s.cancel = fn }

// closeStream releases credit windows so blocked senders/receivers unwind.
func (s *Stream) closeStream() {
	s.closeOnce.Do(func() {
		s.sendWin.close()
		s.recvWin.close()
		if s.cancel != nil {
			s.cancel()
		}
	})
}
