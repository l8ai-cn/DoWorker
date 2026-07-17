package postgres

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildExpertCatalogPayloadEmbedsInstallableSnapshot(t *testing.T) {
	payload, err := buildExpertCatalogPayload(publishedExpertRelease{
		ApplicationID: 3,
		ReleaseID:     7,
		Version:       2,
		ExpertSnapshot: []byte(`{
			"version":1,
			"agent_slug":"video-studio",
			"skill_slugs":["video-editing","video-qa"]
		}`),
		WorkerSpecSnapshot: []byte(`{
			"version":1,
			"spec":{"version":1},
			"summary":{"worker_type":{"slug":"video-studio"}}
		}`),
	})

	require.NoError(t, err)
	require.Len(t, payload.ContentDigest, 64)
	require.Equal(t, "video-studio", payload.AgentSlug)
	require.JSONEq(t, `{"agents":["video-studio"],"locale":"zh-CN"}`,
		string(payload.Compatibility))
	require.JSONEq(t, `{"skills":["video-editing","video-qa"]}`,
		string(payload.DependencyLock))

	var manifest struct {
		InstallationCredits string `json:"installation_credits"`
		SourceRelease       struct {
			ApplicationID int64 `json:"application_id"`
			ReleaseID     int64 `json:"release_id"`
			Version       int   `json:"version"`
		} `json:"source_release"`
		RuntimeSnapshot struct {
			Version    int             `json:"version"`
			Expert     json.RawMessage `json:"expert"`
			WorkerSpec json.RawMessage `json:"worker_spec"`
		} `json:"runtime_snapshot"`
	}
	require.NoError(t, json.Unmarshal(payload.Manifest, &manifest))
	require.Equal(t, "20", manifest.InstallationCredits)
	require.Equal(t, int64(3), manifest.SourceRelease.ApplicationID)
	require.Equal(t, int64(7), manifest.SourceRelease.ReleaseID)
	require.Equal(t, 2, manifest.SourceRelease.Version)
	require.Equal(t, 1, manifest.RuntimeSnapshot.Version)
	require.JSONEq(t, `{"version":1}`, string(manifest.RuntimeSnapshot.WorkerSpec))
}

func TestBuildExpertCatalogPayloadRejectsInvalidWorkerSpec(t *testing.T) {
	_, err := buildExpertCatalogPayload(publishedExpertRelease{
		ReleaseID:          8,
		ExpertSnapshot:     []byte(`{"agent_slug":"video-studio"}`),
		WorkerSpecSnapshot: []byte(`{"version":1,"spec":null}`),
	})

	require.EqualError(t, err, "expert release 8 has an invalid worker spec")
}
