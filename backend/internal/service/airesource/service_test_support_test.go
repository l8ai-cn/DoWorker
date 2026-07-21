package airesource

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/organization"
	"github.com/l8ai-cn/agentcloud/backend/pkg/audit"
	"github.com/l8ai-cn/agentcloud/backend/pkg/crypto"
)

var errInjected = errors.New("injected failure")

type memoryRepository struct {
	mu          sync.Mutex
	nextID      int64
	connections map[int64]*domain.Connection
	resources   map[int64]*domain.ModelResource
	err         map[string]error
	calls       map[string]int
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{nextID: 1, connections: map[int64]*domain.Connection{}, resources: map[int64]*domain.ModelResource{}, err: map[string]error{}, calls: map[string]int{}}
}

func (r *memoryRepository) failure(method string) error { return r.err[method] }

func (r *memoryRepository) GetConnectionByID(_ context.Context, id int64) (*domain.Connection, error) {
	return r.getConnection(id)
}

func (r *memoryRepository) getConnection(id int64) (*domain.Connection, error) {
	if err := r.failure("GetConnectionByID"); err != nil {
		return nil, err
	}
	value := r.connections[id]
	if value == nil {
		return nil, nil
	}
	copy := *value
	copy.ConfiguredFields = append([]string(nil), value.ConfiguredFields...)
	return &copy, nil
}

func (r *memoryRepository) CreateConnection(_ context.Context, value *domain.Connection) error {
	if err := r.failure("CreateConnection"); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	value.ID = r.nextID
	value.Revision = 1
	r.nextID++
	copy := *value
	copy.ConfiguredFields = append([]string(nil), value.ConfiguredFields...)
	r.connections[value.ID] = &copy
	return nil
}

func (r *memoryRepository) SaveConnection(_ context.Context, value *domain.Connection) error {
	if err := r.failure("SaveConnection"); err != nil {
		return err
	}
	if r.connections[value.ID] == nil {
		return errors.New("missing connection")
	}
	if r.connections[value.ID].Revision != value.Revision ||
		!r.connections[value.ID].UpdatedAt.Equal(value.UpdatedAt) {
		return domain.ErrConflict
	}
	value.Revision++
	value.UpdatedAt = time.Now().UTC()
	copy := *value
	copy.ConfiguredFields = append([]string(nil), value.ConfiguredFields...)
	r.connections[value.ID] = &copy
	return nil
}

func (r *memoryRepository) DeleteConnection(_ context.Context, id, expectedRevision int64, expectedUpdatedAt time.Time) error {
	if err := r.failure("DeleteConnection"); err != nil {
		return err
	}
	connection := r.connections[id]
	if connection == nil {
		return errors.New("missing connection")
	}
	if connection.Revision != expectedRevision || !connection.UpdatedAt.Equal(expectedUpdatedAt) {
		return domain.ErrConflict
	}
	delete(r.connections, id)
	for resourceID, resource := range r.resources {
		if resource.ProviderConnectionID == id {
			delete(r.resources, resourceID)
		}
	}
	return nil
}

func (r *memoryRepository) ListConnectionsByOwner(_ context.Context, scope domain.OwnerScope, ownerID int64) ([]*domain.Connection, error) {
	if err := r.failure("ListConnectionsByOwner"); err != nil {
		return nil, err
	}
	values := make([]*domain.Connection, 0)
	for id, value := range r.connections {
		if value.OwnerScope == scope && value.OwnerID == ownerID {
			copy, _ := r.getConnection(id)
			values = append(values, copy)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].ID < values[j].ID })
	return values, nil
}

func (r *memoryRepository) GetResourceByID(_ context.Context, id int64) (*domain.ModelResource, error) {
	if err := r.failure("GetResourceByID"); err != nil {
		return nil, err
	}
	value := r.resources[id]
	if value == nil {
		return nil, nil
	}
	return cloneResource(value), nil
}

func (r *memoryRepository) CreateResource(_ context.Context, value *domain.ModelResource) error {
	if err := r.failure("CreateResource"); err != nil {
		return err
	}
	value.ID = r.nextID
	value.Revision = 1
	r.nextID++
	r.resources[value.ID] = cloneResource(value)
	return nil
}

func (r *memoryRepository) SaveResource(_ context.Context, value *domain.ModelResource) error {
	if err := r.failure("SaveResource"); err != nil {
		return err
	}
	if r.resources[value.ID] == nil {
		return errors.New("missing resource")
	}
	if r.resources[value.ID].Revision != value.Revision ||
		!r.resources[value.ID].UpdatedAt.Equal(value.UpdatedAt) {
		return domain.ErrConflict
	}
	value.Revision++
	value.UpdatedAt = time.Now().UTC()
	r.resources[value.ID] = cloneResource(value)
	return nil
}

func (r *memoryRepository) DeleteResource(_ context.Context, id, expectedRevision int64, expectedUpdatedAt time.Time) error {
	if err := r.failure("DeleteResource"); err != nil {
		return err
	}
	resource := r.resources[id]
	if resource == nil {
		return errors.New("missing resource")
	}
	if resource.Revision != expectedRevision || !resource.UpdatedAt.Equal(expectedUpdatedAt) {
		return domain.ErrConflict
	}
	delete(r.resources, id)
	return nil
}

func (r *memoryRepository) ListResourcesByConnection(_ context.Context, connectionID int64) ([]*domain.ModelResource, error) {
	r.calls["ListResourcesByConnection"]++
	if err := r.failure("ListResourcesByConnection"); err != nil {
		return nil, err
	}
	values := make([]*domain.ModelResource, 0)
	for _, value := range r.resources {
		if value.ProviderConnectionID == connectionID {
			values = append(values, cloneResource(value))
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].ID < values[j].ID })
	return values, nil
}

func (r *memoryRepository) ListResourcesByOwner(_ context.Context, scope domain.OwnerScope, ownerID int64) ([]*domain.ModelResource, error) {
	r.calls["ListResourcesByOwner"]++
	values := make([]*domain.ModelResource, 0)
	for _, value := range r.resources {
		connection := r.connections[value.ProviderConnectionID]
		if connection != nil && connection.OwnerScope == scope && connection.OwnerID == ownerID {
			values = append(values, cloneResource(value))
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].ID < values[j].ID })
	return values, nil
}

func (r *memoryRepository) ListEffective(_ context.Context, userID, orgID int64, modalities []domain.Modality) ([]*domain.ModelResource, error) {
	if err := r.failure("ListEffective"); err != nil {
		return nil, err
	}
	wanted := map[domain.Modality]bool{}
	for _, modality := range modalities {
		wanted[modality] = true
	}
	values := make([]*domain.ModelResource, 0)
	for _, resource := range r.resources {
		connection := r.connections[resource.ProviderConnectionID]
		if connection == nil || !connection.IsEnabled || !resource.IsEnabled ||
			connection.Status != domain.ConnectionStatusValid || resource.Status != domain.ConnectionStatusValid {
			continue
		}
		if !((connection.OwnerScope == domain.OwnerScopeUser && connection.OwnerID == userID) || (connection.OwnerScope == domain.OwnerScopeOrg && connection.OwnerID == orgID)) {
			continue
		}
		if len(wanted) > 0 && !supportsAny(resource.Modalities, wanted) {
			continue
		}
		values = append(values, cloneResource(resource))
	}
	sort.Slice(values, func(i, j int) bool { return values[i].ID < values[j].ID })
	return values, nil
}

func (r *memoryRepository) SetDefault(_ context.Context, resourceID int64, modality domain.Modality) error {
	if err := r.failure("SetDefault"); err != nil {
		return err
	}
	resource := r.resources[resourceID]
	if resource == nil {
		return errors.New("missing resource")
	}
	connection := r.connections[resource.ProviderConnectionID]
	for _, candidate := range r.resources {
		candidateConnection := r.connections[candidate.ProviderConnectionID]
		if candidateConnection != nil && candidateConnection.OwnerScope == connection.OwnerScope && candidateConnection.OwnerID == connection.OwnerID {
			candidate.DefaultModalities = removeModality(candidate.DefaultModalities, modality)
		}
	}
	resource.DefaultModalities = append(resource.DefaultModalities, modality)
	return nil
}

func (r *memoryRepository) SetValidationState(_ context.Context, connectionID, expectedRevision int64, expectedCredentialsEncrypted string, status domain.ConnectionStatus, at time.Time, validationError string) (int64, error) {
	if err := r.failure("SetValidationState"); err != nil {
		return 0, err
	}
	connection := r.connections[connectionID]
	if connection == nil {
		return 0, errors.New("missing connection")
	}
	if connection.Revision != expectedRevision || connection.CredentialsEncrypted != expectedCredentialsEncrypted {
		return 0, domain.ErrConflict
	}
	connection.Status, connection.LastValidatedAt, connection.ValidationError = status, &at, validationError
	for _, resource := range r.resources {
		if resource.ProviderConnectionID == connectionID {
			resource.Status, resource.LastValidatedAt, resource.ValidationError = status, &at, validationError
		}
	}
	return connection.Revision, nil
}

func cloneResource(value *domain.ModelResource) *domain.ModelResource {
	copy := *value
	copy.Modalities = append([]domain.Modality(nil), value.Modalities...)
	copy.Capabilities = append([]domain.Capability(nil), value.Capabilities...)
	copy.DefaultModalities = append([]domain.Modality(nil), value.DefaultModalities...)
	return &copy
}

func supportsAny(values []domain.Modality, wanted map[domain.Modality]bool) bool {
	for _, value := range values {
		if wanted[value] {
			return true
		}
	}
	return false
}

func removeModality(values []domain.Modality, unwanted domain.Modality) []domain.Modality {
	result := values[:0]
	for _, value := range values {
		if value != unwanted {
			result = append(result, value)
		}
	}
	return result
}

type memberReader struct {
	members map[[2]int64]string
	err     error
}

func (r *memberReader) GetMember(_ context.Context, orgID, userID int64) (*organization.Member, error) {
	if r.err != nil {
		return nil, r.err
	}
	role, ok := r.members[[2]int64{orgID, userID}]
	if !ok {
		return nil, organization.ErrMemberNotFound
	}
	return &organization.Member{OrganizationID: orgID, UserID: userID, Role: role}, nil
}

type recordingProber struct {
	calls []ProbeInput
	err   error
}

func (p *recordingProber) Probe(_ context.Context, input ProbeInput) error {
	p.calls = append(p.calls, input)
	return p.err
}

type recordingAudit struct {
	logs       []*audit.Log
	err        error
	calls      int
	failOnCall int
}

func (r *recordingAudit) Record(_ context.Context, log *audit.Log) error {
	r.calls++
	r.logs = append(r.logs, log)
	if r.err != nil && (r.failOnCall == 0 || r.calls == r.failOnCall) {
		return r.err
	}
	return nil
}

type memoryMutationRunner struct {
	repo  *memoryRepository
	audit *recordingAudit
}

func (runner *memoryMutationRunner) Run(ctx context.Context, mutation func(domain.Repository, AuditRecorder) error) error {
	nextID := runner.repo.nextID
	connections := make(map[int64]*domain.Connection, len(runner.repo.connections))
	for id := range runner.repo.connections {
		connections[id], _ = runner.repo.getConnection(id)
	}
	resources := make(map[int64]*domain.ModelResource, len(runner.repo.resources))
	for id, resource := range runner.repo.resources {
		resources[id] = cloneResource(resource)
	}
	logs := append([]*audit.Log(nil), runner.audit.logs...)
	if err := mutation(runner.repo, runner.audit); err != nil {
		runner.repo.nextID, runner.repo.connections, runner.repo.resources = nextID, connections, resources
		runner.audit.logs = logs
		return err
	}
	return nil
}

type allowingEndpoints struct{ err error }

func (p allowingEndpoints) Validate(context.Context, string) error { return p.err }

type fixture struct {
	service   *Service
	repo      *memoryRepository
	prober    *recordingProber
	audit     *recordingAudit
	mutations *memoryMutationRunner
	members   *memberReader
	cipher    *crypto.Encryptor
}

func newFixture() fixture {
	repo := newMemoryRepository()
	prober := &recordingProber{}
	recorder := &recordingAudit{}
	mutations := &memoryMutationRunner{repo: repo, audit: recorder}
	members := &memberReader{members: map[[2]int64]string{{10, 1}: organization.RoleOwner, {10, 2}: organization.RoleAdmin, {10, 3}: organization.RoleMember}}
	cipher := crypto.NewEncryptor("ai-resource-service-test-key")
	service, err := NewService(Dependencies{Repository: repo, Cipher: cipher, Members: members, Prober: prober, Mutations: mutations, Endpoints: allowingEndpoints{}})
	if err != nil {
		panic(err)
	}
	return fixture{service: service, repo: repo, prober: prober, audit: recorder, mutations: mutations, members: members, cipher: cipher}
}

func actor(userID int64) Actor { return Actor{UserID: userID, CorrelationID: "request-123"} }
