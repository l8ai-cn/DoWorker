package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"gorm.io/gorm"
)

var _ workerspec.Repository = (*workerSpecSnapshotRepo)(nil)

type workerSpecSnapshotRepo struct {
	db *gorm.DB
}

type workerSpecSnapshotRecord struct {
	ID             int64              `gorm:"primaryKey"`
	OrganizationID int64              `gorm:"not null"`
	Version        workerspec.Version `gorm:"not null"`
	SpecJSON       json.RawMessage    `gorm:"column:spec_json;type:jsonb;not null"`
	SummaryJSON    json.RawMessage    `gorm:"column:summary_json;type:jsonb;not null"`
	CreatedAt      time.Time          `gorm:"not null"`
}

func (workerSpecSnapshotRecord) TableName() string {
	return "worker_spec_snapshots"
}

func NewWorkerSpecSnapshotRepository(db *gorm.DB) workerspec.Repository {
	return &workerSpecSnapshotRepo{db: db}
}

func (r *workerSpecSnapshotRepo) Create(
	ctx context.Context,
	snapshot *workerspec.Snapshot,
) error {
	if snapshot == nil {
		return errors.New("workerspec snapshot is required")
	}
	canonical, err := workerspec.NewSnapshot(snapshot.OrganizationID, snapshot.Spec)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(snapshot.Summary, canonical.Summary) {
		return workerspec.ErrSummaryMismatch
	}
	specJSON, err := workerspec.EncodeSpec(canonical.Spec)
	if err != nil {
		return err
	}
	summaryJSON, err := workerspec.EncodeSummary(canonical.Summary)
	if err != nil {
		return err
	}
	record := workerSpecSnapshotRecord{
		OrganizationID: canonical.OrganizationID,
		Version:        canonical.Spec.Version,
		SpecJSON:       specJSON,
		SummaryJSON:    summaryJSON,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return err
	}
	snapshot.ID = record.ID
	snapshot.CreatedAt = record.CreatedAt
	snapshot.Spec = canonical.Spec
	snapshot.Summary = canonical.Summary
	return nil
}

func (r *workerSpecSnapshotRepo) GetByID(
	ctx context.Context,
	organizationID, id int64,
) (*workerspec.Snapshot, error) {
	var record workerSpecSnapshotRecord
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND id = ?", organizationID, id).
		First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, workerspec.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if record.Version != workerspec.VersionV1 {
		return nil, fmt.Errorf("%w: %d", workerspec.ErrUnsupportedVersion, record.Version)
	}
	spec, err := workerspec.DecodeSpec(record.SpecJSON)
	if err != nil {
		return nil, err
	}
	summary, err := workerspec.DecodeSummary(record.SummaryJSON)
	if err != nil {
		return nil, err
	}
	expected, err := workerspec.Summarize(spec)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(summary, expected) {
		return nil, workerspec.ErrSummaryMismatch
	}
	return &workerspec.Snapshot{
		ID:             record.ID,
		OrganizationID: record.OrganizationID,
		Spec:           spec,
		Summary:        summary,
		CreatedAt:      record.CreatedAt,
	}, nil
}
