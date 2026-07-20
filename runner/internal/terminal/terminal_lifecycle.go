package terminal

import (
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/safego"
)

func (t *Terminal) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("terminal is closed")
	}

	log := logger.Terminal()
	log.Debug("Starting command", "command", t.command, "args", t.args, "dir", t.workDir, "cols", t.cols, "rows", t.rows)

	var proc ptyProcess
	var err error
	if t.ptyFactory != nil {
		proc, err = t.ptyFactory(t.command, t.args, t.workDir, t.env, t.cols, t.rows)
	} else {
		proc, err = startPTY(t.command, t.args, t.workDir, t.env, t.cols, t.rows)
	}
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}
	t.proc = proc

	log.Debug("PTY started", "pid", t.proc.Pid(), "cols", t.cols, "rows", t.rows)
	safego.Go("pty-read", t.readOutput)
	safego.Go("pty-wait", t.waitExit)
	log.Info("Terminal started", "pid", t.proc.Pid(), "cols", t.cols, "rows", t.rows)

	return nil
}

func (t *Terminal) Stop() {
	log := logger.Terminal()
	log.Info("Terminal stopping")

	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	proc := t.proc
	t.mu.Unlock()

	if proc != nil {
		pid := proc.Pid()
		log.Debug("Sending graceful stop signal", "pid", pid)
		if err := proc.GracefulStop(); err != nil {
			log.Debug("Graceful stop failed (process may have already exited)", "error", err)
		}
		select {
		case <-t.doneCh:
			log.Debug("Process exited gracefully")
		case <-time.After(gracefulStopTimeout):
			log.Warn("Process did not exit after graceful stop, killing",
				"pid", pid, "timeout", gracefulStopTimeout)
			if err := proc.Kill(); err != nil {
				log.Debug("Kill failed (process may have already exited)", "error", err)
			}
			select {
			case <-t.doneCh:
			case <-time.After(time.Second):
				log.Warn("Process did not exit after kill", "pid", pid)
			}
		}
	}

	t.closePTY()
	log.Info("Terminal stopped")
}

func (t *Terminal) closePTY() {
	t.ptyCloseOnce.Do(func() {
		if t.proc != nil {
			t.proc.Close()
		}
	})
}

func (t *Terminal) Detach() {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	t.mu.Unlock()
	logger.Terminal().Info("Terminal detaching (daemon stays alive)")
	t.closePTY()
}
