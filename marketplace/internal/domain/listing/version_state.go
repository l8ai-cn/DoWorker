package listing

import (
	"encoding/json"
	"errors"
)

type VersionState struct {
	ID                   int64
	ListingID            int64
	CatalogItemVersionID int64
	Revision             int
	DisplayName          string
	Tagline              string
	Description          string
	Outcomes             json.RawMessage
	UseCases             json.RawMessage
	TargetAudience       json.RawMessage
	Requirements         json.RawMessage
	Tags                 []string
	ReleaseNotes         string
	ReviewStatus         ReviewStatus
}

func RestoreVersion(state VersionState) (*Version, error) {
	version, err := NewVersion(
		state.ListingID,
		state.CatalogItemVersionID,
		state.Revision,
		state.DisplayName,
		state.Tagline,
		state.Description,
		state.Outcomes,
		state.UseCases,
		state.TargetAudience,
		state.Requirements,
		state.Tags,
		state.ReleaseNotes,
	)
	if err != nil {
		return nil, err
	}
	if state.ReviewStatus != ReviewDraft &&
		state.ReviewStatus != ReviewSubmitted &&
		state.ReviewStatus != ReviewApproved &&
		state.ReviewStatus != ReviewRejected {
		return nil, errors.New("invalid listing review status")
	}
	version.id = state.ID
	version.reviewStatus = state.ReviewStatus
	return version, nil
}
