package workerdependency

import (
	"sort"
	"strconv"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
)

func NormalizeAndValidate(document Document) (Document, error) {
	normalized := Normalize(document)
	if err := Validate(normalized); err != nil {
		return Document{}, err
	}
	return normalized, nil
}

func Normalize(document Document) Document {
	normalized := document
	normalized.Worker.SpecDigest = strings.TrimSpace(document.Worker.SpecDigest)
	normalized.Worker.DefinitionHash = strings.TrimSpace(document.Worker.DefinitionHash)
	normalized.Worker.ModelManagedFields = normalizeFields(
		document.Worker.ModelManagedFields,
	)
	normalized.Worker.CredentialBundleFields = normalizeFields(
		document.Worker.CredentialBundleFields,
	)
	normalized.Worker.AgentfileSourceDigest = strings.TrimSpace(
		document.Worker.AgentfileSourceDigest,
	)
	normalized.Models = normalizeModels(document.Models)
	normalized.Repository = normalizeRepository(document.Repository)
	normalized.Skills = append([]Skill{}, document.Skills...)
	for index := range normalized.Skills {
		skill := &normalized.Skills[index]
		skill.Pin = normalizePin(skill.Pin)
		skill.ContentDigest = strings.TrimSpace(skill.ContentDigest)
		skill.StorageKey = strings.TrimSpace(skill.StorageKey)
	}
	sort.Slice(normalized.Skills, func(left, right int) bool {
		return pinKey(normalized.Skills[left].Pin) < pinKey(normalized.Skills[right].Pin)
	})
	normalized.KnowledgeBases = append([]KnowledgeBase{}, document.KnowledgeBases...)
	for index := range normalized.KnowledgeBases {
		item := &normalized.KnowledgeBases[index]
		item.Pin = normalizePin(item.Pin)
		item.HTTPCloneURL = strings.TrimSpace(item.HTTPCloneURL)
		item.Branch = strings.TrimSpace(item.Branch)
		item.CommitSHA = strings.TrimSpace(item.CommitSHA)
	}
	sort.Slice(normalized.KnowledgeBases, func(left, right int) bool {
		return pinKey(normalized.KnowledgeBases[left].Pin) <
			pinKey(normalized.KnowledgeBases[right].Pin)
	})
	normalized.RuntimeBundles = append([]RuntimeBundle{}, document.RuntimeBundles...)
	for index := range normalized.RuntimeBundles {
		bundle := &normalized.RuntimeBundles[index]
		bundle.Pin = normalizePin(bundle.Pin)
		bundle.Kind = strings.TrimSpace(bundle.Kind)
		bundle.ContentDigest = strings.TrimSpace(bundle.ContentDigest)
		bundle.Values = normalizeRuntimeValues(bundle.Values)
		if bundle.ConfigDocument != nil {
			copy := *bundle.ConfigDocument
			copy.ID = strings.TrimSpace(copy.ID)
			copy.Format = strings.TrimSpace(copy.Format)
			copy.TargetPath = strings.TrimSpace(copy.TargetPath)
			bundle.ConfigDocument = &copy
		}
	}
	normalized.SecretReferences = append(
		[]SecretReference{},
		document.SecretReferences...,
	)
	for index := range normalized.SecretReferences {
		reference := &normalized.SecretReferences[index]
		reference.Pin = normalizePin(reference.Pin)
		reference.Field = strings.TrimSpace(reference.Field)
		reference.BundleKey = strings.TrimSpace(reference.BundleKey)
		reference.OwnerScope = strings.TrimSpace(reference.OwnerScope)
	}
	sort.Slice(normalized.SecretReferences, func(left, right int) bool {
		if normalized.SecretReferences[left].Field !=
			normalized.SecretReferences[right].Field {
			return normalized.SecretReferences[left].Field <
				normalized.SecretReferences[right].Field
		}
		return pinKey(normalized.SecretReferences[left].Pin) <
			pinKey(normalized.SecretReferences[right].Pin)
	})
	normalized.Placement = normalizePlacement(document.Placement)
	return normalized
}

func normalizeModels(models Models) Models {
	normalized := Models{Tools: append([]ToolModel{}, models.Tools...)}
	if models.Primary != nil {
		copy := normalizeModel(*models.Primary)
		normalized.Primary = &copy
	}
	for index := range normalized.Tools {
		normalized.Tools[index].Binding = normalizeReference(
			normalized.Tools[index].Binding,
		)
		normalized.Tools[index].Model = normalizeModel(normalized.Tools[index].Model)
	}
	sort.Slice(normalized.Tools, func(left, right int) bool {
		return normalized.Tools[left].Role < normalized.Tools[right].Role
	})
	return normalized
}

func normalizeModel(model Model) Model {
	model.Pin = normalizePin(model.Pin)
	model.ModelID = strings.TrimSpace(model.ModelID)
	model.BaseURL = strings.TrimSpace(model.BaseURL)
	model.Modalities = append([]airesource.Modality{}, model.Modalities...)
	model.Capabilities = append([]airesource.Capability{}, model.Capabilities...)
	sort.Slice(model.Modalities, func(left, right int) bool {
		return model.Modalities[left] < model.Modalities[right]
	})
	sort.Slice(model.Capabilities, func(left, right int) bool {
		return model.Capabilities[left] < model.Capabilities[right]
	})
	return model
}

func normalizeRepository(repository *Repository) *Repository {
	if repository == nil {
		return nil
	}
	copy := *repository
	copy.Pin = normalizePin(copy.Pin)
	copy.HTTPCloneURL = strings.TrimSpace(copy.HTTPCloneURL)
	copy.SSHCloneURL = strings.TrimSpace(copy.SSHCloneURL)
	copy.Branch = strings.TrimSpace(copy.Branch)
	copy.CommitSHA = strings.TrimSpace(copy.CommitSHA)
	copy.Credential.Type = strings.TrimSpace(copy.Credential.Type)
	copy.Credential.CredentialID = cloneInt64(copy.Credential.CredentialID)
	copy.PreparationScriptDigest = strings.TrimSpace(copy.PreparationScriptDigest)
	return &copy
}

func normalizeRuntimeValues(values []RuntimeValue) []RuntimeValue {
	normalized := append([]RuntimeValue{}, values...)
	for index := range normalized {
		normalized[index].Name = strings.TrimSpace(normalized[index].Name)
	}
	sort.Slice(normalized, func(left, right int) bool {
		return normalized[left].Name < normalized[right].Name
	})
	return normalized
}

func normalizePlacement(placement Placement) Placement {
	placement.CatalogRevision = strings.TrimSpace(placement.CatalogRevision)
	placement.RuntimeImage.Reference = strings.TrimSpace(
		placement.RuntimeImage.Reference,
	)
	placement.RuntimeImage.Digest = strings.TrimSpace(placement.RuntimeImage.Digest)
	placement.ComputeTarget = normalizePin(placement.ComputeTarget)
	if placement.ResourceProfile != nil {
		copy := normalizePin(*placement.ResourceProfile)
		placement.ResourceProfile = &copy
	}
	resources := placement.Spec.ResourceProfile.Resources
	placement.Spec.ResourceProfile.Resources.GPURequest = cloneUint32(resources.GPURequest)
	placement.Spec.ResourceProfile.Resources.GPULimit = cloneUint32(resources.GPULimit)
	return placement
}

func pinKey(pin ResourcePin) string {
	ref := pin.Reference
	return strings.Join([]string{
		ref.APIVersion, ref.Kind, ref.Namespace.String(), ref.Name.String(),
		ref.UID, strconv.FormatInt(ref.Revision, 10), ref.Digest,
		strconv.FormatInt(pin.DomainID, 10),
	}, "\x00")
}

func cloneUint32(value *uint32) *uint32 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
