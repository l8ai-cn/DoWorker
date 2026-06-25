package coordinator

import (
	"context"
	"errors"
	"testing"

	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	runnersvc "github.com/anthropics/agentsmesh/backend/internal/service/runner"
)

type fakeRunnerSelector struct {
	calls int
	err   error
}

func (f *fakeRunnerSelector) SelectRunnerWithAffinity(
	context.Context, int64, int64, string, *runnerDomain.AffinityHints, map[int64]int,
) (*runnerDomain.Runner, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return &runnerDomain.Runner{ID: 1}, nil
}

type fakeRunnerLauncher struct {
	calls int
	err   error
}

func (f *fakeRunnerLauncher) Launch(context.Context, int64, string) error {
	f.calls++
	return f.err
}

func TestRunnerEnsurerSkipsLaunchWhenOnline(t *testing.T) {
	selector := &fakeRunnerSelector{}
	launcher := &fakeRunnerLauncher{}
	e := NewRunnerEnsurer(selector, launcher, nil)
	if err := e.Ensure(context.Background(), 1, 2, "do-agent"); err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if selector.calls != 1 {
		t.Fatalf("selector calls = %d, want 1", selector.calls)
	}
	if launcher.calls != 0 {
		t.Fatalf("launcher calls = %d, want 0", launcher.calls)
	}
}

func TestRunnerEnsurerLaunchesWhenMissing(t *testing.T) {
	selector := &fakeRunnerSelector{err: runnersvc.ErrNoRunnerForAgent}
	launcher := &fakeRunnerLauncher{}
	e := NewRunnerEnsurer(selector, launcher, nil)
	e.wait = 0
	e.poll = 0

	selector.err = runnersvc.ErrNoRunnerForAgent
	if err := e.Ensure(context.Background(), 1, 2, "do-agent"); err == nil {
		t.Fatal("expected error when runner never appears")
	}
	if launcher.calls != 1 {
		t.Fatalf("launcher calls = %d, want 1", launcher.calls)
	}

	selector.err = nil
	if err := e.Ensure(context.Background(), 1, 2, "do-agent"); err != nil {
		t.Fatalf("Ensure after provision: %v", err)
	}
	if launcher.calls != 1 {
		t.Fatalf("launcher calls = %d, want 1 (no relaunch)", launcher.calls)
	}
}

func TestRunnerEnsurerNoLauncherReturnsErr(t *testing.T) {
	e := NewRunnerEnsurer(&fakeRunnerSelector{err: runnersvc.ErrNoRunnerForAgent}, nil, nil)
	err := e.Ensure(context.Background(), 1, 2, "do-agent")
	if !errors.Is(err, runnersvc.ErrNoRunnerForAgent) {
		t.Fatalf("err = %v, want ErrNoRunnerForAgent", err)
	}
}
