package coordinator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
)

// ScriptLauncher runs an external script to provision a runner. The script receives
// org_id and agent_slug as argv[1] and argv[2]. Set via COORDINATOR_RUNNER_LAUNCHER.
type ScriptLauncher struct {
	Script string
	Logger *slog.Logger
}

func NewScriptLauncher(script string, logger *slog.Logger) *ScriptLauncher {
	if logger == nil {
		logger = slog.Default()
	}
	return &ScriptLauncher{Script: strings.TrimSpace(script), Logger: logger}
}

func (l *ScriptLauncher) Launch(ctx context.Context, orgID int64, agentSlug string) error {
	if l.Script == "" {
		return errors.New("coordinator: runner launch script path is empty")
	}
	cmd := exec.CommandContext(ctx, l.Script, strconv.FormatInt(orgID, 10), agentSlug)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script %q failed: %w (stderr: %s)", l.Script, err, strings.TrimSpace(stderr.String()))
	}
	if msg := strings.TrimSpace(stdout.String()); msg != "" {
		l.Logger.Debug("runner launch script output", "stdout", msg)
	}
	return nil
}
