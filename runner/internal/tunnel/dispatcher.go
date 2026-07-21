package tunnel

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/l8ai-cn/agentcloud/runner/internal/safego"
	"github.com/l8ai-cn/agentcloud/runner/internal/tunnelframe"
)

const defaultLocalWindow = 1 << 20

// localStream tracks one in-flight tunneled request being served locally.
type localStream struct {
	id      uint32
	bodyW   *io.PipeWriter
	wsIn    chan tunnelframe.Frame // set instead of bodyW for WebSocket requests
	done    chan struct{}         // closed when the goroutine serving this stream exits
	sendWin *creditWindow
	cancel  context.CancelFunc
}

// localDispatcher implements Dispatcher: it maps inbound REQ_* frames to local
// HTTP requests against the pod's loopback services.
type localDispatcher struct {
	ctx    context.Context
	window int

	mu      sync.Mutex
	streams map[uint32]*localStream
	send    func(tunnelframe.Frame) error
}

// NewDispatcher creates a dispatcher rooted at ctx. window bounds per-stream
// in-flight response bytes.
func NewDispatcher(ctx context.Context, window int) Dispatcher {
	if ctx == nil {
		ctx = context.Background()
	}
	if window <= 0 {
		window = defaultLocalWindow
	}
	return &localDispatcher{
		ctx:     ctx,
		window:  window,
		streams: make(map[uint32]*localStream),
	}
}

func (d *localDispatcher) SetSender(send func(tunnelframe.Frame) error) {
	d.mu.Lock()
	d.send = send
	d.mu.Unlock()
}

func (d *localDispatcher) sender() func(tunnelframe.Frame) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.send
}

func (d *localDispatcher) Dispatch(f tunnelframe.Frame) {
	switch f.Type {
	case tunnelframe.TypeReqStart:
		d.handleReqStart(f)
	case tunnelframe.TypeReqBody:
		if ls := d.get(f.StreamID); ls != nil && ls.bodyW != nil {
			_, _ = ls.bodyW.Write(f.Payload)
		}
	case tunnelframe.TypeReqEnd:
		if ls := d.get(f.StreamID); ls != nil && ls.bodyW != nil {
			_ = ls.bodyW.Close()
		}
	case tunnelframe.TypeWSData, tunnelframe.TypeWSClose:
		if ls := d.get(f.StreamID); ls != nil && ls.wsIn != nil {
			// Block until delivered or the stream's goroutine has exited (done
			// closed) — mirrors bodyW's io.Pipe backpressure for HTTP bodies so
			// WS frames are never silently dropped while the stream is alive.
			select {
			case ls.wsIn <- f:
			case <-ls.done:
			}
		}
	case tunnelframe.TypeCredit:
		if ls := d.get(f.StreamID); ls != nil {
			var c tunnelframe.CreditPayload
			if json.Unmarshal(f.Payload, &c) == nil && c.Bytes > 0 {
				ls.sendWin.add(c.Bytes)
			}
		}
	case tunnelframe.TypeStreamCancel:
		d.cancelStream(f.StreamID)
	}
}

func (d *localDispatcher) handleReqStart(f tunnelframe.Frame) {
	var p tunnelframe.ReqStartPayload
	if err := json.Unmarshal(f.Payload, &p); err != nil {
		if send := d.sender(); send != nil {
			_ = send(respError(f.StreamID, "bad_request", "invalid REQ_START"))
		}
		return
	}

	sctx, cancel := context.WithCancel(d.ctx)
	sw := newCreditWindow(d.window)

	var body io.Reader
	var bodyW *io.PipeWriter
	var wsIn chan tunnelframe.Frame
	if p.IsWebSocket {
		wsIn = make(chan tunnelframe.Frame, 32)
	} else if requestMayHaveBody(p) {
		pr, pw := io.Pipe()
		body = pr
		bodyW = pw
	}

	ls := &localStream{id: f.StreamID, bodyW: bodyW, wsIn: wsIn, done: make(chan struct{}), sendWin: sw, cancel: cancel}
	d.mu.Lock()
	d.streams[f.StreamID] = ls
	send := d.send
	d.mu.Unlock()

	if send == nil {
		cancel()
		close(ls.done)
		return
	}

	safego.Go("tunnel-local-http", func() {
		defer close(ls.done)
		defer d.remove(f.StreamID)
		defer cancel()
		if p.IsWebSocket {
			serveLocalWebSocket(sctx, sendFunc(send), f.StreamID, p, wsIn, sw)
		} else {
			serveLocalHTTP(sctx, sendFunc(send), f.StreamID, p, body, sw)
		}
	})
}

func requestMayHaveBody(p tunnelframe.ReqStartPayload) bool {
	if p.ContentLength > 0 {
		return true
	}
	switch strings.ToUpper(p.Method) {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func (d *localDispatcher) get(id uint32) *localStream {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.streams[id]
}

func (d *localDispatcher) remove(id uint32) {
	d.mu.Lock()
	ls := d.streams[id]
	delete(d.streams, id)
	d.mu.Unlock()
	if ls != nil {
		ls.sendWin.close()
	}
}

func (d *localDispatcher) cancelStream(id uint32) {
	d.mu.Lock()
	ls := d.streams[id]
	delete(d.streams, id)
	d.mu.Unlock()
	if ls != nil {
		ls.cancel()
		ls.sendWin.close()
		if ls.bodyW != nil {
			_ = ls.bodyW.CloseWithError(io.ErrClosedPipe)
		}
	}
}

func (d *localDispatcher) Close() {
	d.mu.Lock()
	streams := d.streams
	d.streams = make(map[uint32]*localStream)
	d.mu.Unlock()
	for _, ls := range streams {
		ls.cancel()
		ls.sendWin.close()
		if ls.bodyW != nil {
			_ = ls.bodyW.CloseWithError(io.ErrClosedPipe)
		}
	}
}

// sendFunc adapts a send closure to the frameSink interface.
type sendFunc func(tunnelframe.Frame) error

func (f sendFunc) Send(fr tunnelframe.Frame) error { return f(fr) }
