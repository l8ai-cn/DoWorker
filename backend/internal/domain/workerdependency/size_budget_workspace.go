package workerdependency

func validateWorkspaceSizeBudget(
	document Document,
	consume func(int) bool,
	consumeText func(...string) bool,
	consumeReference func(ResourcePin) bool,
) error {
	if repository := document.Repository; repository != nil {
		if !consumeReference(repository.Pin) || !consumeText(
			repository.HTTPCloneURL,
			repository.SSHCloneURL,
			repository.Branch,
			repository.CommitSHA,
			repository.Credential.Type,
			repository.PreparationScript,
			repository.PreparationScriptDigest,
		) {
			return ErrDocumentTooLarge
		}
	}
	for _, skill := range document.Skills {
		if !consumeReference(skill.Pin) || !consumeText(
			skill.Slug.String(),
			skill.ContentDigest,
			skill.StorageKey,
		) {
			return ErrDocumentTooLarge
		}
	}
	for _, knowledge := range document.KnowledgeBases {
		if !consumeReference(knowledge.Pin) || !consumeText(
			knowledge.Slug.String(),
			knowledge.HTTPCloneURL,
			knowledge.Branch,
			knowledge.CommitSHA,
		) {
			return ErrDocumentTooLarge
		}
	}
	for _, bundle := range document.RuntimeBundles {
		if !consumeReference(bundle.Pin) ||
			!consumeText(bundle.Kind, bundle.ContentDigest) ||
			!consume(len(bundle.Values)) {
			return ErrDocumentTooLarge
		}
		for _, value := range bundle.Values {
			if !consumeText(value.Name, value.Value) {
				return ErrDocumentTooLarge
			}
		}
		if bundle.ConfigDocument != nil && !consumeText(
			bundle.ConfigDocument.ID,
			bundle.ConfigDocument.Format,
			bundle.ConfigDocument.TargetPath,
		) {
			return ErrDocumentTooLarge
		}
	}
	for _, secret := range document.SecretReferences {
		if !consumeReference(secret.Pin) || !consumeText(
			secret.Field,
			secret.BundleKey,
			secret.OwnerScope,
		) {
			return ErrDocumentTooLarge
		}
	}
	placement := document.Placement
	if !consumeText(
		placement.CatalogRevision,
		placement.RuntimeImage.Reference,
		placement.RuntimeImage.Digest,
	) || !consumeReference(placement.ComputeTarget) {
		return ErrDocumentTooLarge
	}
	if placement.ResourceProfile != nil &&
		!consumeReference(*placement.ResourceProfile) {
		return ErrDocumentTooLarge
	}
	return nil
}
