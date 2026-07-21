package acp

import (
	"fmt"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/processmgr"
)

func (c *ACPClient) Start() error {
	c.setState(StateInitializing)

	wrappedCallbacks := c.wrapCallbacks()
	transport, err := NewTransport(c.cfg.TransportType, wrappedCallbacks, c.logger)
	if err != nil {
		return err
	}
	c.transport = transport

	proc, err := processmgr.Global().Start(c.ctx, processmgr.Spec{
		Owner:       "acp:" + c.cfg.Command,
		Command:     c.cfg.Command,
		Args:        c.cfg.Args,
		Dir:         c.cfg.WorkDir,
		Env:         c.cfg.Env,
		Mode:        processmgr.ModeNormal,
		PipeStdin:   true,
		PipeStdout:  true,
		PipeStderr:  true,
		StopTimeout: time.Second,
	})
	if err != nil {
		return fmt.Errorf("start process: %w", err)
	}
	c.proc = proc
	stdin := proc.StdinWriter()
	stdout := proc.StdoutReader()
	stderr := proc.StderrReader()

	if err := c.transport.Initialize(c.ctx, stdin, stdout, stderr); err != nil {
		c.Stop()
		return fmt.Errorf("transport initialize: %w", err)
	}

	go c.readStderr(stderr)
	go c.transport.ReadLoop(c.ctx)
	go c.watchExit()

	sessionID, err := c.transport.Handshake(c.ctx)
	if err != nil {
		c.Stop()
		return fmt.Errorf("initialize: %w", err)
	}
	if sessionID != "" {
		c.sessionMu.Lock()
		c.sessionID = sessionID
		c.sessionMu.Unlock()
	}

	c.captureTransportCapabilities()
	c.setState(StateIdle)
	return nil
}
