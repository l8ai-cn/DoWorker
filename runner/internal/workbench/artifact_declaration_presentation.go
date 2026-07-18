package workbench

import (
	"fmt"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
)

func declaredPresentationManifest(
	raw []byte,
	representations map[string]declaredArtifactRepresentation,
) (*agentworkbenchv2.ArtifactManifest, error) {
	var declaration presentationDeclarationManifest
	if err := decodeStrictJSON(raw, &declaration); err != nil {
		return nil, fmt.Errorf("presentation manifest: %w", err)
	}
	if declaration.DeckRevision == 0 {
		return nil, fmt.Errorf("deck_revision must be positive")
	}
	if len(declaration.Slides) == 0 {
		return nil, fmt.Errorf("presentation slides are required")
	}
	slides, err := declaredPresentationSlides(declaration.Slides, representations)
	if err != nil {
		return nil, err
	}
	versions, err := declaredPresentationVersions(declaration.Versions)
	if err != nil {
		return nil, err
	}
	if declaration.SelectedVersionID != nil {
		found := false
		for _, version := range versions {
			if version.GetVersionId() == *declaration.SelectedVersionID {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("selected_version_id does not reference a version")
		}
	}
	manifest := &agentworkbenchv2.PresentationManifest{
		DeckRevision:      declaration.DeckRevision,
		Slides:            slides,
		Versions:          versions,
		SelectedVersionId: declaration.SelectedVersionID,
	}
	return &agentworkbenchv2.ArtifactManifest{
		Manifest: &agentworkbenchv2.ArtifactManifest_Presentation{
			Presentation: manifest,
		},
	}, nil
}

func declaredPresentationSlides(
	values []presentationDeclarationSlide,
	representations map[string]declaredArtifactRepresentation,
) ([]*agentworkbenchv2.PresentationSlide, error) {
	slides := make([]*agentworkbenchv2.PresentationSlide, 0, len(values))
	seenIDs := make(map[string]struct{}, len(values))
	seenPositions := make(map[uint32]struct{}, len(values))
	for index, value := range values {
		if len(value.SlideID) < 2 || len(value.SlideID) > 100 ||
			!artifactDeclarationIdentifier.MatchString(value.SlideID) {
			return nil, fmt.Errorf("slides[%d].slide_id is invalid", index)
		}
		if _, exists := seenIDs[value.SlideID]; exists {
			return nil, fmt.Errorf("slide_id %q is duplicated", value.SlideID)
		}
		seenIDs[value.SlideID] = struct{}{}
		if value.Position == 0 {
			return nil, fmt.Errorf("slides[%d].position must be positive", index)
		}
		if _, exists := seenPositions[value.Position]; exists {
			return nil, fmt.Errorf("slide position %d is duplicated", value.Position)
		}
		seenPositions[value.Position] = struct{}{}
		page, err := requiredRepresentation(
			representations,
			"page_representation_id",
			value.PageRepresentationID,
		)
		if err != nil {
			return nil, err
		}
		if !safeDeclaredImage(page.file.mediaType) {
			return nil, fmt.Errorf("page_representation_id must reference a safe image")
		}
		if err := validateOptionalRepresentationMedia(
			representations,
			"thumbnail_representation_id",
			value.ThumbnailRepresentationID,
			safeDeclaredImage,
		); err != nil {
			return nil, err
		}
		slides = append(slides, &agentworkbenchv2.PresentationSlide{
			SlideId: value.SlideID, Position: value.Position,
			Title: value.Title, Notes: value.Notes,
			PageRepresentationId:      optionalString(value.PageRepresentationID),
			ThumbnailRepresentationId: value.ThumbnailRepresentationID,
		})
	}
	return slides, nil
}

func declaredPresentationVersions(
	values []presentationDeclarationVersion,
) ([]*agentworkbenchv2.PresentationVersion, error) {
	versions := make([]*agentworkbenchv2.PresentationVersion, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		if len(value.VersionID) < 2 || len(value.VersionID) > 100 ||
			!artifactDeclarationIdentifier.MatchString(value.VersionID) {
			return nil, fmt.Errorf("versions[%d].version_id is invalid", index)
		}
		if value.Revision == 0 {
			return nil, fmt.Errorf("versions[%d].revision must be positive", index)
		}
		if _, exists := seen[value.VersionID]; exists {
			return nil, fmt.Errorf("version_id %q is duplicated", value.VersionID)
		}
		seen[value.VersionID] = struct{}{}
		versions = append(versions, &agentworkbenchv2.PresentationVersion{
			VersionId: value.VersionID, Revision: value.Revision, Label: value.Label,
		})
	}
	return versions, nil
}

func optionalString(value string) *string {
	return &value
}
