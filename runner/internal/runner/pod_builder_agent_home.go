package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/agentsmesh/runner/internal/agentkit"
	"github.com/anthropics/agentsmesh/runner/internal/agents/codex"
	"github.com/anthropics/agentsmesh/runner/internal/agents/openclaw"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

// prepareAgentHome copies the user's agent config directory to a per-pod
// isolated directory when a registered AgentHomeSpec matches an env var.
// After copying, it merges platform config using the spec's MergeConfig.
func (b *PodBuilder) prepareAgentHome(sandboxRoot, workDir string) error {
	if b.cmd == nil || b.cmd.EnvVars == nil {
		return nil
	}

	spec, agentHome := agentkit.MatchAgentHome(b.cmd.EnvVars)
	if spec == nil {
		return nil
	}

	agentHome = b.resolvePath(agentHome, sandboxRoot, workDir)

	log := logger.Pod()
	log.Info("Preparing agent home", "pod_key", b.cmd.PodKey, "env_var", spec.EnvVar, "path", agentHome)

	home := userHomeDir()
	targetDir := agentHome
	if spec.EnvVar == "OPENCLAW_HOME" {
		targetDir = filepath.Join(agentHome, ".openclaw")
	}
	if home != "" {
		userDir := filepath.Join(home, spec.UserDirName)
		if dirExists(userDir) {
			if err := copyDirSelective(userDir, targetDir); err != nil {
				log.Warn("Failed to copy user agent dir, creating empty",
					"source", userDir, "dest", targetDir, "error", err)
				_ = os.RemoveAll(targetDir)
				if mkErr := os.MkdirAll(targetDir, 0755); mkErr != nil {
					return fmt.Errorf("failed to create agent home: %w", mkErr)
				}
			}
		} else {
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create agent home: %w", err)
			}
		}
	} else {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create agent home: %w", err)
		}
	}

	// Find matching config file in FilesToCreate and merge
	mergeIdx := -1
	for i, f := range b.cmd.FilesToCreate {
		resolvedPath := b.resolvePath(f.Path, sandboxRoot, workDir)
		parentDir := filepath.Dir(resolvedPath)
		if parentDir == targetDir && !f.IsDirectory && spec.MergeConfig != nil {
			mergeIdx = i
			break
		}
	}
	if mergeIdx >= 0 {
		f := b.cmd.FilesToCreate[mergeIdx]
		configPath := b.resolvePath(f.Path, sandboxRoot, workDir)
		if err := spec.MergeConfig(configPath, f.Content); err != nil {
			log.Warn("Failed to merge agent config, writing fresh",
				"path", configPath, "error", err)
		} else {
			b.cmd.FilesToCreate = append(b.cmd.FilesToCreate[:mergeIdx], b.cmd.FilesToCreate[mergeIdx+1:]...)
			log.Info("Merged platform config into existing agent config", "path", configPath)
		}
	}

	if spec.EnvVar == "CODEX_HOME" {
		configPath := filepath.Join(targetDir, "config.toml")
		baseURL := ""
		model := ""
		if b.cmd.EnvVars != nil {
			baseURL = b.cmd.EnvVars["OPENAI_BASE_URL"]
			model = b.cmd.EnvVars["OPENAI_MODEL"]
		}
		if err := codex.ApplyOpenAIProviderFromEnv(configPath, baseURL, model); err != nil {
			log.Warn("Failed to apply codex provider env", "path", configPath, "error", err)
		}
		if err := codex.WriteAuthJSONFromEnv(agentHome, b.cmd.EnvVars["OPENAI_API_KEY"]); err != nil {
			log.Warn("Failed to write codex auth.json", "path", agentHome, "error", err)
		}
		if err := codex.AppendCodexProjectTrust(configPath, workDir); err != nil {
			log.Warn("Failed to trust codex workspace paths", "path", configPath, "error", err)
		}
	}

	if spec.EnvVar == "OPENCLAW_HOME" {
		configPath := filepath.Join(targetDir, "openclaw.json")
		if err := openclaw.ApplyOpenAIProviderFromEnv(
			configPath,
			workDir,
			b.cmd.EnvVars["OPENAI_BASE_URL"],
			b.cmd.EnvVars["OPENAI_MODEL"],
		); err != nil {
			return fmt.Errorf("prepare openclaw provider config: %w", err)
		}
	}

	return nil
}
