package runner

import (
	"os"
	"path/filepath"
	"testing"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/config"
	"github.com/stretchr/testify/require"
)

func TestBuild_CodexACPWithCopiedHomeAndMCPFile(t *testing.T) {
	userHome := t.TempDir()
	userCodex := filepath.Join(userHome, ".codex")
	require.NoError(t, os.MkdirAll(userCodex, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(userCodex, "config.toml"), []byte(`approval_policy = "never"
sandbox_mode = "danger-full-access"
model = "gpt-5.5"
model_provider = "OpenAI"

[model_providers.OpenAI]
name = "OpenAI"
base_url = "https://token.aiedulab.cn"
wire_api = "responses"
requires_openai_auth = false
`), 0o644))
	t.Setenv("HOME", userHome)

	cmd := &runnerv1.CreatePodCommand{
		PodKey:          "codex-acp-build-test",
		LaunchCommand:   "codex",
		LaunchArgs:      []string{"app-server"},
		InteractionMode: "acp",
		EnvVars: map[string]string{
			"CODEX_HOME":     "{{sandbox_root}}/codex-home",
			"OPENAI_API_KEY": "test-key",
		},
		FilesToCreate: []*runnerv1.FileToCreate{
			{
				Path:    "{{sandbox_root}}/codex-home/config.toml",
				Content: "[mcp_servers.agentsmesh]\nurl = \"http://127.0.0.1:19000/mcp\"\n",
			},
			{Path: "{{work_dir}}/.codex", IsDirectory: true},
			{Path: "{{work_dir}}/.codex/mcp.json", Content: "{}"},
		},
	}

	tmpDir := t.TempDir()
	r := &Runner{cfg: &config.Config{WorkspaceRoot: tmpDir}}
	builder := NewPodBuilderFromRunner(r).WithCommand(cmd)
	pod, err := builder.Build(t.Context())
	require.NoError(t, err)
	require.NotNil(t, pod)
	require.Equal(t, InteractionModeACP, pod.InteractionMode)
}
