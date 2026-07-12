package agentpod

import (
	"context"
	"strconv"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	envbundleservice "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing.
// Delegates to testkit.SetupTestDB for shared schema.
func setupTestDB(t *testing.T) *gorm.DB {
	db := testkit.SetupTestDB(t)

	db.Exec(`
INSERT INTO execution_clusters (id, organization_id, slug, name, kind, status)
VALUES (1, 1, 'local', 'Local cluster', 'local', 'ready'),
       (2, 2, 'local', 'Local cluster', 'local', 'ready'),
       (3, 3, 'local', 'Local cluster', 'local', 'ready')
`)
	db.Exec(`
INSERT INTO runners (id, organization_id, cluster_id, node_id, status, current_pods)
VALUES (1, 1, 1, 'runner-001', 'online', 0)
`)
	db.Exec("INSERT INTO users (id, username, name, email) VALUES (1, 'testuser', 'Test User', 'test@example.com')")

	return db
}

// Helper functions
func intPtr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// newTestPodService wraps *gorm.DB into PodRepository for testing.
func newTestPodService(db *gorm.DB) *PodService {
	return NewPodService(infra.NewPodRepository(db))
}

func seedTestRunner(t *testing.T, db *gorm.DB, id, organizationID int64) {
	t.Helper()
	if err := db.Exec(`
INSERT OR IGNORE INTO execution_clusters (organization_id, slug, name, kind, status)
VALUES (?, 'local', 'Local cluster', 'local', 'ready')
`, organizationID).Error; err != nil {
		t.Fatalf("seed test execution cluster: %v", err)
	}
	var clusterID int64
	if err := db.Raw(`
SELECT id FROM execution_clusters WHERE organization_id = ? AND slug = 'local'
`, organizationID).Scan(&clusterID).Error; err != nil {
		t.Fatalf("find test execution cluster: %v", err)
	}
	if err := db.Exec(`
INSERT INTO runners (id, organization_id, cluster_id, node_id, status, current_pods)
VALUES (?, ?, ?, ?, 'online', 0)
`, id, organizationID, clusterID, "runner-"+strconv.FormatInt(id, 10)).Error; err != nil {
		t.Fatalf("seed test runner: %v", err)
	}
}

// newTestSettingsService wraps *gorm.DB into SettingsRepository for testing.
func newTestSettingsService(db *gorm.DB) *SettingsService {
	return NewSettingsService(infra.NewSettingsRepository(db))
}

// newTestAIProviderService wraps *gorm.DB into AIProviderRepository for testing.
// Accepts nil db for tests that don't hit the DB (pure logic tests).
func newTestAIProviderService(db *gorm.DB, enc *crypto.Encryptor) *AIProviderService {
	if db == nil {
		return NewAIProviderService(nil, enc)
	}
	return NewAIProviderService(infra.NewAIProviderRepository(db), enc)
}

// newTestAutopilotService wraps *gorm.DB into AutopilotRepository for testing.
func newTestAutopilotService(db *gorm.DB) *AutopilotControllerService {
	return NewAutopilotControllerService(infra.NewAutopilotRepository(db))
}

// noopBundleLoader satisfies agent.EnvBundleLoader with zero bundles. Used by
// tests where bundle wiring isn't the focus — keeps ConfigBuilder construction
// satisfied without standing up a real EnvBundle service.
type noopBundleLoader struct{}

func (noopBundleLoader) GetEffectiveForUser(_ context.Context, _, _ int64, _ string) ([]*envbundleservice.EffectiveBundle, error) {
	return nil, nil
}

func TestNewPodService(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestPodService(db)
	if svc == nil {
		t.Error("NewPodService returned nil")
	}
	if svc.repo == nil {
		t.Error("Service repo not set correctly")
	}
}

// suppress unused import for agentpod domain
var _ = agentpod.StatusRunning
