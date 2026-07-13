package expertmarket

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReleaseValidateAcceptsPersistableSnapshot(t *testing.T) {
	release := validRelease()
	require.NoError(t, release.Validate())
}

func TestReleaseValidateRejectsInvalidStatus(t *testing.T) {
	release := validRelease()
	release.Status = ReleaseStatus("approved")
	require.ErrorIs(t, release.Validate(), ErrInvalidStatus)
}

func TestReleaseValidateRejectsInvalidVersion(t *testing.T) {
	release := validRelease()
	release.Version = 0
	require.ErrorIs(t, release.Validate(), ErrInvalidVersion)
}

func TestReleaseValidateRejectsInvalidJSONShapes(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Release)
	}{
		{"expert array", func(r *Release) { r.ExpertSnapshot = json.RawMessage(`[]`) }},
		{"expert missing version", func(r *Release) { r.ExpertSnapshot = json.RawMessage(`{"name":"x"}`) }},
		{"expert fractional version", func(r *Release) { r.ExpertSnapshot = json.RawMessage(`{"version":1.5}`) }},
		{"worker array", func(r *Release) { r.WorkerSpecSnapshot = json.RawMessage(`[]`) }},
		{"worker zero version", func(r *Release) { r.WorkerSpecSnapshot = json.RawMessage(`{"version":0}`) }},
		{"dependencies object", func(r *Release) { r.SkillDependencies = json.RawMessage(`{}`) }},
		{"invalid json", func(r *Release) { r.SkillDependencies = json.RawMessage(`[`) }},
		{"trailing json", func(r *Release) { r.ExpertSnapshot = json.RawMessage(`{"version":1}{}`) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			release := validRelease()
			tt.mutate(&release)
			require.ErrorIs(t, release.Validate(), ErrInvalidSnapshot)
		})
	}
}

func validRelease() Release {
	return Release{
		ApplicationID:           1,
		SourceExpertID:          2,
		PublisherOrganizationID: 3,
		PublisherUserID:         4,
		Version:                 1,
		Status:                  ReleaseStatusDraft,
		Name:                    "Video Expert",
		Tags:                    []string{"video"},
		Outcomes:                []string{"render"},
		ExpertSnapshot:          json.RawMessage(`{"version":1,"slug":"video-expert"}`),
		WorkerSpecSnapshot:      json.RawMessage(`{"version":1,"runtime":{}}`),
		SkillDependencies:       json.RawMessage(`[{"slug":"video-use","version":"abc"}]`),
	}
}
