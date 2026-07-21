package agent

import (
	"context"
	"testing"

	agentdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	envbundledomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	envbundleservice "github.com/l8ai-cn/agentcloud/backend/internal/service/envbundle"
	"github.com/stretchr/testify/assert"
)

type configBundleAgentProvider struct {
	agent *agentdomain.Agent
}

func (provider configBundleAgentProvider) GetAgent(
	context.Context,
	string,
) (*agentdomain.Agent, error) {
	return provider.agent, nil
}

type malformedConfigBundleLoader struct{}

func (malformedConfigBundleLoader) GetEffectiveForUser(
	context.Context,
	int64,
	int64,
	string,
) ([]*envbundleservice.EffectiveBundle, error) {
	return []*envbundleservice.EffectiveBundle{{
		Name: "settings",
		Kind: envbundledomain.KindConfig,
		Data: map[string]string{envbundledomain.ConfigJSONDataKey: "{"},
	}}, nil
}

func TestBuildPodCommandRejectsMalformedConfigBundle(t *testing.T) {
	source := "USE_CONFIG_BUNDLE \"settings\"\n"
	builder := NewConfigBuilder(
		configBundleAgentProvider{
			agent: &agentdomain.Agent{
				Slug:            "do-agent",
				AgentfileSource: &source,
			},
		},
		malformedConfigBundleLoader{},
	)

	_, err := builder.BuildPodCommand(context.Background(), &ConfigBuildRequest{
		AgentSlug:             "do-agent",
		MergedAgentfileSource: source,
	})

	assert.ErrorContains(t, err, "config bundle data")
}
