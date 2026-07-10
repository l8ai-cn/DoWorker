package workerspec

import (
	"bytes"
	"fmt"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

type ResolvedSnapshot struct {
	organizationID int64
	version        domain.Version
	specJSON       []byte
	summaryJSON    []byte
}

func (snapshot ResolvedSnapshot) OrganizationID() int64 {
	return snapshot.organizationID
}

func (snapshot ResolvedSnapshot) Version() domain.Version {
	return snapshot.version
}

func (snapshot ResolvedSnapshot) SpecJSON() []byte {
	return bytes.Clone(snapshot.specJSON)
}

func (snapshot ResolvedSnapshot) SummaryJSON() []byte {
	return bytes.Clone(snapshot.summaryJSON)
}

func resolveSnapshot(
	organizationID int64,
	spec domain.Spec,
) (ResolvedSnapshot, error) {
	summary, err := domain.Summarize(spec)
	if err != nil {
		return ResolvedSnapshot{}, fmt.Errorf("summarize workerspec: %w", err)
	}
	specJSON, err := domain.EncodeSpec(spec)
	if err != nil {
		return ResolvedSnapshot{}, fmt.Errorf("encode workerspec: %w", err)
	}
	summaryJSON, err := domain.EncodeSummary(summary)
	if err != nil {
		return ResolvedSnapshot{}, fmt.Errorf("encode workerspec summary: %w", err)
	}
	return ResolvedSnapshot{
		organizationID: organizationID,
		version:        spec.Version,
		specJSON:       bytes.Clone(specJSON),
		summaryJSON:    bytes.Clone(summaryJSON),
	}, nil
}
