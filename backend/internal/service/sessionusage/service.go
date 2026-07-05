package sessionusage

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Row struct {
	PodKey              string    `gorm:"primaryKey;size:100"`
	Model               string    `gorm:"primaryKey;size:100"`
	InputTokens         int64     `gorm:"not null;default:0"`
	OutputTokens        int64     `gorm:"not null;default:0"`
	CacheReadTokens     int64     `gorm:"not null;default:0"`
	CacheCreationTokens int64     `gorm:"not null;default:0"`
	UpdatedAt           time.Time `gorm:"not null"`
}

func (Row) TableName() string { return "pod_session_usage" }

type ModelUsage struct {
	Model               string   `json:"model"`
	InputTokens         int64    `json:"input_tokens"`
	OutputTokens        int64    `json:"output_tokens"`
	CacheReadTokens     int64    `json:"cache_read_tokens"`
	CacheCreationTokens int64    `json:"cache_creation_tokens"`
	TotalCostUSD        *float64 `json:"total_cost_usd,omitempty"`
}

type Aggregate struct {
	TotalCostUSD *float64              `json:"total_cost_usd,omitempty"`
	UsageByModel map[string]ModelUsage `json:"usage_by_model,omitempty"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Upsert(ctx context.Context, podKey, model string, in, out, cacheRead, cacheCreate int64) error {
	if podKey == "" || model == "" {
		return nil
	}
	row := Row{
		PodKey: podKey, Model: model,
		InputTokens: in, OutputTokens: out,
		CacheReadTokens: cacheRead, CacheCreationTokens: cacheCreate,
		UpdatedAt: time.Now(),
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "pod_key"}, {Name: "model"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"input_tokens", "output_tokens", "cache_read_tokens", "cache_creation_tokens", "updated_at",
		}),
	}).Create(&row).Error
}

func (s *Service) Aggregate(ctx context.Context, podKey string) (Aggregate, error) {
	var rows []Row
	if err := s.db.WithContext(ctx).Where("pod_key = ?", podKey).Find(&rows).Error; err != nil {
		return Aggregate{}, err
	}
	return aggregateRows(s.db.WithContext(ctx), rows)
}

func (s *Service) AggregateOrg(ctx context.Context, orgID int64) (Aggregate, error) {
	var podKeys []string
	if err := s.db.WithContext(ctx).Table("pods").
		Where("organization_id = ?", orgID).
		Pluck("pod_key", &podKeys).Error; err != nil {
		return Aggregate{}, err
	}
	if len(podKeys) == 0 {
		return Aggregate{}, nil
	}
	var rows []Row
	if err := s.db.WithContext(ctx).Where("pod_key IN ?", podKeys).Find(&rows).Error; err != nil {
		return Aggregate{}, err
	}
	return aggregateRows(s.db.WithContext(ctx), rows)
}

func aggregateRows(db *gorm.DB, rows []Row) (Aggregate, error) {
	if len(rows) == 0 {
		return Aggregate{}, nil
	}
	type priceRow struct {
		Model            string  `gorm:"column:model"`
		InputPerMillion  float64 `gorm:"column:input_per_million"`
		OutputPerMillion float64 `gorm:"column:output_per_million"`
	}
	modelSet := make(map[string]struct{})
	for _, r := range rows {
		modelSet[r.Model] = struct{}{}
	}
	models := make([]string, 0, len(modelSet))
	for m := range modelSet {
		models = append(models, m)
	}
	var prices []priceRow
	_ = db.Table("model_prices").Where("model IN ?", models).Find(&prices)
	priceMap := make(map[string]priceRow, len(prices))
	for _, p := range prices {
		priceMap[p.Model] = p
	}
	byModel := make(map[string]ModelUsage, len(modelSet))
	var total float64
	hasCost := false
	for _, r := range rows {
		prev := byModel[r.Model]
		mu := ModelUsage{
			Model: r.Model,
			InputTokens: prev.InputTokens + r.InputTokens,
			OutputTokens: prev.OutputTokens + r.OutputTokens,
			CacheReadTokens: prev.CacheReadTokens + r.CacheReadTokens,
			CacheCreationTokens: prev.CacheCreationTokens + r.CacheCreationTokens,
		}
		if p, ok := priceMap[r.Model]; ok {
			cost := (float64(r.InputTokens)/1e6)*p.InputPerMillion +
				(float64(r.OutputTokens)/1e6)*p.OutputPerMillion
			total += cost
			hasCost = true
			if prev.TotalCostUSD != nil {
				sum := *prev.TotalCostUSD + cost
				mu.TotalCostUSD = &sum
			} else {
				mu.TotalCostUSD = &cost
			}
		} else if prev.TotalCostUSD != nil {
			mu.TotalCostUSD = prev.TotalCostUSD
		}
		byModel[r.Model] = mu
	}
	agg := Aggregate{UsageByModel: byModel}
	if hasCost {
		agg.TotalCostUSD = &total
	}
	return agg, nil
}
