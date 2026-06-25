package coordinator

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type commandRunner interface {
	Run(ctx context.Context, name string, args ...string) (stdout, stderr string, err error)
}

type execCommandRunner struct{}

func (execCommandRunner) Run(ctx context.Context, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func runnerInstanceID(orgID int64, agentSlug string) string {
	return sanitizeRunnerResourceName("amesh-runner-", fmt.Sprintf("%d-%s", orgID, agentSlug))
}

func sanitizeRunnerResourceName(prefix, seed string) string {
	seed = strings.TrimSpace(seed)
	if seed == "" {
		seed = "default"
	}
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		default:
			return '-'
		}
	}, seed)
	clean = strings.Trim(clean, "-")
	if len(clean) > 30 {
		clean = clean[:30]
	}
	if clean == "" {
		clean = "agent"
	}
	return prefix + clean
}
