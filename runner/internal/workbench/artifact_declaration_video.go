package workbench

import (
	"fmt"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
)

func declaredVideoManifest(
	raw []byte,
	representations map[string]declaredArtifactRepresentation,
) (*agentworkbenchv2.ArtifactManifest, error) {
	var declaration videoDeclarationManifest
	if err := decodeStrictJSON(raw, &declaration); err != nil {
		return nil, fmt.Errorf("video manifest: %w", err)
	}
	stage, err := declaredVideoStage(declaration.Stage)
	if err != nil {
		return nil, err
	}
	if stage != agentworkbenchv2.VideoStage_VIDEO_STAGE_READY {
		return nil, fmt.Errorf("published video manifest stage must be ready")
	}
	if declaration.ProgressFraction != nil &&
		(*declaration.ProgressFraction < 0 || *declaration.ProgressFraction > 1) {
		return nil, fmt.Errorf("progress_fraction must be between 0 and 1")
	}
	if declaration.Dimensions != nil &&
		(declaration.Dimensions.Width == 0 || declaration.Dimensions.Height == 0) {
		return nil, fmt.Errorf("video dimensions must be positive")
	}
	if err := validateOptionalRepresentationMedia(
		representations,
		"original_representation_id",
		declaration.OriginalRepresentationID,
		declaredVideo,
	); err != nil {
		return nil, err
	}
	if err := validateOptionalRepresentationMedia(
		representations,
		"playable_representation_id",
		declaration.PlayableRepresentationID,
		declaredVideo,
	); err != nil {
		return nil, err
	}
	if err := validateOptionalRepresentationMedia(
		representations,
		"poster_representation_id",
		declaration.PosterRepresentationID,
		safeDeclaredImage,
	); err != nil {
		return nil, err
	}
	if declaration.PlayableRepresentationID == nil {
		return nil, fmt.Errorf("ready video requires playable_representation_id")
	}
	if err := validateRepresentationIDs(
		representations,
		"thumbnail_representation_ids",
		declaration.ThumbnailRepresentationIDs,
		safeDeclaredImage,
	); err != nil {
		return nil, err
	}
	if err := validateRepresentationIDs(
		representations,
		"derivative_representation_ids",
		declaration.DerivativeRepresentationIDs,
		declaredVideo,
	); err != nil {
		return nil, err
	}
	manifest := &agentworkbenchv2.VideoManifest{
		Stage:                       stage,
		ProgressFraction:            declaration.ProgressFraction,
		DurationMillis:              declaration.DurationMillis,
		Dimensions:                  declarationDimensions(declaration.Dimensions),
		OriginalRepresentationId:    declaration.OriginalRepresentationID,
		PlayableRepresentationId:    declaration.PlayableRepresentationID,
		PosterRepresentationId:      declaration.PosterRepresentationID,
		ThumbnailRepresentationIds:  declaration.ThumbnailRepresentationIDs,
		DerivativeRepresentationIds: declaration.DerivativeRepresentationIDs,
	}
	return &agentworkbenchv2.ArtifactManifest{
		Manifest: &agentworkbenchv2.ArtifactManifest_Video{Video: manifest},
	}, nil
}

func declaredVideoStage(value string) (agentworkbenchv2.VideoStage, error) {
	switch value {
	case "queued":
		return agentworkbenchv2.VideoStage_VIDEO_STAGE_QUEUED, nil
	case "rendering":
		return agentworkbenchv2.VideoStage_VIDEO_STAGE_RENDERING, nil
	case "transcoding":
		return agentworkbenchv2.VideoStage_VIDEO_STAGE_TRANSCODING, nil
	case "ready":
		return agentworkbenchv2.VideoStage_VIDEO_STAGE_READY, nil
	case "failed":
		return agentworkbenchv2.VideoStage_VIDEO_STAGE_FAILED, nil
	default:
		return agentworkbenchv2.VideoStage_VIDEO_STAGE_UNSPECIFIED,
			fmt.Errorf("video stage %q is unsupported", value)
	}
}

func validateOptionalRepresentationMedia(
	representations map[string]declaredArtifactRepresentation,
	field string,
	id *string,
	media func(string) bool,
) error {
	representation, exists, err := optionalRepresentation(representations, field, id)
	if err != nil || !exists {
		return err
	}
	if !media(representation.file.mediaType) {
		return fmt.Errorf(
			"%s representation %q has invalid media_type %q",
			field,
			*id,
			representation.file.mediaType,
		)
	}
	return nil
}
