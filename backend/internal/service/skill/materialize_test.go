package skill

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

func TestMaterializeRepo_RoundTripAndCleanup(t *testing.T) {
	fake := gitops.NewFake("am-skills")
	_, err := fake.Provision(context.Background(), gitops.ProvisionParams{
		OrgID: 7, Slug: "web-search",
		Seed: []gitops.FileChange{
			{Path: "SKILL.md", Content: []byte("---\nname: web-search\n---\n# body")},
			{Path: "skill.json", Content: []byte(`{"slug":"web-search"}`)},
			{Path: "reference/notes.md", Content: []byte("nested content")},
		},
	})
	require.NoError(t, err)

	dir, cleanup, err := materializeRepo(context.Background(), fake, "org7-web-search", "main")
	require.NoError(t, err)
	require.NotEmpty(t, dir)

	md, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	require.NoError(t, err)
	assert.Contains(t, string(md), "name: web-search")

	nested, err := os.ReadFile(filepath.Join(dir, "reference", "notes.md"))
	require.NoError(t, err)
	assert.Equal(t, "nested content", string(nested))

	cleanup()
	_, statErr := os.Stat(dir)
	assert.True(t, os.IsNotExist(statErr), "cleanup must remove the temp dir")
}

func TestMaterializeRepo_RejectsUnsafePath(t *testing.T) {
	fake := gitops.NewFake("am-skills")
	// Seed a repo whose tree contains a traversal path.
	_, err := fake.Provision(context.Background(), gitops.ProvisionParams{
		OrgID: 7, Slug: "evil",
		Seed: []gitops.FileChange{{Path: "../escape.txt", Content: []byte("x")}},
	})
	require.NoError(t, err)

	_, cleanup, err := materializeRepo(context.Background(), fake, "org7-evil", "main")
	require.Error(t, err)
	assert.Nil(t, cleanup)
}
