package workbench

import (
	"encoding/json"
	"fmt"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
)

func declaredImageEditManifest(
	raw []byte,
	representations map[string]declaredArtifactRepresentation,
) (*agentworkbenchv2.ArtifactManifest, error) {
	var declaration imageEditDeclarationManifest
	if err := decodeStrictJSON(raw, &declaration); err != nil {
		return nil, fmt.Errorf("image_edit manifest: %w", err)
	}
	source, err := requiredRepresentation(
		representations,
		"source_representation_id",
		declaration.SourceRepresentationID,
	)
	if err != nil {
		return nil, err
	}
	if !safeDeclaredImage(source.file.mediaType) {
		return nil, fmt.Errorf("source_representation_id must reference a safe image")
	}
	if declaration.SourceWidth == 0 || declaration.SourceHeight == 0 {
		return nil, fmt.Errorf("source_width and source_height must be positive")
	}
	if result, exists, err := optionalRepresentation(
		representations,
		"result_representation_id",
		declaration.ResultRepresentationID,
	); err != nil {
		return nil, err
	} else if exists && !safeDeclaredImage(result.file.mediaType) {
		return nil, fmt.Errorf("result_representation_id must reference a safe image")
	}
	if mask, exists, err := optionalRepresentation(
		representations,
		"mask_representation_id",
		declaration.MaskRepresentationID,
	); err != nil {
		return nil, err
	} else if exists && !safeDeclaredImage(mask.file.mediaType) {
		return nil, fmt.Errorf("mask_representation_id must reference a safe image")
	}
	if err := validateRepresentationIDs(
		representations,
		"candidate_representation_ids",
		declaration.CandidateRepresentationIDs,
		safeDeclaredImage,
	); err != nil {
		return nil, err
	}
	regions, err := declaredImageRegions(declaration.Regions)
	if err != nil {
		return nil, err
	}
	annotations, err := declaredImageAnnotations(declaration.Annotations)
	if err != nil {
		return nil, err
	}
	manifest := &agentworkbenchv2.ImageEditManifest{
		SourceRepresentationId:     declaration.SourceRepresentationID,
		ResultRepresentationId:     declaration.ResultRepresentationID,
		CandidateRepresentationIds: declaration.CandidateRepresentationIDs,
		MaskRepresentationId:       declaration.MaskRepresentationID,
		SourceWidth:                declaration.SourceWidth,
		SourceHeight:               declaration.SourceHeight,
		ExifOrientation:            declaration.ExifOrientation,
		Regions:                    regions,
		Annotations:                annotations,
	}
	return &agentworkbenchv2.ArtifactManifest{
		Manifest: &agentworkbenchv2.ArtifactManifest_ImageEdit{
			ImageEdit: manifest,
		},
	}, nil
}

func declaredImageRegions(
	values []artifactDeclarationRegion,
) ([]*agentworkbenchv2.NormalizedRegion, error) {
	regions := make([]*agentworkbenchv2.NormalizedRegion, 0, len(values))
	for index, value := range values {
		if !normalizedValue(value.X) || !normalizedValue(value.Y) ||
			value.Width <= 0 || value.Height <= 0 ||
			value.X+value.Width > 1 || value.Y+value.Height > 1 {
			return nil, fmt.Errorf("regions[%d] is outside normalized bounds", index)
		}
		regions = append(regions, &agentworkbenchv2.NormalizedRegion{
			X: value.X, Y: value.Y, Width: value.Width, Height: value.Height,
		})
	}
	return regions, nil
}

func declaredImageAnnotations(
	values []artifactDeclarationAnnotation,
) ([]*agentworkbenchv2.ImageAnnotation, error) {
	annotations := make([]*agentworkbenchv2.ImageAnnotation, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		if len(value.AnnotationID) < 2 || len(value.AnnotationID) > 100 ||
			!artifactDeclarationIdentifier.MatchString(value.AnnotationID) {
			return nil, fmt.Errorf("annotations[%d].annotation_id is invalid", index)
		}
		if _, exists := seen[value.AnnotationID]; exists {
			return nil, fmt.Errorf("annotation_id %q is duplicated", value.AnnotationID)
		}
		seen[value.AnnotationID] = struct{}{}
		path := make([]*agentworkbenchv2.NormalizedPoint, 0, len(value.Path))
		for pointIndex, point := range value.Path {
			if !normalizedValue(point.X) || !normalizedValue(point.Y) {
				return nil, fmt.Errorf(
					"annotations[%d].path[%d] is outside normalized bounds",
					index,
					pointIndex,
				)
			}
			path = append(path, &agentworkbenchv2.NormalizedPoint{X: point.X, Y: point.Y})
		}
		var style *agentworkbenchv2.StructuredPayload
		if len(value.Style) > 0 {
			if !json.Valid(value.Style) {
				return nil, fmt.Errorf("annotations[%d].style is invalid JSON", index)
			}
			style = rawPayload("application/json", string(value.Style))
		}
		annotations = append(annotations, &agentworkbenchv2.ImageAnnotation{
			AnnotationId: value.AnnotationID, Path: path, Label: value.Label, Style: style,
		})
	}
	return annotations, nil
}
