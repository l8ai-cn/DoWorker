package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/agentsmesh/runner/internal/fsutil"
)

func writeKnowledgeMountKey(sandboxRoot, privateKey, knownHosts string) (string, error) {
	file, err := os.CreateTemp(sandboxRoot, ".agentsmesh-kb-key-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary SSH key: %w", err)
	}
	path := file.Name()
	_, err = file.WriteString(privateKey)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", joinKnowledgeMountCredentialCleanup(
			path,
			fmt.Errorf("failed to write temporary SSH key: %w", err),
		)
	}
	if err := os.WriteFile(path+temporaryKnownHostsFileSuffix, []byte(knownHosts+"\n"), 0600); err != nil {
		return "", joinKnowledgeMountCredentialCleanup(
			path,
			fmt.Errorf("failed to create temporary SSH known_hosts: %w", err),
		)
	}
	return path, nil
}

func persistKnowledgeMountKey(
	ctx context.Context,
	sandboxRoot, dest, temporaryKey, privateKey string,
) error {
	if err := validateKnowledgeMountRepoPath(sandboxRoot, dest); err != nil {
		return err
	}
	keyPath := filepath.Join(dest, ".git", knowledgeMountDeployKey)
	knownHostsPath := filepath.Join(dest, ".git", knowledgeMountKnownHosts)
	if err := os.Rename(temporaryKey, keyPath); err != nil {
		return fmt.Errorf("failed to persist knowledge mount SSH key: %w", err)
	}
	if err := os.Rename(temporaryKey+temporaryKnownHostsFileSuffix, knownHostsPath); err != nil {
		return fmt.Errorf("failed to persist knowledge mount SSH known_hosts: %w", err)
	}
	if err := os.Chmod(keyPath, 0600); err != nil {
		return fmt.Errorf("failed to secure knowledge mount SSH key: %w", err)
	}
	if err := os.Chmod(knownHostsPath, 0600); err != nil {
		return fmt.Errorf("failed to secure knowledge mount SSH known_hosts: %w", err)
	}
	configCmd := exec.CommandContext(ctx, "git", "config", "--local", "core.sshCommand",
		knowledgeMountSSHCommand(keyPath, knownHostsPath))
	configCmd.Dir = dest
	configCmd.Env = knowledgeMountSSHEnv(keyPath, knownHostsPath)
	if output, err := configCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s", knowledgeMountCommandError(
			"configure SSH push access for", filepath.Base(dest), err, output, privateKey))
	}
	return nil
}

func removeKnowledgeMountCredential(keyPath string) error {
	keyErr := fsutil.RemoveAll(keyPath)
	knownHostsErr := fsutil.RemoveAll(keyPath + temporaryKnownHostsFileSuffix)
	if keyErr != nil {
		return keyErr
	}
	return knownHostsErr
}

func joinKnowledgeMountCredentialCleanup(keyPath string, primary error) error {
	if keyPath == "" {
		return primary
	}
	cleanupErr := removeKnowledgeMountCredential(keyPath)
	if cleanupErr == nil {
		return primary
	}
	return errors.Join(
		primary,
		fmt.Errorf("remove temporary knowledge mount credential: %w", cleanupErr),
	)
}

func knowledgeMountSSHEnv(keyPath, knownHostsPath string) []string {
	return append(os.Environ(),
		"GIT_ASKPASS=", "SSH_ASKPASS=", "GIT_TERMINAL_PROMPT=0",
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_SSH_VARIANT=ssh", "GIT_SSH_COMMAND="+knowledgeMountSSHCommand(keyPath, knownHostsPath))
}

func knowledgeMountSSHCommand(keyPath, knownHostsPath string) string {
	quotedKey := shellQuoteKnowledgeMountPath(keyPath)
	quotedKnownHosts := shellQuoteKnowledgeMountPath(knownHostsPath)
	return "ssh -i " + quotedKey + " -o IdentitiesOnly=yes -o IdentityAgent=none" +
		" -o StrictHostKeyChecking=yes -o UserKnownHostsFile=" + quotedKnownHosts +
		" -o GlobalKnownHostsFile=/dev/null -o BatchMode=yes"
}

func shellQuoteKnowledgeMountPath(path string) string {
	return "'" + strings.ReplaceAll(path, "'", "'\"'\"'") + "'"
}

func trimKnowledgeMountOutput(value string) string {
	return strings.TrimSpace(value)
}

func redactKnowledgeMountPrivateKey(value, privateKey string) string {
	if privateKey == "" {
		return value
	}
	value = strings.ReplaceAll(value, privateKey, "[REDACTED]")
	for _, line := range strings.Split(strings.ReplaceAll(privateKey, "\r\n", "\n"), "\n") {
		if len(line) >= 4 {
			value = strings.ReplaceAll(value, line, "[REDACTED]")
		}
	}
	return value
}
