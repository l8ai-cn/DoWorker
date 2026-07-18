package podconnect

import (
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
)

func workerCredentialRequirementsToProto(
	requirements []workercreation.WorkerCredentialRequirement,
) []*podv1.WorkerCredentialRequirement {
	items := make([]*podv1.WorkerCredentialRequirement, len(requirements))
	for index, requirement := range requirements {
		items[index] = &podv1.WorkerCredentialRequirement{
			Id: requirement.ID, SourceKind: requirement.SourceKind,
			SourceRef: requirement.SourceRef, TargetKind: requirement.TargetKind,
			TargetName: requirement.TargetName,
		}
	}
	return items
}

func workerConfigDocumentRequirementsToProto(
	requirements []workercreation.WorkerConfigDocumentRequirement,
) []*podv1.WorkerConfigDocumentRequirement {
	items := make([]*podv1.WorkerConfigDocumentRequirement, len(requirements))
	for index, requirement := range requirements {
		items[index] = &podv1.WorkerConfigDocumentRequirement{
			DocumentId: requirement.DocumentID,
			Format:     requirement.Format,
			TargetPath: requirement.TargetPath,
			Required:   requirement.Required,
		}
	}
	return items
}
