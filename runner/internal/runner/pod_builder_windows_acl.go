package runner

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
)

var (
	runnerGOOS = runtime.GOOS
	runICACLS  = func(path, username string) error {
		return exec.Command("icacls", path, "/inheritance:r", "/grant:r", username+":R").Run()
	}
)

func secureWindowsPrivateKey(path string) error {
	if runnerGOOS != "windows" {
		return nil
	}
	username := os.Getenv("USERNAME")
	if username == "" {
		if u, err := user.Current(); err == nil {
			username = u.Username
		}
	}
	if username == "" {
		return fmt.Errorf("windows username is required to secure SSH key ACL")
	}
	if err := runICACLS(path, username); err != nil {
		return fmt.Errorf("failed to set SSH key ACL: %w", err)
	}
	return nil
}
