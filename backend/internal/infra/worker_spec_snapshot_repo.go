package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workerspecservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"gorm.io/gorm"
)

var _ workerspecservice.SnapshotRepository = (*workerSpecSnapshotRepo)(nil)

type workerSpecSnapshotRepo struct {
	db *gorm.DB
}

type workerSpecSnapshotRecord struct {
	ID             int64           `gorm:"primaryKey"`
	OrganizationID int64           `gorm:"not null"`
	Version        domain.Version  `gorm:"not null"`
	SpecJSON       json.RawMessage `gorm:"column:spec_json;type:jsonb;not null"`
	SummaryJSON    json.RawMessage `gorm:"column:summary_json;type:jsonb;not null"`
	CreatedAt      time.Time       `gorm:"not null"`
}

func (workerSpecSnapshotRecord) TableName() string {
	return "worker_spec_snapshots"
}

func NewWorkerSpecSnapshotRepository(db *gorm.DB) workerspecservice.SnapshotRepository {
	return &workerSpecSnapshotRepo{db: db}
}

func (r *workerSpecSnapshotRepo) Create(
	ctx context.Context,
	resolved workerspecservice.ResolvedSnapshot,
) (domain.Snapshot, error) {
	return createWorkerSpecSnapshot(r.db.WithContext(ctx), resolved)
}

func (r *workerSpecSnapshotRepo) GetByID(
	ctx context.Context,
	organizationID, id int64,
) (domain.Snapshot, error) {
	var record workerSpecSnapshotRecord
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", organizationID, id).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Snapshot{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Snapshot{}, err
	}
	return workerSpecSnapshotFromRecord(record)
}

func (r *workerSpecSnapshotRepo) ListByOrganization(
	ctx context.Context,
	organizationID int64,
) ([]domain.Snapshot, error) {
	var records []workerSpecSnapshotRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Order("created_at DESC, id DESC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	snapshots := make([]domain.Snapshot, 0, len(records))
	for _, record := range records {
		snapshot, err := workerSpecSnapshotFromRecord(record)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

func workerSpecSnapshotFromRecord(record workerSpecSnapshotRecord) (domain.Snapshot, error) {
	if record.Version != domain.VersionV1 {
		return domain.Snapshot{}, fmt.Errorf("%w: %d", domain.ErrUnsupportedVersion, record.Version)
	}
	spec, err := domain.DecodePersistedSpec(record.SpecJSON)
	if err != nil {
		return domain.Snapshot{}, err
	}
	summary, err := domain.DecodePersistedSummary(record.SummaryJSON)
	if err != nil {
		return domain.Snapshot{}, err
	}
	expected, err := domain.SummarizePersisted(spec)
	if err != nil {
		return domain.Snapshot{}, err
	}
	if !reflect.DeepEqual(summary, expected) {
		return domain.Snapshot{}, domain.ErrSummaryMismatch
	}
	return domain.Snapshot{
		ID:             record.ID,
		OrganizationID: record.OrganizationID,
		Spec:           spec,
		Summary:        summary,
		CreatedAt:      record.CreatedAt,
	}, nil
}

func (r *workerSpecSnapshotRepo) Delete(
	ctx context.Context,
	organizationID, id int64,
) error {
	result := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", organizationID, id).
		Delete(&workerSpecSnapshotRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func decodeResolvedSnapshot(
	resolved workerspecservice.ResolvedSnapshot,
) (domain.Spec, domain.Summary, []byte, []byte, error) {
	specJSON := resolved.SpecJSON()
	summaryJSON := resolved.SummaryJSON()
	if resolved.OrganizationID() <= 0 ||
		resolved.Version() != domain.VersionV1 ||
		len(specJSON) == 0 ||
		len(summaryJSON) == 0 {
		return invalidResolvedSnapshot("zero or incomplete aggregate")
	}
	spec, err := domain.DecodeSpec(specJSON)
	if err != nil || spec.Version != resolved.Version() {
		return invalidResolvedSnapshot("invalid spec document")
	}
	summary, err := domain.DecodeSummary(summaryJSON)
	if err != nil {
		return invalidResolvedSnapshot("invalid summary document")
	}
	expected, err := domain.Summarize(spec)
	if err != nil || !reflect.DeepEqual(summary, expected) {
		return invalidResolvedSnapshot("summary does not match spec")
	}
	canonicalSpec, err := domain.EncodeSpec(spec)
	if err != nil || !bytes.Equal(specJSON, canonicalSpec) {
		return invalidResolvedSnapshot("spec document is not canonical")
	}
	canonicalSummary, err := domain.EncodeSummary(summary)
	if err != nil || !bytes.Equal(summaryJSON, canonicalSummary) {
		return invalidResolvedSnapshot("summary document is not canonical")
	}
	return spec, summary, specJSON, summaryJSON, nil
}

func invalidResolvedSnapshot(
	reason string,
) (domain.Spec, domain.Summary, []byte, []byte, error) {
	return domain.Spec{}, domain.Summary{}, nil, nil, fmt.Errorf(
		"%w: %s",
		workerspecservice.ErrInvalidResolvedSnapshot,
		reason,
	)
}
