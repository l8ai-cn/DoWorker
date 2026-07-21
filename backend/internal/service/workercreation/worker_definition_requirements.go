package workercreation

import "github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"

type WorkerCredentialRequirement struct {
	ID         string
	SourceKind string
	SourceRef  string
	TargetKind string
	TargetName string
}

type WorkerConfigDocumentRequirement struct {
	DocumentID string
	Format     string
	TargetPath string
	Required   bool
}

func workerDefinitionRequirements(
	definition workerdefinition.Definition,
) ([]WorkerCredentialRequirement, []WorkerConfigDocumentRequirement) {
	credentials := make(
		[]WorkerCredentialRequirement,
		len(definition.CredentialBindings),
	)
	for index, binding := range definition.CredentialBindings {
		credentials[index] = WorkerCredentialRequirement{
			ID:         binding.ID,
			SourceKind: binding.Source.Kind,
			SourceRef:  binding.Source.Ref,
			TargetKind: binding.Target.Kind,
			TargetName: binding.Target.Name,
		}
	}
	documents := make(
		[]WorkerConfigDocumentRequirement,
		len(definition.ConfigDocuments),
	)
	for index, document := range definition.ConfigDocuments {
		documents[index] = WorkerConfigDocumentRequirement{
			DocumentID: document.ID,
			Format:     document.Format,
			TargetPath: document.TargetPath,
			Required:   document.Required,
		}
	}
	return credentials, documents
}
