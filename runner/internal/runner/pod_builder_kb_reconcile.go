package runner

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/fsutil"
)

func clearManagedKnowledgeCredentials(ctx context.Context, sandboxRoot string) error {
	temporary, err := filepath.Glob(filepath.Join(sandboxRoot, ".agentcloud-kb-key-*"))
	if err != nil {
		return fmt.Errorf("find temporary knowledge mount credentials: %w", err)
	}
	for _, path := range temporary {
		if err := fsutil.RemoveAll(path); err != nil {
			return fmt.Errorf("remove temporary knowledge mount credential: %w", err)
		}
	}
	return filepath.WalkDir(sandboxRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && entry.Name() == ".git" {
			if err := clearKnowledgeRepoCredential(ctx, filepath.Dir(path)); err != nil {
				return err
			}
			return filepath.SkipDir
		}
		return nil
	})
}

func clearKnowledgeRepoCredential(ctx context.Context, repoDir string) error {
	gitDir := filepath.Join(repoDir, ".git")
	for _, name := range []string{knowledgeMountDeployKey, knowledgeMountKnownHosts} {
		if err := fsutil.RemoveAll(filepath.Join(gitDir, name)); err != nil {
			return fmt.Errorf("remove managed knowledge mount credential: %w", err)
		}
	}
	get := exec.CommandContext(ctx, "git", "config", "--local", "--get", "core.sshCommand")
	get.Dir = repoDir
	get.Env = knowledgeMountGitConfigEnv()
	output, err := get.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil
		}
		return fmt.Errorf("inspect knowledge mount SSH config: %w", err)
	}
	if !strings.Contains(string(output), knowledgeMountDeployKey) {
		return nil
	}
	unset := exec.CommandContext(ctx, "git", "config", "--local", "--unset-all", "core.sshCommand")
	unset.Dir = repoDir
	unset.Env = knowledgeMountGitConfigEnv()
	if output, err := unset.CombinedOutput(); err != nil {
		return fmt.Errorf("remove knowledge mount SSH config: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func reconcileExistingKnowledgeMount(
	ctx context.Context,
	sandboxRoot, dest string,
	m *runnerv1.KnowledgeMount,
) error {
	if err := validateKnowledgeMountRepoPath(sandboxRoot, dest); err != nil {
		return err
	}
	remoteURL := knowledgeMountRemoteURL(m)
	if m.GetMode() != "rw" {
		return setKnowledgeMountRemote(ctx, dest, remoteURL, m.GetGitPrivateKey())
	}
	if err := setKnowledgeMountRemote(ctx, dest, remoteURL, m.GetGitPrivateKey()); err != nil {
		return err
	}
	temporaryKey, err := writeKnowledgeMountKey(
		sandboxRoot,
		m.GetGitPrivateKey(),
		m.GetGitKnownHosts(),
	)
	if err != nil {
		return err
	}
	if err := persistKnowledgeMountKey(
		ctx,
		sandboxRoot,
		dest,
		temporaryKey,
		m.GetGitPrivateKey(),
	); err != nil {
		clearErr := clearKnowledgeRepoCredential(ctx, dest)
		return errors.Join(
			err,
			clearErr,
			joinKnowledgeMountCredentialCleanup(temporaryKey, nil),
		)
	}
	return nil
}

func validateKnowledgeMountRepoPath(sandboxRoot, dest string) error {
	if err := validateKnowledgeMountDestinationPath(sandboxRoot, dest); err != nil {
		return err
	}
	info, err := os.Lstat(filepath.Join(dest, ".git"))
	if err != nil {
		return fmt.Errorf("inspect knowledge mount repository: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("knowledge mount repository path must contain only directories")
	}
	return nil
}

func validateKnowledgeMountDestinationPath(sandboxRoot, dest string) error {
	root, err := filepath.Abs(sandboxRoot)
	if err != nil {
		return fmt.Errorf("resolve sandbox root: %w", err)
	}
	target, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("resolve knowledge mount path: %w", err)
	}
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("knowledge mount repository escapes sandbox")
	}
	current := root
	for _, component := range strings.Split(relative, string(os.PathSeparator)) {
		current = filepath.Join(current, component)
		info, statErr := os.Lstat(current)
		if os.IsNotExist(statErr) {
			return nil
		}
		if statErr != nil {
			return fmt.Errorf("inspect knowledge mount repository: %w", statErr)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("knowledge mount repository path must contain only directories")
		}
	}
	return nil
}

func setKnowledgeMountRemote(ctx context.Context, dest, remoteURL, privateKey string) error {
	if remoteURL == "" {
		return fmt.Errorf("knowledge mount remote URL is required")
	}
	command := exec.CommandContext(ctx, "git", "remote", "set-url", "origin", remoteURL)
	command.Dir = dest
	command.Env = knowledgeMountGitConfigEnv()
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("%s", knowledgeMountCommandError(
			"configure remote for",
			filepath.Base(dest),
			err,
			output,
			privateKey,
		))
	}
	return nil
}

func knowledgeMountGitConfigEnv() []string {
	env := removeKnowledgeMountInheritedGitEnv(os.Environ())
	return append(env,
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_COUNT=2",
		"GIT_CONFIG_KEY_0=credential.helper", "GIT_CONFIG_VALUE_0=",
		"GIT_CONFIG_KEY_1=http.extraHeader", "GIT_CONFIG_VALUE_1=",
		"GIT_ASKPASS=", "SSH_ASKPASS=", "SSH_AUTH_SOCK=",
	)
}

func knowledgeMountRemoteURL(m *runnerv1.KnowledgeMount) string {
	if m.GetMode() == "rw" || (m.GetGitPrivateKey() != "" && m.GetHttpCloneUrl() == "") {
		return m.GetSshCloneUrl()
	}
	return m.GetHttpCloneUrl()
}
