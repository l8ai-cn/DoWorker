package workspace

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

type repositoryProbeCandidate struct {
	url  string
	desc string
}

func (m *Manager) probeRepositoryAccess(
	ctx context.Context,
	httpURL, sshURL string,
	opts *WorktreeOptions,
) (string, error) {
	for _, candidateURL := range []string{httpURL, sshURL} {
		if err := validateRepositoryURL(candidateURL); err != nil {
			return "", err
		}
	}
	if opts == nil {
		opts = &WorktreeOptions{}
	}
	candidates, err := repositoryProbeCandidates(httpURL, sshURL, opts)
	if err != nil {
		return "", err
	}
	var errors []string
	for _, candidate := range candidates {
		if selected, probeErr := m.probeRepositoryCandidate(ctx, candidate, opts); probeErr == nil {
			return selected, nil
		} else {
			errors = append(errors, probeErr.Error())
		}
	}
	return "", fmt.Errorf("all repository access methods failed:\n  %s", strings.Join(errors, "\n  "))
}

func repositoryProbeCandidates(
	httpURL, sshURL string,
	opts *WorktreeOptions,
) ([]repositoryProbeCandidate, error) {
	var candidates []repositoryProbeCandidate
	if opts.GitToken != "" && httpURL != "" {
		if err := validateTokenRepositoryURL(httpURL); err != nil {
			return nil, err
		}
		candidates = append(candidates, repositoryProbeCandidate{url: httpURL, desc: "HTTP+token"})
	}
	if opts.SSHKeyPath != "" && sshURL != "" {
		candidates = append(candidates, repositoryProbeCandidate{url: sshURL, desc: "SSH+key"})
	}
	if opts.AnonymousAuth {
		if httpURL != "" {
			candidates = append(candidates, repositoryProbeCandidate{url: httpURL, desc: "HTTP(anonymous)"})
		}
		if sshURL != "" {
			candidates = append(candidates, repositoryProbeCandidate{url: sshURL, desc: "SSH(anonymous)"})
		}
	}
	if len(candidates) == 0 {
		if opts.GitToken != "" || opts.SSHKeyPath != "" || opts.AnonymousAuth {
			return nil, fmt.Errorf("no clone URL is compatible with explicit git credential type")
		}
		if sshURL != "" {
			candidates = append(candidates, repositoryProbeCandidate{url: sshURL, desc: "SSH(local)"})
		}
		if httpURL != "" {
			candidates = append(candidates, repositoryProbeCandidate{url: httpURL, desc: "HTTP(local)"})
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no clone URLs available to probe")
	}
	return candidates, nil
}

func (m *Manager) probeRepositoryCandidate(
	ctx context.Context,
	candidate repositoryProbeCandidate,
	opts *WorktreeOptions,
) (string, error) {
	displayURL := RepositoryURLForDisplay(candidate.url)
	log := logger.Workspace()
	log.Debug("Probing repository access", "url", displayURL, "method", candidate.desc)
	probeCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, "git", "ls-remote", "--exit-code", candidate.url, "HEAD")
	m.setProbeEnv(cmd, opts)
	output, err := cmd.CombinedOutput()
	if err == nil {
		log.Info("Repository access probe succeeded", "url", displayURL, "method", candidate.desc)
		return candidate.url, nil
	}
	errDetail := err.Error()
	if probeCtx.Err() == context.DeadlineExceeded {
		errDetail = "timeout (connection took too long)"
	}
	errMessage := fmt.Sprintf("%s (%s): %s", candidate.desc, displayURL, errDetail)
	if strings.TrimSpace(string(output)) != "" {
		errMessage += " — " + m.redactGitOutput(opts, output)
	}
	log.Warn("Repository access probe failed",
		"url", displayURL, "method", candidate.desc, "error", errDetail,
		"output", m.redactGitOutput(opts, output))
	return "", fmt.Errorf("%s", errMessage)
}
