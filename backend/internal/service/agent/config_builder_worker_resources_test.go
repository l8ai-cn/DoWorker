package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/extension"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	envbundleservice "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
	extensionservice "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type exactBundleLoader struct {
	bundles []*envbundleservice.EffectiveBundle
	err     error
	ids     []int64
}

func (loader *exactBundleLoader) GetEffectiveForUser(
	context.Context,
	int64,
	int64,
	string,
) ([]*envbundleservice.EffectiveBundle, error) {
	return nil, nil
}

func (loader *exactBundleLoader) GetEffectiveByIDs(
	_ context.Context,
	_, _ int64,
	_ string,
	ids []int64,
) ([]*envbundleservice.EffectiveBundle, error) {
	loader.ids = append([]int64{}, ids...)
	return loader.bundles, loader.err
}

type exactSkillProvider struct {
	skills   []*extensionservice.ResolvedSkill
	err      error
	ids      []int64
	packages []specdomain.SkillPackageBinding
}

func (provider *exactSkillProvider) GetWorkerSkillsByPackages(
	_ context.Context,
	packages []specdomain.SkillPackageBinding,
	_ string,
) ([]*extensionservice.ResolvedSkill, error) {
	provider.packages = append([]specdomain.SkillPackageBinding{}, packages...)
	return provider.skills, provider.err
}

func (*exactSkillProvider) GetEffectiveMcpServers(
	context.Context,
	int64,
	int64,
	int64,
	string,
) ([]*extension.InstalledMcpServer, error) {
	return nil, nil
}

func (*exactSkillProvider) GetEffectiveSkills(
	context.Context,
	int64,
	int64,
	int64,
	string,
) ([]*extensionservice.ResolvedSkill, error) {
	return nil, nil
}

func (provider *exactSkillProvider) GetWorkerSkillsByIDs(
	_ context.Context,
	_ int64,
	ids []int64,
	_ string,
) ([]*extensionservice.ResolvedSkill, error) {
	provider.ids = append([]int64{}, ids...)
	return provider.skills, provider.err
}

func TestWorkerSpecEnvBundlesLoadByExactID(t *testing.T) {
	loader := &exactBundleLoader{
		bundles: []*envbundleservice.EffectiveBundle{
			{ID: 6, Name: "signing", Data: map[string]string{"TOKEN": "secret"}},
		},
	}
	builder := NewConfigBuilder(nilAgentConfigProvider{}, loader)

	contextMap, err := builder.buildEnvBundleContext(
		context.Background(),
		&ConfigBuildRequest{RequiredEnvBundleIDs: []int64{6}},
		"codex-cli",
	)

	require.NoError(t, err)
	assert.Equal(t, []int64{6}, loader.ids)
	assert.Equal(t, "secret", contextMap["signing"]["TOKEN"])
}

func TestWorkerSpecEnvBundleFailureIsNotIgnored(t *testing.T) {
	builder := NewConfigBuilder(nilAgentConfigProvider{}, &exactBundleLoader{
		err: errors.New("decrypt failed"),
	})

	_, err := builder.buildEnvBundleContext(
		context.Background(),
		&ConfigBuildRequest{RequiredEnvBundleIDs: []int64{6}},
		"codex-cli",
	)

	assert.ErrorContains(t, err, "decrypt failed")
}

func TestWorkerSpecSkillsLoadByExactCatalogID(t *testing.T) {
	provider := &exactSkillProvider{
		skills: []*extensionservice.ResolvedSkill{
			{
				CatalogSkillID: 3,
				Slug:           "reviewer",
				ContentSha:     "sha-reviewer",
				DownloadURL:    "https://example/reviewer",
			},
		},
	}
	builder := NewConfigBuilder(nilAgentConfigProvider{}, &exactBundleLoader{})
	builder.SetExtensionProvider(provider)

	resources, err := builder.buildSkillResources(
		context.Background(),
		&ConfigBuildRequest{RequiredSkillIDs: []int64{3}},
		"codex-cli",
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, []int64{3}, provider.ids)
	require.Len(t, resources, 1)
	assert.Equal(t, "sha-reviewer", resources[0].Sha)
}

func TestWorkerSpecSkillFailureIsNotIgnored(t *testing.T) {
	provider := &exactSkillProvider{err: errors.New("sign URL failed")}
	builder := NewConfigBuilder(nilAgentConfigProvider{}, &exactBundleLoader{})
	builder.SetExtensionProvider(provider)

	_, err := builder.buildSkillResources(
		context.Background(),
		&ConfigBuildRequest{RequiredSkillIDs: []int64{3}},
		"codex-cli",
		nil,
	)

	assert.ErrorContains(t, err, "sign URL failed")
}

func TestWorkerSpecSkillsLoadFromPinnedPackages(t *testing.T) {
	binding := specdomain.SkillPackageBinding{
		SkillID:     3,
		Slug:        "reviewer",
		Version:     2,
		ContentSHA:  "sha-reviewer",
		StorageKey:  "skills/reviewer-v2.tar.gz",
		PackageSize: 123,
	}
	provider := &exactSkillProvider{
		skills: []*extensionservice.ResolvedSkill{{
			CatalogSkillID: binding.SkillID,
			Slug:           binding.Slug,
			ContentSha:     binding.ContentSHA,
			DownloadURL:    "https://example/reviewer-v2",
			PackageSize:    binding.PackageSize,
		}},
	}
	builder := NewConfigBuilder(nilAgentConfigProvider{}, &exactBundleLoader{})
	builder.SetExtensionProvider(provider)

	resources, err := builder.buildSkillResources(
		context.Background(),
		&ConfigBuildRequest{RequiredSkillPackages: []specdomain.SkillPackageBinding{binding}},
		"codex-cli",
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, []specdomain.SkillPackageBinding{binding}, provider.packages)
	assert.Empty(t, provider.ids)
	require.Len(t, resources, 1)
	assert.Equal(t, binding.ContentSHA, resources[0].Sha)
}
