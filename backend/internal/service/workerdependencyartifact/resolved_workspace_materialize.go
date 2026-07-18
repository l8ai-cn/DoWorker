package workerdependencyartifact

import (
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
)

func materializeRepository(
	resolution *RepositoryResolution,
) *workerdependency.Repository {
	if resolution == nil {
		return nil
	}
	var credentialID *int64
	if resolution.CredentialID != nil {
		value := *resolution.CredentialID
		credentialID = &value
	}
	return &workerdependency.Repository{
		Pin:          pin(resolution.ResourceResolution),
		HTTPCloneURL: resolution.HTTPCloneURL,
		SSHCloneURL:  resolution.SSHCloneURL,
		Branch:       resolution.Branch, CommitSHA: resolution.CommitSHA,
		Credential: workerdependency.RepositoryCredential{
			Type: resolution.CredentialType, CredentialID: credentialID,
			OwnerUserID: resolution.CredentialOwnerUserID,
		},
		PreparationScript:         resolution.PreparationScript,
		PreparationScriptDigest:   resolution.PreparationScriptDigest,
		PreparationTimeoutSeconds: resolution.PreparationTimeoutSeconds,
	}
}

func materializeSkills(
	resolutions []SkillResolution,
) []workerdependency.Skill {
	result := make([]workerdependency.Skill, len(resolutions))
	for index, resolution := range resolutions {
		result[index] = workerdependency.Skill{
			Pin: pin(resolution.ResourceResolution), Slug: resolution.Slug,
			Version: resolution.Version, ContentDigest: resolution.ContentDigest,
			StorageKey: resolution.StorageKey, PackageSize: resolution.PackageSize,
		}
	}
	return result
}

func materializeKnowledgeBases(
	resolutions []KnowledgeBaseResolution,
) []workerdependency.KnowledgeBase {
	result := make([]workerdependency.KnowledgeBase, len(resolutions))
	for index, resolution := range resolutions {
		result[index] = workerdependency.KnowledgeBase{
			Pin: pin(resolution.ResourceResolution), Slug: resolution.Slug,
			HTTPCloneURL: resolution.HTTPCloneURL, Branch: resolution.Branch,
			CommitSHA: resolution.CommitSHA, Mode: resolution.Mode,
		}
	}
	return result
}

func materializeRuntimeBundles(
	resolutions []RuntimeBundleResolution,
) []workerdependency.RuntimeBundle {
	result := make([]workerdependency.RuntimeBundle, len(resolutions))
	for index, resolution := range resolutions {
		values := make([]workerdependency.RuntimeValue, len(resolution.Values))
		for valueIndex, value := range resolution.Values {
			values[valueIndex] = workerdependency.RuntimeValue{
				Name: value.Name, Value: value.Value,
			}
		}
		var document *workerdependency.ConfigDocument
		if resolution.ConfigDocument != nil {
			document = &workerdependency.ConfigDocument{
				ID:         resolution.ConfigDocument.ID,
				Format:     resolution.ConfigDocument.Format,
				TargetPath: resolution.ConfigDocument.TargetPath,
			}
		}
		result[index] = workerdependency.RuntimeBundle{
			Pin: pin(resolution.ResourceResolution), Kind: resolution.Kind,
			ContentDigest: resolution.ContentDigest, Values: values,
			ConfigDocument: document,
		}
	}
	return result
}

func materializeSecretReferences(
	resolutions []SecretReferenceResolution,
) []workerdependency.SecretReference {
	result := make([]workerdependency.SecretReference, len(resolutions))
	for index, resolution := range resolutions {
		result[index] = workerdependency.SecretReference{
			Pin: pin(resolution.ResourceResolution), Field: resolution.Field,
			BundleKey: resolution.BundleKey, OwnerScope: resolution.OwnerScope,
			OwnerID: resolution.OwnerID,
		}
	}
	return result
}

func materializePlacement(
	resolution PlacementResolution,
) workerdependency.Placement {
	var profile *workerdependency.ResourcePin
	if resolution.ResourceProfile != nil {
		value := pin(*resolution.ResourceProfile)
		profile = &value
	}
	return workerdependency.Placement{
		CatalogRevision: resolution.CatalogRevision,
		RuntimeImage: workerdependency.RuntimeImage{
			ID: resolution.RuntimeImageID, Reference: resolution.ImageReference,
			Digest: resolution.ImageDigest,
		},
		ComputeTarget:   pin(resolution.ComputeTarget),
		ResourceProfile: profile, Spec: resolution.Spec,
	}
}
