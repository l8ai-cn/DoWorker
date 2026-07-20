package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
)

func secureGitCredentialFile(path string) error {
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("failed to secure Git credential file: %w", err)
	}
	if runtime.GOOS != "windows" {
		return nil
	}
	username := os.Getenv("USERNAME")
	if username == "" {
		if current, err := user.Current(); err == nil {
			username = current.Username
		}
	}
	if username == "" {
		return fmt.Errorf("windows username is required to secure Git credential file")
	}
	if err := exec.Command("icacls", path, "/inheritance:r", "/grant:r", username+":R").Run(); err != nil {
		return fmt.Errorf("failed to set Git credential file ACL: %w", err)
	}
	return nil
}
