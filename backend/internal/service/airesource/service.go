package airesource

import (
	"context"
	"fmt"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/organization"
	"github.com/l8ai-cn/agentcloud/backend/pkg/audit"
)

type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type OrganizationMemberReader interface {
	GetMember(ctx context.Context, orgID, userID int64) (*organization.Member, error)
}

type ConnectionProber interface {
	Probe(ctx context.Context, input ProbeInput) error
}

type AuditRecorder interface {
	Record(ctx context.Context, log *audit.Log) error
}

type MutationRunner interface {
	Run(ctx context.Context, mutation func(repo domain.Repository, audit AuditRecorder) error) error
}

type EndpointValidator interface {
	Validate(ctx context.Context, rawURL string) error
}

type Dependencies struct {
	Repository domain.Repository
	Cipher     Cipher
	Members    OrganizationMemberReader
	Prober     ConnectionProber
	Mutations  MutationRunner
	Endpoints  EndpointValidator
}

type Service struct {
	repository domain.Repository
	cipher     Cipher
	members    OrganizationMemberReader
	prober     ConnectionProber
	mutations  MutationRunner
	endpoints  EndpointValidator
}

func NewService(deps Dependencies) (*Service, error) {
	if deps.Repository == nil {
		return nil, fmt.Errorf("AI resource repository is required")
	}
	if deps.Cipher == nil {
		return nil, fmt.Errorf("AI resource cipher is required")
	}
	if deps.Members == nil {
		return nil, fmt.Errorf("organization member reader is required")
	}
	if deps.Prober == nil {
		return nil, fmt.Errorf("AI resource connection prober is required")
	}
	if deps.Mutations == nil {
		return nil, fmt.Errorf("AI resource mutation runner is required")
	}
	if deps.Endpoints == nil {
		return nil, fmt.Errorf("AI resource endpoint validator is required")
	}
	return &Service{repository: deps.Repository, cipher: deps.Cipher, members: deps.Members, prober: deps.Prober, mutations: deps.Mutations, endpoints: deps.Endpoints}, nil
}

func (s *Service) Catalog() []domain.ProviderDefinition { return domain.Providers() }
