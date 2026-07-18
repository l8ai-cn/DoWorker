package workerdependency

func validateDocumentSizeBudget(document Document) error {
	budget := MaxDocumentBytes
	consume := func(size int) bool {
		if size < 0 || size > budget {
			return false
		}
		budget -= size
		return true
	}
	consumeText := func(values ...string) bool {
		for _, value := range values {
			if !consume(len(value)) {
				return false
			}
		}
		return true
	}
	consumeReference := func(pin ResourcePin) bool {
		reference := pin.Reference
		return consumeText(
			reference.APIVersion,
			reference.Kind,
			reference.Namespace.String(),
			reference.Name.String(),
			reference.UID,
			reference.Digest,
		)
	}
	worker := document.Worker
	if !consumeText(
		document.Namespace.String(),
		worker.WorkerType.String(),
		worker.AdapterID.String(),
		worker.SpecDigest,
		worker.DefinitionHash,
		worker.AgentfileSource,
		worker.AgentfileSourceDigest,
	) || !consumeStringSlice(consumeText, worker.ModelManagedFields) ||
		!consumeStringSlice(consumeText, worker.CredentialBundleFields) {
		return ErrDocumentTooLarge
	}
	models := make([]Model, 0, len(document.Models.Tools)+1)
	if document.Models.Primary != nil {
		models = append(models, *document.Models.Primary)
	}
	for _, tool := range document.Models.Tools {
		if !consumeText(
			tool.Role.String(),
			string(tool.Modality),
			string(tool.Capability),
			tool.Environment.APIKeyTarget,
			tool.Environment.BaseURLTarget,
			tool.Environment.ModelIDTarget,
		) || !consumeReference(ResourcePin{Reference: tool.Binding}) {
			return ErrDocumentTooLarge
		}
		models = append(models, tool.Model)
	}
	for _, model := range models {
		if !consumeReference(model.Pin) || !consumeText(
			model.ProviderKey.String(),
			model.ProtocolAdapter.String(),
			model.ModelID,
			model.BaseURL,
		) || !consume(len(model.Modalities)+len(model.Capabilities)) {
			return ErrDocumentTooLarge
		}
	}
	return validateWorkspaceSizeBudget(document, consume, consumeText, consumeReference)
}

func consumeStringSlice(
	consume func(...string) bool,
	values []string,
) bool {
	for _, value := range values {
		if !consume(value) {
			return false
		}
	}
	return true
}
