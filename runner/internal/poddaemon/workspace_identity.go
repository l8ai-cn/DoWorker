package poddaemon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var errWorkspaceIdentityMissing = errors.New("workspace identity is missing")

type WorkspaceIdentity struct {
	Kind          string `json:"kind"`
	Device        uint64 `json:"device"`
	File          uint64 `json:"file"`
	CanonicalPath string `json:"canonical_path"`
}

func CaptureWorkspaceIdentity(path string) (*WorkspaceIdentity, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("workspace path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace path: %w", err)
	}
	file, err := openWorkspaceIdentityFile(absolute)
	if err != nil {
		return nil, fmt.Errorf("open workspace identity: %w", err)
	}
	defer file.Close()
	kind, device, fileID, err := workspaceFileIdentity(file)
	if err != nil {
		return nil, err
	}
	canonical, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace symlinks: %w", err)
	}
	return &WorkspaceIdentity{
		Kind: kind, Device: device, File: fileID,
		CanonicalPath: filepath.Clean(canonical),
	}, nil
}

func ValidateWorkspaceIdentity(
	path string,
	expected *WorkspaceIdentity,
) error {
	if err := validateExpectedWorkspaceIdentity(expected); err != nil {
		return err
	}
	actual, err := CaptureWorkspaceIdentity(path)
	if err != nil {
		return err
	}
	if !sameWorkspaceIdentity(actual, expected, true) {
		return fmt.Errorf("workspace identity changed")
	}
	return nil
}

func ValidateWorkspaceFile(
	file *os.File,
	expected *WorkspaceIdentity,
) error {
	if err := validateExpectedWorkspaceIdentity(expected); err != nil {
		return err
	}
	kind, device, fileID, err := workspaceFileIdentity(file)
	if err != nil {
		return err
	}
	actual := &WorkspaceIdentity{Kind: kind, Device: device, File: fileID}
	if !sameWorkspaceIdentity(actual, expected, false) {
		return fmt.Errorf("workspace identity changed")
	}
	return nil
}

func OpenWorkspaceLaunchGuard(
	path string,
	expected *WorkspaceIdentity,
) (*os.File, error) {
	if err := validateExpectedWorkspaceIdentity(expected); err != nil {
		return nil, err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace path: %w", err)
	}
	file, err := openWorkspaceLaunchFile(absolute)
	if err != nil {
		return nil, fmt.Errorf("lock workspace path: %w", err)
	}
	if err := ValidateWorkspaceFile(file, expected); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

func workspaceIdentityForSession(
	path string,
	expected *WorkspaceIdentity,
) (*WorkspaceIdentity, error) {
	if expected == nil {
		return CaptureWorkspaceIdentity(path)
	}
	if err := ValidateWorkspaceIdentity(path, expected); err != nil {
		return nil, err
	}
	identity := *expected
	return &identity, nil
}

func validateExpectedWorkspaceIdentity(identity *WorkspaceIdentity) error {
	if identity == nil ||
		identity.Kind == "" ||
		identity.CanonicalPath == "" {
		return errWorkspaceIdentityMissing
	}
	return nil
}

func sameWorkspaceIdentity(
	actual, expected *WorkspaceIdentity,
	comparePath bool,
) bool {
	if actual.Kind != expected.Kind ||
		actual.Device != expected.Device ||
		actual.File != expected.File {
		return false
	}
	return !comparePath ||
		filepath.Clean(actual.CanonicalPath) ==
			filepath.Clean(expected.CanonicalPath)
}
