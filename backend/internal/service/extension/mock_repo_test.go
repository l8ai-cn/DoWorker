package extension

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/extension"
)

// mockExtensionRepo is a no-op implementation of the trimmed extension.Repository
// (MCP catalog + installed MCP/skill). Sync/packager tests embed it and override
// only the handful of methods they exercise.
type mockExtensionRepo struct{}

func newMockExtensionRepo() *mockExtensionRepo { return &mockExtensionRepo{} }

func (m *mockExtensionRepo) ListMcpMarketItems(_ context.Context, _ string, _ string, _, _ int) ([]*extension.McpMarketItem, int64, error) {
	return nil, 0, nil
}

func (m *mockExtensionRepo) GetMcpMarketItem(_ context.Context, _ int64) (*extension.McpMarketItem, error) {
	return nil, nil
}

func (m *mockExtensionRepo) FindMcpMarketItemByRegistryName(_ context.Context, _ string) (*extension.McpMarketItem, error) {
	return nil, errors.New("not found")
}

func (m *mockExtensionRepo) UpsertMcpMarketItem(_ context.Context, _ *extension.McpMarketItem) error {
	return nil
}

func (m *mockExtensionRepo) BatchUpsertMcpMarketItems(_ context.Context, _ []*extension.McpMarketItem) error {
	return nil
}

func (m *mockExtensionRepo) DeactivateMcpMarketItemsNotIn(_ context.Context, _ string, _ []string) (int64, error) {
	return 0, nil
}

func (m *mockExtensionRepo) ListInstalledMcpServers(_ context.Context, _, _, _ int64, _ string) ([]*extension.InstalledMcpServer, error) {
	return nil, nil
}

func (m *mockExtensionRepo) GetInstalledMcpServer(_ context.Context, _ int64) (*extension.InstalledMcpServer, error) {
	return nil, nil
}

func (m *mockExtensionRepo) CreateInstalledMcpServer(_ context.Context, _ *extension.InstalledMcpServer) error {
	return nil
}

func (m *mockExtensionRepo) UpdateInstalledMcpServer(_ context.Context, _ *extension.InstalledMcpServer) error {
	return nil
}

func (m *mockExtensionRepo) DeleteInstalledMcpServer(_ context.Context, _ int64) error {
	return nil
}

func (m *mockExtensionRepo) GetEffectiveMcpServers(_ context.Context, _, _, _ int64) ([]*extension.InstalledMcpServer, error) {
	return nil, nil
}

func (m *mockExtensionRepo) ListInstalledSkills(_ context.Context, _, _, _ int64, _ string) ([]*extension.InstalledSkill, error) {
	return nil, nil
}

func (m *mockExtensionRepo) GetInstalledSkill(_ context.Context, _ int64) (*extension.InstalledSkill, error) {
	return nil, nil
}

func (m *mockExtensionRepo) CreateInstalledSkill(_ context.Context, _ *extension.InstalledSkill) error {
	return nil
}

func (m *mockExtensionRepo) UpdateInstalledSkill(_ context.Context, _ *extension.InstalledSkill) error {
	return nil
}

func (m *mockExtensionRepo) DeleteInstalledSkill(_ context.Context, _ int64) error {
	return nil
}

func (m *mockExtensionRepo) GetEffectiveSkills(_ context.Context, _, _, _ int64) ([]*extension.InstalledSkill, error) {
	return nil, nil
}

var _ extension.Repository = (*mockExtensionRepo)(nil)
