package workerdependencyartifact

import (
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type ResolvedDependencies struct {
	PrimaryModel     *ModelResolution
	ToolModels       []ToolModelResolution
	Repository       *RepositoryResolution
	Skills           []SkillResolution
	KnowledgeBases   []KnowledgeBaseResolution
	RuntimeBundles   []RuntimeBundleResolution
	SecretReferences []SecretReferenceResolution
	Placement        PlacementResolution
}

type ResourceResolution struct {
	reference control.ResolvedReference
	domainID  int64
}

func BindResourceProjection(
	scope control.Scope,
	reference control.ResolvedReference,
	domainID int64,
) (ResourceResolution, error) {
	if err := reference.Validate(scope); err != nil {
		return ResourceResolution{}, fmt.Errorf("validate resource projection: %w", err)
	}
	if domainID <= 0 {
		return ResourceResolution{}, fmt.Errorf(
			"resource projection domain id must be positive",
		)
	}
	return ResourceResolution{reference: reference, domainID: domainID}, nil
}

func (resolution ResourceResolution) ResolvedReference() control.ResolvedReference {
	return resolution.reference
}

func (resolution ResourceResolution) DomainID() int64 {
	return resolution.domainID
}

type ModelResolution struct {
	ResourceResolution
	ResourceRevision   int64
	ConnectionID       int64
	ConnectionRevision int64
	ProviderKey        slugkit.Slug
	ProtocolAdapter    slugkit.Slug
	ModelID            string
	BaseURL            string
	Modalities         []airesource.Modality
	Capabilities       []airesource.Capability
}

type ToolModelResolution struct {
	Binding     control.ResolvedReference
	Role        slugkit.Slug
	Model       ModelResolution
	Modality    airesource.Modality
	Capability  airesource.Capability
	Environment ToolModelEnvironmentResolution
}

type ToolModelEnvironmentResolution struct {
	APIKeyTarget  string
	BaseURLTarget string
	ModelIDTarget string
}

type RepositoryResolution struct {
	ResourceResolution
	HTTPCloneURL              string
	SSHCloneURL               string
	Branch                    string
	CommitSHA                 string
	CredentialType            string
	CredentialID              *int64
	CredentialOwnerUserID     int64
	PreparationScript         string
	PreparationScriptDigest   string
	PreparationTimeoutSeconds uint32
}

type SkillResolution struct {
	ResourceResolution
	Slug          slugkit.Slug
	Version       int
	ContentDigest string
	StorageKey    string
	PackageSize   int64
}

type KnowledgeBaseResolution struct {
	ResourceResolution
	Slug         slugkit.Slug
	HTTPCloneURL string
	Branch       string
	CommitSHA    string
	Mode         workerspec.KnowledgeMountMode
}

type RuntimeBundleResolution struct {
	ResourceResolution
	Kind           string
	ContentDigest  string
	Values         []RuntimeValueResolution
	ConfigDocument *ConfigDocumentResolution
}

type RuntimeValueResolution struct {
	Name  string
	Value string
}

type ConfigDocumentResolution struct {
	ID         string
	Format     string
	TargetPath string
}

type SecretReferenceResolution struct {
	ResourceResolution
	Field      string
	BundleKey  string
	OwnerScope string
	OwnerID    int64
}

type PlacementResolution struct {
	CatalogRevision string
	RuntimeImageID  int64
	ImageReference  string
	ImageDigest     string
	ComputeTarget   ResourceResolution
	ResourceProfile *ResourceResolution
	Spec            workerspec.Placement
}
