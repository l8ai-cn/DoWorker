package workerdependency

import (
	"github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type Version uint16

const VersionV1 Version = 1

const RepositoryCredentialTypeNone = "none"

type Document struct {
	Version          Version           `json:"version"`
	OrganizationID   int64             `json:"organization_id"`
	Namespace        slugkit.Slug      `json:"namespace"`
	Worker           Worker            `json:"worker"`
	Models           Models            `json:"models"`
	Repository       *Repository       `json:"repository"`
	Skills           []Skill           `json:"skills"`
	KnowledgeBases   []KnowledgeBase   `json:"knowledge_bases"`
	RuntimeBundles   []RuntimeBundle   `json:"runtime_bundles"`
	SecretReferences []SecretReference `json:"secret_refs"`
	Placement        Placement         `json:"placement"`
}

type Worker struct {
	WorkerType             slugkit.Slug       `json:"worker_type"`
	AdapterID              slugkit.Slug       `json:"adapter_id"`
	SpecVersion            workerspec.Version `json:"spec_version"`
	SpecDigest             string             `json:"spec_digest"`
	DefinitionHash         string             `json:"definition_hash"`
	ModelManagedFields     []string           `json:"model_managed_fields"`
	CredentialBundleFields []string           `json:"credential_bundle_fields"`
	AgentfileSource        string             `json:"agentfile_source"`
	AgentfileSourceDigest  string             `json:"agentfile_source_digest"`
}

type ResourcePin struct {
	Reference orchestrationresource.Reference `json:"reference"`
	DomainID  int64                           `json:"domain_id"`
}

type Models struct {
	Primary *Model      `json:"primary"`
	Tools   []ToolModel `json:"tools"`
}

type Model struct {
	Pin                ResourcePin             `json:"pin"`
	ResourceRevision   int64                   `json:"resource_revision"`
	ConnectionID       int64                   `json:"connection_id"`
	ConnectionRevision int64                   `json:"connection_revision"`
	ProviderKey        slugkit.Slug            `json:"provider_key"`
	ProtocolAdapter    slugkit.Slug            `json:"protocol_adapter"`
	ModelID            string                  `json:"model_id"`
	BaseURL            string                  `json:"base_url"`
	Modalities         []airesource.Modality   `json:"modalities"`
	Capabilities       []airesource.Capability `json:"capabilities"`
}

type ToolModel struct {
	Binding     orchestrationresource.Reference `json:"binding"`
	Role        slugkit.Slug                    `json:"role"`
	Model       Model                           `json:"model"`
	Modality    airesource.Modality             `json:"modality"`
	Capability  airesource.Capability           `json:"capability"`
	Environment ToolModelEnvironment            `json:"environment"`
}

type ToolModelEnvironment struct {
	APIKeyTarget  string `json:"api_key_target"`
	BaseURLTarget string `json:"base_url_target"`
	ModelIDTarget string `json:"model_id_target"`
}

type Repository struct {
	Pin                       ResourcePin          `json:"pin"`
	HTTPCloneURL              string               `json:"http_clone_url"`
	SSHCloneURL               string               `json:"ssh_clone_url"`
	Branch                    string               `json:"branch"`
	CommitSHA                 string               `json:"commit_sha"`
	Credential                RepositoryCredential `json:"credential_ref"`
	PreparationScript         string               `json:"preparation_script"`
	PreparationScriptDigest   string               `json:"preparation_script_digest"`
	PreparationTimeoutSeconds uint32               `json:"preparation_timeout_seconds"`
}

type RepositoryCredential struct {
	Type         string `json:"type"`
	CredentialID *int64 `json:"credential_id"`
	OwnerUserID  int64  `json:"owner_user_id"`
}

type Skill struct {
	Pin           ResourcePin  `json:"pin"`
	Slug          slugkit.Slug `json:"slug"`
	Version       int          `json:"version"`
	ContentDigest string       `json:"content_digest"`
	StorageKey    string       `json:"storage_key"`
	PackageSize   int64        `json:"package_size"`
}

type KnowledgeBase struct {
	Pin          ResourcePin                   `json:"pin"`
	Slug         slugkit.Slug                  `json:"slug"`
	HTTPCloneURL string                        `json:"http_clone_url"`
	Branch       string                        `json:"branch"`
	CommitSHA    string                        `json:"commit_sha"`
	Mode         workerspec.KnowledgeMountMode `json:"mode"`
}

type RuntimeBundle struct {
	Pin            ResourcePin     `json:"pin"`
	Kind           string          `json:"kind"`
	ContentDigest  string          `json:"content_digest"`
	Values         []RuntimeValue  `json:"values"`
	ConfigDocument *ConfigDocument `json:"config_document"`
}

type RuntimeValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ConfigDocument struct {
	ID         string `json:"id"`
	Format     string `json:"format"`
	TargetPath string `json:"target_path"`
}

type SecretReference struct {
	Pin        ResourcePin `json:"pin"`
	Field      string      `json:"field"`
	BundleKey  string      `json:"bundle_key"`
	OwnerScope string      `json:"owner_scope"`
	OwnerID    int64       `json:"owner_id"`
}

type Placement struct {
	CatalogRevision string               `json:"catalog_revision"`
	RuntimeImage    RuntimeImage         `json:"runtime_image"`
	ComputeTarget   ResourcePin          `json:"compute_target"`
	ResourceProfile *ResourcePin         `json:"resource_profile"`
	Spec            workerspec.Placement `json:"spec"`
}

type RuntimeImage struct {
	ID        int64  `json:"id"`
	Reference string `json:"reference"`
	Digest    string `json:"digest"`
}
