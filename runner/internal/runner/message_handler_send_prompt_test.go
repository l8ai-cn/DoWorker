package runner

import (
	"sync"
	"testing"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

// sendPromptMockIO records SendInput and SendKeys calls (with timestamps so
// tests can assert ordering and the post-text submission gap).
type sendPromptMockIO struct {
	stubPodIOZero
	mu          sync.Mutex
	mode        string
	inputs      []timedCall
	keys        []timedCall
	inputErr    error
	keysErr     error
	agentStatus string
	subscribers map[string]func(string)
}

type timedCall struct {
	payload string
	at      time.Time
}

func (m *sendPromptMockIO) Mode() string { return m.mode }

func (m *sendPromptMockIO) GetAgentStatus() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.agentStatus == "" {
		return "waiting"
	}
	return m.agentStatus
}

func (m *sendPromptMockIO) SubscribeStateChange(id string, cb func(string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.subscribers == nil {
		m.subscribers = make(map[string]func(string))
	}
	m.subscribers[id] = cb
}

func (m *sendPromptMockIO) UnsubscribeStateChange(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subscribers, id)
}

func (m *sendPromptMockIO) setAgentStatus(status string) {
	m.mu.Lock()
	m.agentStatus = status
	callbacks := make([]func(string), 0, len(m.subscribers))
	for _, cb := range m.subscribers {
		callbacks = append(callbacks, cb)
	}
	m.mu.Unlock()
	for _, cb := range callbacks {
		cb(status)
	}
}

func (m *sendPromptMockIO) SendInput(text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inputs = append(m.inputs, timedCall{payload: text, at: time.Now()})
	return m.inputErr
}

// ptyTerminalMock satisfies TerminalAccess on top of sendPromptMockIO so the
// PTY branch of OnSendPrompt resolves through SendKeys (the "press Enter"
// path) instead of SendInput (the "raw bytes" path).
type ptyTerminalMock struct {
	*sendPromptMockIO
}

func (p *ptyTerminalMock) SendKeys(keys []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, k := range keys {
		p.keys = append(p.keys, timedCall{payload: k, at: time.Now()})
	}
	return p.keysErr
}

func (p *ptyTerminalMock) Resize(int, int) (bool, error) { return true, nil }
func (p *ptyTerminalMock) CursorPosition() (int, int)    { return 0, 0 }
func (p *ptyTerminalMock) GetScreenSnapshot() string     { return "" }
func (p *ptyTerminalMock) Redraw() error                 { return nil }
func (p *ptyTerminalMock) WriteOutput([]byte)            {}

// stubPodIOZero satisfies PodIO with no-ops so each test mock only overrides
// the methods it cares about.
type stubPodIOZero struct{}

func (stubPodIOZero) GetSnapshot(int) (string, error)           { return "", nil }
func (stubPodIOZero) GetAgentStatus() string                    { return "idle" }
func (stubPodIOZero) SubscribeStateChange(string, func(string)) {}
func (stubPodIOZero) UnsubscribeStateChange(string)             {}
func (stubPodIOZero) GetPID() int                               { return 0 }
func (stubPodIOZero) Start() error                              { return nil }
func (stubPodIOZero) Stop()                                     {}
func (stubPodIOZero) Teardown() string                          { return "" }
func (stubPodIOZero) SetExitHandler(func(int))                  {}
func (stubPodIOZero) SetIOErrorHandler(func(error))             {}
func (stubPodIOZero) Detach()                                   {}

func TestOnSendPrompt_PTY_PressesEnterViaSendKeys(t *testing.T) {
	base := &sendPromptMockIO{mode: InteractionModePTY}
	io := &ptyTerminalMock{sendPromptMockIO: base}
	pod := &Pod{PodKey: "pty-pod", InteractionMode: InteractionModePTY, IO: io}

	store := NewInMemoryPodStore()
	store.Put(pod.PodKey, pod)
	h := &RunnerMessageHandler{podStore: store}

	if err := h.OnSendPrompt(&runnerv1.SendPromptCommand{PodKey: pod.PodKey, Prompt: "hello"}); err != nil {
		t.Fatalf("OnSendPrompt error: %v", err)
	}

	base.mu.Lock()
	defer base.mu.Unlock()
	if len(base.inputs) != 1 {
		t.Fatalf("expected 1 SendInput for the body; got %d: %v", len(base.inputs), base.inputs)
	}
	if base.inputs[0].payload != "hello" {
		t.Errorf("body payload = %q, want %q", base.inputs[0].payload, "hello")
	}
	if len(base.keys) != 1 || base.keys[0].payload != "enter" {
		t.Fatalf("expected one SendKeys([\"enter\"]); got %v", base.keys)
	}
}

func TestOnSendPrompt_PTY_AdaptsVideoStudioPrompt(t *testing.T) {
	base := &sendPromptMockIO{mode: InteractionModePTY}
	io := &ptyTerminalMock{sendPromptMockIO: base}
	pod := &Pod{
		PodKey:          "video-pod",
		Agent:           "video-studio-codex",
		InteractionMode: InteractionModePTY,
		IO:              io,
	}

	store := NewInMemoryPodStore()
	store.Put(pod.PodKey, pod)
	h := &RunnerMessageHandler{podStore: store}

	prompt := "请生成视频\n输出 MP4"
	if err := h.OnSendPrompt(&runnerv1.SendPromptCommand{PodKey: pod.PodKey, Prompt: prompt}); err != nil {
		t.Fatalf("OnSendPrompt error: %v", err)
	}

	base.mu.Lock()
	defer base.mu.Unlock()
	if len(base.inputs) != 1 {
		t.Fatalf("expected 1 SendInput for the body; got %d: %v", len(base.inputs), base.inputs)
	}
	want := "\x1b[200~请生成视频\n输出 MP4\x1b[201~"
	if base.inputs[0].payload != want {
		t.Fatalf("body payload = %q, want %q", base.inputs[0].payload, want)
	}
	if len(base.keys) != 1 || base.keys[0].payload != "enter" {
		t.Fatalf("expected one SendKeys([\"enter\"]); got %v", base.keys)
	}
}

func TestOnSendPrompt_PTY_SubmitsVideoStudioWhenReadyDetectorStaysExecuting(t *testing.T) {
	oldTimeout := ptyPromptReadyTimeout
	ptyPromptReadyTimeout = 10 * time.Millisecond
	defer func() { ptyPromptReadyTimeout = oldTimeout }()

	base := &sendPromptMockIO{mode: InteractionModePTY, agentStatus: "executing"}
	io := &ptyTerminalMock{sendPromptMockIO: base}
	pod := &Pod{
		PodKey:          "video-pod",
		Agent:           "video-studio-codex",
		InteractionMode: InteractionModePTY,
		IO:              io,
	}

	store := NewInMemoryPodStore()
	store.Put(pod.PodKey, pod)
	h := &RunnerMessageHandler{podStore: store}

	if err := h.OnSendPrompt(&runnerv1.SendPromptCommand{PodKey: pod.PodKey, Prompt: "brief"}); err != nil {
		t.Fatalf("OnSendPrompt error: %v", err)
	}

	base.mu.Lock()
	defer base.mu.Unlock()
	if len(base.inputs) != 1 || base.inputs[0].payload != "\x1b[200~brief\x1b[201~" {
		t.Fatalf("prompt input = %v, want bracketed paste despite stale executing state", base.inputs)
	}
	if len(base.keys) != 1 || base.keys[0].payload != "enter" {
		t.Fatalf("submit keys = %v, want one enter key", base.keys)
	}
}

func TestOnSendPrompt_PTY_DoesNotSubmitGenericAgentWhenNotReady(t *testing.T) {
	oldTimeout := ptyPromptReadyTimeout
	ptyPromptReadyTimeout = 10 * time.Millisecond
	defer func() { ptyPromptReadyTimeout = oldTimeout }()

	base := &sendPromptMockIO{mode: InteractionModePTY, agentStatus: "executing"}
	io := &ptyTerminalMock{sendPromptMockIO: base}
	pod := &Pod{
		PodKey:          "generic-pod",
		Agent:           "generic-agent",
		InteractionMode: InteractionModePTY,
		IO:              io,
	}

	store := NewInMemoryPodStore()
	store.Put(pod.PodKey, pod)
	h := &RunnerMessageHandler{podStore: store}

	if err := h.OnSendPrompt(&runnerv1.SendPromptCommand{PodKey: pod.PodKey, Prompt: "brief"}); err == nil {
		t.Fatal("expected stale non-Codex PTY prompt to fail")
	}

	base.mu.Lock()
	defer base.mu.Unlock()
	if len(base.inputs) != 0 || len(base.keys) != 0 {
		t.Fatalf("generic agent must not receive input while not ready; inputs=%v keys=%v", base.inputs, base.keys)
	}
}

func TestOnSendPrompt_PTY_GapBetweenBodyAndEnter(t *testing.T) {
	base := &sendPromptMockIO{mode: InteractionModePTY}
	io := &ptyTerminalMock{sendPromptMockIO: base}
	pod := &Pod{PodKey: "pty-pod", InteractionMode: InteractionModePTY, IO: io}

	store := NewInMemoryPodStore()
	store.Put(pod.PodKey, pod)
	h := &RunnerMessageHandler{podStore: store}

	if err := h.OnSendPrompt(&runnerv1.SendPromptCommand{PodKey: pod.PodKey, Prompt: "hello"}); err != nil {
		t.Fatalf("OnSendPrompt error: %v", err)
	}

	base.mu.Lock()
	defer base.mu.Unlock()
	gap := base.keys[0].at.Sub(base.inputs[0].at)
	// Allow a small floor under ptySubmitGap to absorb scheduler jitter, but
	// require enough separation that the TUI's read loop can tick. Without
	// this gap the trailing Enter is folded into the body paste.
	const minGap = 50 * time.Millisecond
	if gap < minGap {
		t.Fatalf("gap between body and Enter = %v, want >= %v (TUI read-loop separation)", gap, minGap)
	}
}

func TestOnSendPrompt_PTY_WaitsUntilAgentAcceptsInput(t *testing.T) {
	base := &sendPromptMockIO{
		mode:        InteractionModePTY,
		agentStatus: "executing",
	}
	io := &ptyTerminalMock{sendPromptMockIO: base}
	pod := &Pod{PodKey: "pty-pod", InteractionMode: InteractionModePTY, IO: io}

	store := NewInMemoryPodStore()
	store.Put(pod.PodKey, pod)
	h := &RunnerMessageHandler{podStore: store}

	done := make(chan error, 1)
	go func() {
		done <- h.OnSendPrompt(&runnerv1.SendPromptCommand{
			PodKey: pod.PodKey,
			Prompt: "hello",
		})
	}()

	time.Sleep(20 * time.Millisecond)
	base.mu.Lock()
	inputCount := len(base.inputs)
	base.mu.Unlock()
	if inputCount != 0 {
		t.Fatalf("prompt was written before the PTY became ready")
	}

	base.setAgentStatus("waiting")
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("OnSendPrompt error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("OnSendPrompt did not resume after the PTY became ready")
	}

	base.mu.Lock()
	defer base.mu.Unlock()
	if len(base.inputs) != 1 || base.inputs[0].payload != "hello" {
		t.Fatalf("prompt input = %v, want one hello input", base.inputs)
	}
	if len(base.keys) != 1 || base.keys[0].payload != "enter" {
		t.Fatalf("submit keys = %v, want one enter key", base.keys)
	}
}

func TestOnSendPrompt_ACP_NoEnterKey(t *testing.T) {
	io := &sendPromptMockIO{mode: InteractionModeACP}
	pod := &Pod{PodKey: "acp-pod", InteractionMode: InteractionModeACP, IO: io}

	store := NewInMemoryPodStore()
	store.Put(pod.PodKey, pod)
	h := &RunnerMessageHandler{podStore: store}

	if err := h.OnSendPrompt(&runnerv1.SendPromptCommand{PodKey: pod.PodKey, Prompt: "hello"}); err != nil {
		t.Fatalf("OnSendPrompt error: %v", err)
	}

	io.mu.Lock()
	defer io.mu.Unlock()
	if len(io.inputs) != 1 {
		t.Fatalf("ACP must submit via the ACP RPC only (1 SendInput); got %d: %v", len(io.inputs), io.inputs)
	}
	if io.inputs[0].payload != "hello" {
		t.Errorf("body payload = %q, want %q", io.inputs[0].payload, "hello")
	}
	if len(io.keys) != 0 {
		t.Fatalf("ACP must not press Enter via SendKeys; got %v", io.keys)
	}
}
