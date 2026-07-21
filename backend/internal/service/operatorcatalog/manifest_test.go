package operatorcatalog

import (
	"strings"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

func TestVideoExpertManifestIsCompleteAndInternallyConsistent(t *testing.T) {
	skills, err := Skills()
	require.NoError(t, err)
	require.Len(t, skills, 7)
	require.Len(t, Experts(), 3)

	skillSlugs := make(map[string]struct{}, len(skills))
	for _, skill := range skills {
		require.NoError(t, slugkit.Validate(skill.Slug))
		require.NotEmpty(t, strings.TrimSpace(skill.Name))
		require.Equal(t, "Apache-2.0", skill.License)
		require.Contains(t, skill.Tags, "video")
		require.NotEmpty(t, strings.TrimSpace(skill.Instructions))
		require.NotContains(t, skill.Instructions, "TODO")
		require.NotContains(t, skillSlugs, skill.Slug)
		skillSlugs[skill.Slug] = struct{}{}
		for _, source := range skill.ResearchSources {
			require.True(t, strings.HasPrefix(source.URL, "https://github.com/"))
			require.Len(t, source.Commit, 40)
			require.NotEmpty(t, source.License)
		}
	}
	expertSlugs := map[string]struct{}{}
	for _, expert := range Experts() {
		require.NoError(t, slugkit.Validate(expert.Slug))
		require.NotContains(t, expertSlugs, expert.Slug)
		expertSlugs[expert.Slug] = struct{}{}
		require.Equal(t, "video", expert.Category)
		require.NotEmpty(t, expert.Prompt)
		require.NotEmpty(t, expert.Outcomes)
		require.NotEmpty(t, expert.SkillSlugs)
		for _, skillSlug := range expert.SkillSlugs {
			require.Contains(t, skillSlugs, skillSlug)
		}
	}
}
