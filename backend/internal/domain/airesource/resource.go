package airesource

import (
	"fmt"
	"strings"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type Modality string

const (
	ModalityChat       Modality = "chat"
	ModalityImage      Modality = "image"
	ModalityAudio      Modality = "audio"
	ModalityVideo      Modality = "video"
	ModalityEmbedding  Modality = "embedding"
	ModalityMultimodal Modality = "multimodal"
)

type Capability string

const (
	CapabilityTextGeneration  Capability = "text-generation"
	CapabilityVisionInput     Capability = "vision-input"
	CapabilityImageGeneration Capability = "image-generation"
	CapabilitySpeechToText    Capability = "speech-to-text"
	CapabilityTextToSpeech    Capability = "text-to-speech"
	CapabilityVideoGeneration Capability = "video-generation"
	CapabilityEmbedding       Capability = "embedding"
)

type ModelResource struct {
	ID                   int64            `json:"id"`
	ProviderConnectionID int64            `json:"provider_connection_id"`
	Identifier           slugkit.Slug     `json:"identifier"`
	ModelID              string           `json:"model_id"`
	DisplayName          string           `json:"display_name"`
	Modalities           []Modality       `json:"modalities"`
	Capabilities         []Capability     `json:"capabilities"`
	DefaultModalities    []Modality       `json:"default_modalities"`
	Status               ConnectionStatus `json:"status"`
	LastValidatedAt      *time.Time       `json:"last_validated_at,omitempty"`
	ValidationError      string           `json:"validation_error,omitempty"`
	UsageSummary         *UsageSummary    `json:"usage_summary,omitempty"`
	IsEnabled            bool             `json:"is_enabled"`
	Revision             int64            `json:"-"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

type UsageSummary struct {
	QuotaTotal *float64   `json:"quota_total,omitempty"`
	UsageTotal *float64   `json:"usage_total,omitempty"`
	Remaining  *float64   `json:"remaining,omitempty"`
	Unit       *string    `json:"unit,omitempty"`
	Period     *string    `json:"period,omitempty"`
	MeasuredAt *time.Time `json:"measured_at,omitempty"`
}

func (modality Modality) Valid() bool {
	switch modality {
	case ModalityChat, ModalityImage, ModalityAudio, ModalityVideo, ModalityEmbedding, ModalityMultimodal:
		return true
	default:
		return false
	}
}

func (capability Capability) Valid() bool {
	switch capability {
	case CapabilityTextGeneration, CapabilityVisionInput, CapabilityImageGeneration,
		CapabilitySpeechToText, CapabilityTextToSpeech, CapabilityVideoGeneration,
		CapabilityEmbedding:
		return true
	default:
		return false
	}
}

func ValidateModelResource(resource ModelResource) error {
	if err := slugkit.Validate(resource.Identifier.String()); err != nil {
		return fmt.Errorf("model resource identifier: %w", err)
	}
	if strings.TrimSpace(resource.ModelID) == "" {
		return fmt.Errorf("model resource %q has no provider model ID", resource.Identifier)
	}
	if len(resource.Modalities) == 0 {
		return fmt.Errorf("model resource %q has no modalities", resource.Identifier)
	}
	supportedModalities := make(map[Modality]struct{}, len(resource.Modalities))
	for _, modality := range resource.Modalities {
		if !modality.Valid() {
			return fmt.Errorf("model resource %q has invalid modality %q", resource.Identifier, modality)
		}
		supportedModalities[modality] = struct{}{}
	}
	defaultModalities := make(map[Modality]struct{}, len(resource.DefaultModalities))
	for _, modality := range resource.DefaultModalities {
		if !modality.Valid() {
			return fmt.Errorf("model resource %q has invalid default modality %q", resource.Identifier, modality)
		}
		if _, exists := defaultModalities[modality]; exists {
			return fmt.Errorf("model resource %q has duplicate default modality %q", resource.Identifier, modality)
		}
		if _, supported := supportedModalities[modality]; !supported {
			return fmt.Errorf("model resource %q defaults unsupported modality %q", resource.Identifier, modality)
		}
		defaultModalities[modality] = struct{}{}
	}
	for _, capability := range resource.Capabilities {
		if !capability.Valid() {
			return fmt.Errorf("model resource %q has invalid capability %q", resource.Identifier, capability)
		}
	}
	return nil
}

func (resource ModelResource) ValidateIdentifiers() error {
	return ValidateModelResource(resource)
}
