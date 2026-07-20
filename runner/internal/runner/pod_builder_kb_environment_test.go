package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveKnowledgeMountInheritedGitEnvIgnoresKeyCase(t *testing.T) {
	env := []string{
		"PATH=/usr/bin",
		"git_config_global=/tmp/global",
		"Git_Config_NoSystem=0",
		"git_config_count=1",
		"Git_Config_Key_0=http.extraHeader",
		"git_config_value_0=Authorization: leaked",
		"git_ssh_command=ssh -i leaked",
		"ssh_auth_sock=/tmp/agent",
		"git_askpass=/tmp/askpass",
		"ssh_askpass=/tmp/ssh-askpass",
	}

	assert.Equal(t, []string{"PATH=/usr/bin"}, removeKnowledgeMountInheritedGitEnv(env))
}
