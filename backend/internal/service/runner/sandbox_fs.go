package runner

import (
	"context"
	"sync"
	"time"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/google/uuid"
)

const (
	SandboxFsTimeout       = 30 * time.Second
	SandboxFsUploadTimeout = 150 * time.Second
)

type SandboxFsSender interface {
	SendSandboxFs(runnerID int64, cmd *runnerv1.SandboxFsCommand) error
	IsConnected(runnerID int64) bool
}

type pendingFsQuery struct {
	resultCh chan *runnerv1.SandboxFsResultEvent
	timeout  time.Time
}

type SandboxFsService struct {
	pending sync.Map
	done    chan struct{}
	sender  SandboxFsSender
}

func NewSandboxFsService(cm *RunnerConnectionManager) *SandboxFsService {
	s := &SandboxFsService{done: make(chan struct{})}
	if cm != nil {
		cm.SetSandboxFsResultCallback(func(_ int64, data *runnerv1.SandboxFsResultEvent) {
			s.complete(data)
		})
	}
	go s.cleanupLoop()
	return s
}

func (s *SandboxFsService) Stop() { close(s.done) }

func (s *SandboxFsService) SetSender(sender SandboxFsSender) { s.sender = sender }

func (s *SandboxFsService) IsConnected(runnerID int64) bool {
	return s.sender != nil && s.sender.IsConnected(runnerID)
}

func (s *SandboxFsService) Exec(ctx context.Context, runnerID int64, cmd *runnerv1.SandboxFsCommand) (*runnerv1.SandboxFsResultEvent, error) {
	if s.sender == nil {
		return nil, ErrCommandSenderNotSet
	}
	if cmd.RequestId == "" {
		cmd.RequestId = uuid.New().String()
	}
	ch := make(chan *runnerv1.SandboxFsResultEvent, 1)
	timeout := sandboxFsCommandTimeout(cmd.GetOp())
	s.pending.Store(cmd.RequestId, &pendingFsQuery{resultCh: ch, timeout: time.Now().Add(timeout)})
	if err := s.sender.SendSandboxFs(runnerID, cmd); err != nil {
		s.pending.Delete(cmd.RequestId)
		return nil, err
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case res := <-ch:
		return res, nil
	case <-ctx.Done():
		s.pending.Delete(cmd.RequestId)
		return nil, ctx.Err()
	case <-timer.C:
		s.pending.Delete(cmd.RequestId)
		return &runnerv1.SandboxFsResultEvent{RequestId: cmd.RequestId, Error: "query timeout"}, nil
	}
}

func sandboxFsCommandTimeout(op string) time.Duration {
	if op == "upload" || op == "read_verified_bytes" {
		return SandboxFsUploadTimeout
	}
	return SandboxFsTimeout
}

func (s *SandboxFsService) complete(event *runnerv1.SandboxFsResultEvent) {
	if v, ok := s.pending.LoadAndDelete(event.RequestId); ok {
		pq := v.(*pendingFsQuery)
		select {
		case pq.resultCh <- event:
		default:
		}
	}
}

func (s *SandboxFsService) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			now := time.Now()
			s.pending.Range(func(key, value any) bool {
				pq := value.(*pendingFsQuery)
				if now.After(pq.timeout) {
					if v, ok := s.pending.LoadAndDelete(key); ok {
						pending := v.(*pendingFsQuery)
						select {
						case pending.resultCh <- &runnerv1.SandboxFsResultEvent{
							RequestId: key.(string),
							Error:     "query timeout",
						}:
						default:
						}
					}
				}
				return true
			})
		}
	}
}
