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
	workspace, err := openSandboxWorkspace(workspaceRoot)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	defer workspace.Close()
	return h.sandboxFsWorkerSkillDiscoverWorkspace(workspace, path)
}

func (h *RunnerMessageHandler) sandboxFsWorkerSkillDiscoverWorkspace(
	workspace *sandboxWorkspace,
	path string,
) (*runnerv1.SandboxFsResultEvent, error) {
	relative, display, err := resolveWorkerSkillRelativePath(path)
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	info, err := workspace.root.Stat(relative)
	if err != nil {
		if os.IsNotExist(err) {
			return fsErrResult("worker skill not found"), nil
		}
		return fsErrResult(err.Error()), nil
	}
	if !info.IsDir() {
		return fsErrResult("worker skill path is not a directory"), nil
	}
	content, err := workspace.root.ReadFile(filepath.Join(relative, "SKILL.md"))
	if os.IsNotExist(err) {
		return fsErrResult("worker skill SKILL.md not found"), nil
	}
	if err != nil {
		return fsErrResult(err.Error()), nil
	}
	if err := validateWorkerSkillFrontmatter(content); err != nil {
		return fsErrResult(err.Error()), nil
	}
	return &runnerv1.SandboxFsResultEvent{
		Entries: []*runnerv1.SandboxFsEntry{{
			Path: display,
			Name: filepath.Base(display),
			Type: "directory",
		}},
		WorkspaceRoot: workspace.displayPath(),
	}, nil
}

type workerSkillFrontmatter struct {
	Name string `yaml:"name"`
}

func validateWorkerSkillFrontmatter(content []byte) error {
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

func resolveWorkerSkillRelativePath(requested string) (string, string, error) {
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
	return filepath.FromSlash(display), display, nil
}

func isWorkerSkillPathAllowed(path string) bool {
	for _, root := range workerSkillRoots {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}
