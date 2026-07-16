package workbench

import (
	"bytes"
	"encoding/json"
	"fmt"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
)

func declaredArtifactManifest(
	raw []byte,
	representations map[string]declaredArtifactRepresentation,
) (*agentworkbenchv2.ArtifactManifest, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, nil
	}
	var kind artifactDeclarationManifestKind
	if err := json.Unmarshal(raw, &kind); err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}
	switch kind.Kind {
	case "image_edit":
		return declaredImageEditManifest(raw, representations)
	case "video":
		return declaredVideoManifest(raw, representations)
	case "presentation":
		return declaredPresentationManifest(raw, representations)
	default:
		return nil, fmt.Errorf("manifest.kind %q is unsupported", kind.Kind)
	}
}

func requiredRepresentation(
	representations map[string]declaredArtifactRepresentation,
	field string,
	id string,
) (declaredArtifactRepresentation, error) {
	if id == "" {
		return declaredArtifactRepresentation{}, fmt.Errorf("%s is required", field)
	}
	representation, exists := representations[id]
	if !exists {
		return declaredArtifactRepresentation{}, fmt.Errorf(
			"%s %q does not reference a representation",
			field,
			id,
		)
	}
	return representation, nil
}

func optionalRepresentation(
	representations map[string]declaredArtifactRepresentation,
	field string,
	id *string,
) (declaredArtifactRepresentation, bool, error) {
	if id == nil {
		return declaredArtifactRepresentation{}, false, nil
	}
	representation, err := requiredRepresentation(representations, field, *id)
	return representation, true, err
}

func validateRepresentationIDs(
	representations map[string]declaredArtifactRepresentation,
	field string,
	ids []string,
	media func(string) bool,
) error {
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			return fmt.Errorf("%s contains duplicate representation %q", field, id)
		}
		seen[id] = struct{}{}
		representation, err := requiredRepresentation(representations, field, id)
		if err != nil {
			return err
		}
		if media != nil && !media(representation.file.mediaType) {
			return fmt.Errorf(
				"%s representation %q has invalid media_type %q",
				field,
				id,
				representation.file.mediaType,
			)
		}
	}
	return nil
}

func safeDeclaredImage(mediaType string) bool {
	switch mediaType {
	case "image/avif", "image/gif", "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

func declaredVideo(mediaType string) bool {
	return len(mediaType) > len("video/") && mediaType[:len("video/")] == "video/"
}

func normalizedValue(value float64) bool {
	return value >= 0 && value <= 1
}

func declarationDimensions(
	value *artifactDeclarationDimensions,
) *agentworkbenchv2.ArtifactDimensions {
	if value == nil {
		return nil
	}
	return &agentworkbenchv2.ArtifactDimensions{
		Width: value.Width, Height: value.Height,
	}
}
