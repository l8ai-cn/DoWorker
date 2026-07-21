package agent

import (
	"context"
	"testing"

	agentdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/extension"
	envbundleservice "github.com/l8ai-cn/agentcloud/backend/internal/service/envbundle"
	extensionservice "github.com/l8ai-cn/agentcloud/backend/internal/service/extension"
	"github.com/stretchr/testify/assert"
)

type mockSkillExtensionProvider struct {
	skills []*extensionservice.ResolvedSkill
}

func (m *mockSkillExtensionProvider) GetEffectiveMcpServers(
	context.Context, int64, int64, int64, string,
) ([]*extension.InstalledMcpServer, error) {
	return nil, nil
}

func (m *mockSkillExtensionProvider) GetEffectiveSkills(
	_ context.Context, _, _, _ int64, _ string,
) ([]*extensionservice.ResolvedSkill, error) {
	return m.skills, nil
}

type stubEnvBundleLoader struct{}

func (stubEnvBundleLoader) GetEffectiveForUser(
	context.Context, int64, int64, string,
) ([]*envbundleservice.EffectiveBundle, error) {
	return nil, nil
}

func TestFilterResolvedSkillsBySlugs(t *testing.T) {
	all := []*extensionservice.ResolvedSkill{
		{Slug: "alpha", ContentSha: "sha-a", DownloadURL: "https://example/a"},
		{Slug: "beta", ContentSha: "sha-b", DownloadURL: "https://example/b"},
		{Slug: "gamma", ContentSha: "sha-c", DownloadURL: "https://example/c"},
	}

	t.Run("empty request keeps all", func(t *testing.T) {
		got := filterResolvedSkillsBySlugs(all, nil)
		if len(got) != 3 {
			t.Fatalf("len = %d, want 3", len(got))
		}
	})

	t.Run("filters to requested slugs", func(t *testing.T) {
		got := filterResolvedSkillsBySlugs(all, []string{"beta", "missing"})
		if len(got) != 1 || got[0].Slug != "beta" {
			t.Fatalf("got %#v, want only beta", got)
		}
	})

	t.Run("builtin-only request keeps all installed skills", func(t *testing.T) {
		// Base AgentFile always declares SKILLS am-delegate, am-channel — none of
		// which are installed marketplace skills. With no user selection the
		// directive must NOT drop the repo's installed skills.
		got := filterResolvedSkillsBySlugs(all, []string{"am-delegate", "am-channel"})
		if len(got) != 3 {
			t.Fatalf("len = %d, want 3 (builtin slugs must not filter installed skills)", len(got))
		}
	})

	t.Run("mixed builtin + selected slug keeps only the installed match", func(t *testing.T) {
		got := filterResolvedSkillsBySlugs(all, []string{"am-delegate", "am-channel", "gamma"})
		if len(got) != 1 || got[0].Slug != "gamma" {
			t.Fatalf("got %#v, want only gamma", got)
		}
	})
}

func TestBuildSkillResources_FiltersByAgentfileSlugs(t *testing.T) {
	repoID := int64(99)
	builder := NewConfigBuilder(nilAgentConfigProvider{}, stubEnvBundleLoader{})
	builder.SetExtensionProvider(&mockSkillExtensionProvider{
		skills: []*extensionservice.ResolvedSkill{
			{Slug: "keep-me", ContentSha: "sha1", DownloadURL: "https://example/1", PackageSize: 10},
			{Slug: "drop-me", ContentSha: "sha2", DownloadURL: "https://example/2", PackageSize: 20},
		},
	})

	resources, err := builder.buildSkillResources(
		context.Background(),
		&ConfigBuildRequest{OrganizationID: 1, UserID: 2, RepositoryID: &repoID},
		"claude-code",
		[]string{"keep-me"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("len = %d, want 1", len(resources))
	}
	if resources[0].Sha != "sha1" {
		t.Fatalf("sha = %q, want sha1", resources[0].Sha)
	}
}

func TestSkillTargetPathMatchesAgentDiscoveryRoots(t *testing.T) {
	assert.Equal(
		t,
		"{{.sandbox.root_path}}/codex-home/skills/pattern-generate",
		skillTargetPath("pattern-designer", "pattern-generate"),
	)
	for _, agentSlug := range []string{"do-agent", "seedance-expert"} {
		assert.Equal(
			t,
			"{{.sandbox.work_dir}}/.agent/skills/seedance-expert",
			skillTargetPath(agentSlug, "seedance-expert"),
			agentSlug,
		)
	}
}

type nilAgentConfigProvider struct{}

func (nilAgentConfigProvider) GetAgent(context.Context, string) (*agentdomain.Agent, error) {
	return nil, nil
}
