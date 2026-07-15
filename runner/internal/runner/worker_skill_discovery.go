package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"gopkg.in/yaml.v3"
)

var workerSkillRoots = []string{"skills", ".agents/skills", ".codex/skills"}

func (h *RunnerMessageHandler) sandboxFsWorkerSkillDiscover(workspaceRoot, path string) (*runnerv1.SandboxFsResultEvent, error) {
	skillDir, display, err := resolveWorkerSkillDirectory(workspaceRoot, path)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	info, err := os.Stat(skillDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fsErrResult("worker skill not found"), nil
		}
		return fsErrResult(err.Error()), nil
	}
	if !info.IsDir() {
		return fsErrResult("worker skill path is not a directory"), nil
	}
	if err := validateWorkerSkillFrontmatter(filepath.Join(skillDir, "SKILL.md")); err != nil {
		return fsErrResult(err.Error()), nil
	}
	return &runnerv1.SandboxFsResultEvent{
		Entries: []*runnerv1.SandboxFsEntry{{
			Path: display,
			Name: filepath.Base(display),
			Type: "directory",
		}},
		WorkspaceRoot: workspaceRoot,
	}, nil
}

type workerSkillFrontmatter struct {
	Name string `yaml:"name"`
}

func validateWorkerSkillFrontmatter(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("worker skill SKILL.md not found")
		}
		return err
	}
	sections := strings.SplitN(string(content), "\n---\n", 2)
	if len(sections) != 2 || strings.TrimSpace(sections[0]) == "" || !strings.HasPrefix(sections[0], "---\n") {
		return fmt.Errorf("worker skill frontmatter is invalid")
	}
	var frontmatter workerSkillFrontmatter
	if err := yaml.Unmarshal([]byte(strings.TrimPrefix(sections[0], "---\n")), &frontmatter); err != nil {
		return fmt.Errorf("worker skill frontmatter is invalid")
	}
	if err := slugkit.Validate(strings.TrimSpace(frontmatter.Name)); err != nil {
		return fmt.Errorf("worker skill name is invalid")
	}
	return nil
}

func resolveWorkerSkillDirectory(workspaceRoot, requested string) (string, string, error) {
	raw := strings.TrimSpace(requested)
	if filepath.IsAbs(raw) {
		return "", "", fmt.Errorf("worker skill path must be relative")
	}
	clean := filepath.Clean(raw)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("worker skill path escapes workspace")
	}
	display := filepath.ToSlash(clean)
	if !isWorkerSkillPathAllowed(display) {
		return "", "", fmt.Errorf("worker skill path is outside allowed roots")
	}
	workspace, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return "", "", err
	}
	candidate := filepath.Join(workspace, filepath.FromSlash(display))
	resolvedWorkspace, err := filepath.EvalSymlinks(workspace)
	if err != nil {
		return "", "", err
	}
	resolvedCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		if os.IsNotExist(err) {
			return candidate, display, nil
		}
		return "", "", err
	}
	if !pathIsWithin(resolvedWorkspace, resolvedCandidate) {
		return "", "", fmt.Errorf("worker skill path escapes workspace")
	}
	return resolvedCandidate, display, nil
}

func isWorkerSkillPathAllowed(path string) bool {
	for _, root := range workerSkillRoots {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}

func pathIsWithin(root, path string) bool {
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}
