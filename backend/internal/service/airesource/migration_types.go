package airesource

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type migrationStringList []string

func (list *migrationStringList) Scan(value any) error {
	if value == nil {
		*list = nil
		return nil
	}
	var raw []byte
	switch typed := value.(type) {
	case []byte:
		raw = typed
	case string:
		raw = []byte(typed)
	default:
		return fmt.Errorf("cannot scan migration string list from %T", value)
	}
	return json.Unmarshal(raw, list)
}

func (list migrationStringList) Value() (driver.Value, error) {
	if list == nil {
		return "[]", nil
	}
	raw, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return string(raw), nil
}

type MigrationReport struct {
	AIModelsMigrated    int
	EnvBundlesMigrated  int
	VirtualKeysRemapped int
}

type MigrationCheckReport struct {
	UnmigratedAIModels   int
	UnmigratedEnvBundles int
	UnmappedVirtualKeys  int
	BrokenMappings       int
	DecryptFailures      int
	FieldMismatches      int
	ScopeMismatches      int
}

func (r MigrationCheckReport) Clean() bool {
	return r.UnmigratedAIModels == 0 &&
		r.UnmigratedEnvBundles == 0 &&
		r.UnmappedVirtualKeys == 0 &&
		r.BrokenMappings == 0 &&
		r.DecryptFailures == 0 &&
		r.FieldMismatches == 0 &&
		r.ScopeMismatches == 0
}

type legacyAIModelRow struct {
	ID                   int64
	OrganizationID       *int64
	UserID               *int64
	Name                 string
	ProviderType         string
	Model                string
	BaseURL              string
	EncryptedCredentials string
	IsDefault            bool
	IsEnabled            bool
}

type legacyEnvBundleRow struct {
	ID          int64
	OwnerScope  string
	OwnerID     int64
	AgentSlug   *string
	Name        string
	Kind        string
	KindPrimary bool
	Data        map[string]string `gorm:"serializer:json"`
	IsActive    bool
}

type migrationConnectionRow struct {
	ID                   int64
	OwnerScope           string
	OwnerID              int64
	Identifier           string
	ProviderKey          string
	Name                 string
	BaseURL              string
	CredentialsEncrypted string
	ConfiguredFields     migrationStringList
	Status               string
	IsEnabled            bool
	CreatedBy            int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (migrationConnectionRow) TableName() string { return "provider_connections" }

type migrationResourceRow struct {
	ID                   int64
	ProviderConnectionID int64
	Identifier           string
	ModelID              string
	DisplayName          string
	Modalities           migrationStringList
	Capabilities         migrationStringList
	Status               string
	IsEnabled            bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (migrationResourceRow) TableName() string { return "model_resources" }
