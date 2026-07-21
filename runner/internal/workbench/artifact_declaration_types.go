package workbench

import (
	"encoding/json"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
)

const (
	artifactDeclarationDirectory = ".agent-cloud/workbench/artifacts"
	artifactDeclarationSchema    = "agentcloud.agent-workbench.artifact/v1"
)

type artifactDeclaration struct {
	SchemaVersion           string                              `json:"schema_version"`
	ArtifactID              string                              `json:"artifact_id"`
	Revision                uint64                              `json:"revision"`
	Role                    string                              `json:"role"`
	PrimaryRepresentationID string                              `json:"primary_representation_id"`
	Producer                artifactDeclarationProducer         `json:"producer"`
	Representations         []artifactDeclarationRepresentation `json:"representations"`
	Manifest                json.RawMessage                     `json:"manifest,omitempty"`
}

type artifactDeclarationProducer struct {
	Namespace       string `json:"namespace"`
	Type            string `json:"type"`
	ID              string `json:"id,omitempty"`
	CommandID       string `json:"command_id,omitempty"`
	ToolExecutionID string `json:"tool_execution_id,omitempty"`
}

type artifactDeclarationRepresentation struct {
	RepresentationID string                         `json:"representation_id"`
	Path             string                         `json:"path"`
	MediaType        string                         `json:"media_type"`
	Role             string                         `json:"role,omitempty"`
	Dimensions       *artifactDeclarationDimensions `json:"dimensions,omitempty"`
	DurationMillis   *uint64                        `json:"duration_millis,omitempty"`
}

type artifactDeclarationDimensions struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
}

type declaredArtifact struct {
	artifactID              string
	revision                uint64
	role                    string
	primaryRepresentationID string
	producer                artifactDeclarationProducer
	representations         []declaredArtifactRepresentation
	manifest                *agentworkbenchv2.ArtifactManifest
	fingerprint             string
}

type declaredArtifactRepresentation struct {
	representationID string
	role             string
	file             artifactFile
	dimensions       *agentworkbenchv2.ArtifactDimensions
	durationMillis   *uint64
}

type emittedDeclaredArtifact struct {
	artifact declaredArtifact
	revision uint64
}

type artifactDeclarationManifestKind struct {
	Kind string `json:"kind"`
}

type imageEditDeclarationManifest struct {
	Kind                       string                          `json:"kind"`
	SourceRepresentationID     string                          `json:"source_representation_id"`
	ResultRepresentationID     *string                         `json:"result_representation_id,omitempty"`
	CandidateRepresentationIDs []string                        `json:"candidate_representation_ids,omitempty"`
	MaskRepresentationID       *string                         `json:"mask_representation_id,omitempty"`
	SourceWidth                uint32                          `json:"source_width"`
	SourceHeight               uint32                          `json:"source_height"`
	ExifOrientation            *string                         `json:"exif_orientation,omitempty"`
	Regions                    []artifactDeclarationRegion     `json:"regions,omitempty"`
	Annotations                []artifactDeclarationAnnotation `json:"annotations,omitempty"`
}

type artifactDeclarationRegion struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type artifactDeclarationAnnotation struct {
	AnnotationID string                     `json:"annotation_id"`
	Path         []artifactDeclarationPoint `json:"path"`
	Label        *string                    `json:"label,omitempty"`
	Style        json.RawMessage            `json:"style,omitempty"`
}

type artifactDeclarationPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type videoDeclarationManifest struct {
	Kind                        string                         `json:"kind"`
	Stage                       string                         `json:"stage"`
	ProgressFraction            *float64                       `json:"progress_fraction,omitempty"`
	DurationMillis              *uint64                        `json:"duration_millis,omitempty"`
	Dimensions                  *artifactDeclarationDimensions `json:"dimensions,omitempty"`
	OriginalRepresentationID    *string                        `json:"original_representation_id,omitempty"`
	PlayableRepresentationID    *string                        `json:"playable_representation_id,omitempty"`
	PosterRepresentationID      *string                        `json:"poster_representation_id,omitempty"`
	ThumbnailRepresentationIDs  []string                       `json:"thumbnail_representation_ids,omitempty"`
	DerivativeRepresentationIDs []string                       `json:"derivative_representation_ids,omitempty"`
}

type presentationDeclarationManifest struct {
	Kind              string                           `json:"kind"`
	DeckRevision      uint64                           `json:"deck_revision"`
	Slides            []presentationDeclarationSlide   `json:"slides"`
	Versions          []presentationDeclarationVersion `json:"versions,omitempty"`
	SelectedVersionID *string                          `json:"selected_version_id,omitempty"`
}

type presentationDeclarationSlide struct {
	SlideID                   string  `json:"slide_id"`
	Position                  uint32  `json:"position"`
	Title                     *string `json:"title,omitempty"`
	Notes                     *string `json:"notes,omitempty"`
	PageRepresentationID      string  `json:"page_representation_id"`
	ThumbnailRepresentationID *string `json:"thumbnail_representation_id,omitempty"`
}

type presentationDeclarationVersion struct {
	VersionID string  `json:"version_id"`
	Revision  uint64  `json:"revision"`
	Label     *string `json:"label,omitempty"`
}
