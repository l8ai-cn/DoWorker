package podconnect

import (
	"fmt"
	"strings"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
)

func workerConfigDocumentBindingsFromProto(
	items []*podv1.WorkerConfigDocumentBinding,
) ([]specdomain.ConfigDocumentBinding, error) {
	bindings := make([]specdomain.ConfigDocumentBinding, 0, len(items))
	documents := make(map[string]struct{}, len(items))
	bundles := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if item == nil {
			return nil, invalidWorkerDraft(
				"config_document_bindings",
				"contains an empty item",
			)
		}
		documentID := item.GetDocumentId()
		if documentID == "" || strings.TrimSpace(documentID) != documentID {
			return nil, invalidWorkerDraft(
				"config_document_bindings",
				"document_id must be normalized",
			)
		}
		if _, exists := documents[documentID]; exists {
			return nil, invalidWorkerDraft(
				"config_document_bindings",
				fmt.Sprintf("contains duplicate document %q", documentID),
			)
		}
		bundleID := item.GetConfigBundleId()
		if bundleID <= 0 {
			return nil, invalidWorkerDraft(
				"config_document_bindings",
				"config_bundle_id must be positive",
			)
		}
		if _, exists := bundles[bundleID]; exists {
			return nil, invalidWorkerDraft(
				"config_document_bindings",
				fmt.Sprintf("contains duplicate bundle id %d", bundleID),
			)
		}
		documents[documentID] = struct{}{}
		bundles[bundleID] = struct{}{}
		bindings = append(bindings, specdomain.ConfigDocumentBinding{
			DocumentID: documentID, ConfigBundleID: bundleID,
		})
	}
	return bindings, nil
}
