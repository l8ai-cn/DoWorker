package catalog

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/Masterminds/semver/v3"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type ValidationStatus string

const (
	ValidationPending ValidationStatus = "pending"
	ValidationPassed  ValidationStatus = "passed"
)

var (
	ErrInvalidVersion       = errors.New("invalid catalog version")
	ErrInvalidDigest        = errors.New("invalid content digest")
	ErrInvalidManifest      = errors.New("invalid catalog manifest")
	ErrInvalidCompatibility = errors.New("invalid catalog compatibility")
)

type Version struct {
	id                      int64
	catalogItemID           int64
	version                 string
	sourceRevision          string
	contentDigest           string
	manifest                json.RawMessage
	compatibility           json.RawMessage
	validationStatus        ValidationStatus
	createdByPlatformUserID int64
}

type VersionState struct {
	ID                      int64
	CatalogItemID           int64
	Version                 string
	SourceRevision          string
	ContentDigest           string
	Manifest                json.RawMessage
	Compatibility           json.RawMessage
	ValidationStatus        ValidationStatus
	CreatedByPlatformUserID int64
}

func NewVersion(
	itemID int64,
	version string,
	sourceRevision string,
	contentDigest string,
	manifest json.RawMessage,
	compatibility json.RawMessage,
	actorUserID int64,
) (*Version, error) {
	if itemID <= 0 || actorUserID <= 0 || sourceRevision == "" {
		return nil, ErrInvalidVersion
	}
	if _, err := semver.StrictNewVersion(version); err != nil {
		return nil, ErrInvalidVersion
	}
	digest, err := hex.DecodeString(contentDigest)
	if err != nil || len(digest) != 32 {
		return nil, ErrInvalidDigest
	}
	if !json.Valid(manifest) {
		return nil, ErrInvalidManifest
	}
	if err := validateCompatibility(compatibility); err != nil {
		return nil, err
	}
	return &Version{
		catalogItemID:           itemID,
		version:                 version,
		sourceRevision:          sourceRevision,
		contentDigest:           contentDigest,
		manifest:                append(json.RawMessage(nil), manifest...),
		compatibility:           append(json.RawMessage(nil), compatibility...),
		validationStatus:        ValidationPending,
		createdByPlatformUserID: actorUserID,
	}, nil
}

func validateCompatibility(raw json.RawMessage) error {
	var compatibility struct {
		Agents []string `json:"agents"`
	}
	if err := json.Unmarshal(raw, &compatibility); err != nil {
		return ErrInvalidCompatibility
	}
	for _, agent := range compatibility.Agents {
		if slugkit.Validate(agent) != nil {
			return ErrInvalidCompatibility
		}
	}
	return nil
}

func RestoreVersion(state VersionState) (*Version, error) {
	version, err := NewVersion(
		state.CatalogItemID,
		state.Version,
		state.SourceRevision,
		state.ContentDigest,
		state.Manifest,
		state.Compatibility,
		state.CreatedByPlatformUserID,
	)
	if err != nil {
		return nil, err
	}
	if state.ValidationStatus != ValidationPending &&
		state.ValidationStatus != ValidationPassed {
		return nil, errors.New("invalid validation status")
	}
	version.id = state.ID
	version.validationStatus = state.ValidationStatus
	return version, nil
}

func (v *Version) MarkValidationPassed() {
	v.validationStatus = ValidationPassed
}

func (v Version) ID() int64                          { return v.id }
func (v Version) CatalogItemID() int64               { return v.catalogItemID }
func (v Version) Version() string                    { return v.version }
func (v Version) SourceRevision() string             { return v.sourceRevision }
func (v Version) ContentDigest() string              { return v.contentDigest }
func (v Version) ValidationStatus() ValidationStatus { return v.validationStatus }
func (v Version) CreatedByPlatformUserID() int64     { return v.createdByPlatformUserID }
func (v Version) Manifest() json.RawMessage {
	return append(json.RawMessage(nil), v.manifest...)
}
func (v Version) Compatibility() json.RawMessage {
	return append(json.RawMessage(nil), v.compatibility...)
}
