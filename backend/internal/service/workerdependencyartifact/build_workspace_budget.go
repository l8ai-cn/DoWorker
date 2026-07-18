package workerdependencyartifact

func consumeWorkspaceBudget(
	resolved ResolvedDependencies,
	text func(...string) bool,
	count func(int) bool,
	reference func(ResourceResolution) bool,
) bool {
	if repository := resolved.Repository; repository != nil {
		if !reference(repository.ResourceResolution) || !text(
			repository.HTTPCloneURL,
			repository.SSHCloneURL,
			repository.Branch,
			repository.CommitSHA,
			repository.CredentialType,
			repository.PreparationScript,
			repository.PreparationScriptDigest,
		) {
			return false
		}
	}
	for _, skill := range resolved.Skills {
		if !reference(skill.ResourceResolution) || !text(
			skill.Slug.String(),
			skill.ContentDigest,
			skill.StorageKey,
		) {
			return false
		}
	}
	for _, knowledge := range resolved.KnowledgeBases {
		if !reference(knowledge.ResourceResolution) || !text(
			knowledge.Slug.String(),
			knowledge.HTTPCloneURL,
			knowledge.Branch,
			knowledge.CommitSHA,
			string(knowledge.Mode),
		) {
			return false
		}
	}
	for _, bundle := range resolved.RuntimeBundles {
		if !reference(bundle.ResourceResolution) ||
			!text(bundle.Kind, bundle.ContentDigest) ||
			!count(len(bundle.Values)) {
			return false
		}
		for _, value := range bundle.Values {
			if !text(value.Name, value.Value) {
				return false
			}
		}
		if bundle.ConfigDocument != nil && !text(
			bundle.ConfigDocument.ID,
			bundle.ConfigDocument.Format,
			bundle.ConfigDocument.TargetPath,
		) {
			return false
		}
	}
	for _, secret := range resolved.SecretReferences {
		if !reference(secret.ResourceResolution) || !text(
			secret.Field,
			secret.BundleKey,
			secret.OwnerScope,
		) {
			return false
		}
	}
	placement := resolved.Placement
	if !text(
		placement.CatalogRevision,
		placement.ImageReference,
		placement.ImageDigest,
		string(placement.Spec.Policy),
		string(placement.Spec.ComputeTarget.Kind),
		string(placement.Spec.DeploymentMode),
	) || !reference(placement.ComputeTarget) {
		return false
	}
	if placement.ResourceProfile != nil &&
		!reference(*placement.ResourceProfile) {
		return false
	}
	return true
}
