package runner

import (
	"fmt"
	"os"
	"path/filepath"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

// createFiles creates files from the FilesToCreate list.
func (b *PodBuilder) createFiles(sandboxRoot, workDir string) error {
	allowedRoots, err := resolvedFileCreationRoots(sandboxRoot, workDir)
	if err != nil {
		return &client.PodError{
			Code:    client.ErrCodeFileCreate,
			Message: fmt.Sprintf("failed to resolve file creation roots: %v", err),
		}
	}

	for _, f := range b.cmd.FilesToCreate {
		path := b.resolvePath(f.Path, sandboxRoot, workDir)

		absPath, err := filepath.Abs(path)
		if err != nil {
			return &client.PodError{
				Code:    client.ErrCodeFileCreate,
				Message: fmt.Sprintf("failed to resolve file path: %v", err),
				Details: map[string]string{"path": f.Path},
			}
		}
		resolvedPath, err := resolvePathThroughExistingSymlinks(absPath)
		if err != nil {
			return &client.PodError{
				Code:    client.ErrCodeFileCreate,
				Message: fmt.Sprintf("failed to resolve file path symlinks: %v", err),
				Details: map[string]string{"path": f.Path},
			}
		}
		if !pathWithinFileCreationRoots(resolvedPath, allowedRoots) {
			return &client.PodError{
				Code:    client.ErrCodeFileCreate,
				Message: fmt.Sprintf("path %q escapes allowed file roots (resolved: %q)", f.Path, resolvedPath),
				Details: map[string]string{"path": f.Path, "resolved_path": resolvedPath},
			}
		}

		if f.IsDirectory {
			if err := os.MkdirAll(path, 0755); err != nil {
				return &client.PodError{
					Code:    client.ErrCodeFileCreate,
					Message: fmt.Sprintf("failed to create directory: %v", err),
					Details: map[string]string{"path": path},
				}
			}
			continue
		}

		parentDir := filepath.Dir(path)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return &client.PodError{
				Code:    client.ErrCodeFileCreate,
				Message: fmt.Sprintf("failed to create parent directory: %v", err),
				Details: map[string]string{"path": parentDir},
			}
		}

		mode := os.FileMode(0644)
		if f.Mode != 0 {
			mode = os.FileMode(f.Mode)
		}

		if err := os.WriteFile(path, []byte(f.Content), mode); err != nil {
			return &client.PodError{
				Code:    client.ErrCodeFileCreate,
				Message: fmt.Sprintf("failed to write file: %v", err),
				Details: map[string]string{"path": path},
			}
		}

		logger.Pod().Debug("Created file", "path", path, "mode", fmt.Sprintf("%o", mode))
	}

	return nil
}

// createFilesFromProto creates files from a proto FileToCreate list.
// Paths may contain placeholders ({{sandbox_root}}, {{work_dir}}) which are resolved before use.
func (b *PodBuilder) createFilesFromProto(files []*runnerv1.FileToCreate, sandboxRoot, workDir string) error {
	if len(files) == 0 {
		return nil
	}

	allowedRoots, err := resolvedFileCreationRoots(sandboxRoot, workDir)
	if err != nil {
		return &client.PodError{
			Code:    client.ErrCodeFileCreate,
			Message: fmt.Sprintf("failed to resolve file creation roots: %v", err),
		}
	}

	for _, f := range files {
		// Resolve path placeholders before validation
		path := resolvePathPlaceholders(f.Path, sandboxRoot, workDir)
		content := resolvePathPlaceholders(f.Content, sandboxRoot, workDir)

		absPath, err := filepath.Abs(path)
		if err != nil {
			return &client.PodError{
				Code:    client.ErrCodeFileCreate,
				Message: fmt.Sprintf("failed to resolve file path: %v", err),
				Details: map[string]string{"path": path},
			}
		}
		resolvedPath, err := resolvePathThroughExistingSymlinks(absPath)
		if err != nil {
			return &client.PodError{
				Code:    client.ErrCodeFileCreate,
				Message: fmt.Sprintf("failed to resolve agentfile path symlinks: %v", err),
				Details: map[string]string{"path": path},
			}
		}
		if !pathWithinFileCreationRoots(resolvedPath, allowedRoots) {
			return &client.PodError{
				Code:    client.ErrCodeFileCreate,
				Message: fmt.Sprintf("agentfile path %q escapes allowed file roots", path),
				Details: map[string]string{"path": path},
			}
		}

		if f.IsDirectory {
			if err := os.MkdirAll(path, 0755); err != nil {
				return &client.PodError{
					Code:    client.ErrCodeFileCreate,
					Message: fmt.Sprintf("failed to create directory: %v", err),
					Details: map[string]string{"path": path},
				}
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return &client.PodError{
				Code:    client.ErrCodeFileCreate,
				Message: fmt.Sprintf("failed to create parent directory: %v", err),
				Details: map[string]string{"path": filepath.Dir(path)},
			}
		}
		mode := os.FileMode(0644)
		if f.Mode != 0 {
			mode = os.FileMode(f.Mode)
		}
		if err := os.WriteFile(path, []byte(content), mode); err != nil {
			return &client.PodError{
				Code:    client.ErrCodeFileCreate,
				Message: fmt.Sprintf("failed to write file: %v", err),
				Details: map[string]string{"path": path},
			}
		}
		logger.Pod().Debug("Created file (agentfile)", "path", path)
	}

	return nil
}

func resolvedFileCreationRoots(sandboxRoot, workDir string) ([]string, error) {
	roots := make([]string, 0, 2)
	for _, root := range []string{sandboxRoot, workDir} {
		absRoot, err := resolvePathThroughExistingSymlinks(root)
		if err != nil {
			return nil, err
		}
		duplicate := false
		for _, existing := range roots {
			if existing == absRoot {
				duplicate = true
				break
			}
		}
		if !duplicate {
			roots = append(roots, absRoot)
		}
	}
	return roots, nil
}

func pathWithinFileCreationRoots(path string, roots []string) bool {
	for _, root := range roots {
		if pathWithinRoot(path, root) {
			return true
		}
	}
	return false
}
