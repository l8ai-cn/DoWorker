package testkit

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"
)

func CreateUser(t *testing.T, db *gorm.DB, email, username string) (id int64) {
	t.Helper()
	result := db.Exec(
		`INSERT INTO users (email, username, name, password_hash, is_active, is_email_verified) VALUES (?, ?, ?, ?, 1, 1)`,
		email, username, username, "$2a$10$dummyhash",
	)
	if result.Error != nil {
		t.Fatalf("testkit.CreateUser: %v", result.Error)
	}
	db.Raw(`SELECT id FROM users WHERE email = ?`, email).Scan(&id)
	return id
}

func CreateOrg(t *testing.T, db *gorm.DB, slug string, ownerID int64) (id int64) {
	t.Helper()
	result := db.Exec(`INSERT INTO organizations (name, slug) VALUES (?, ?)`, "Org "+slug, slug)
	if result.Error != nil {
		t.Fatalf("testkit.CreateOrg: %v", result.Error)
	}
	db.Raw(`SELECT id FROM organizations WHERE slug = ?`, slug).Scan(&id)
	if ownerID > 0 {
		db.Exec(`INSERT INTO organization_members (organization_id, user_id, role) VALUES (?, ?, 'owner')`, id, ownerID)
	}
	result = db.Exec(`
		INSERT INTO execution_clusters (organization_id, slug, name, kind, status)
		VALUES (?, 'local', 'Local cluster', 'local', 'ready'),
		       (?, 'online', 'Online cluster', 'online', 'ready')
	`, id, id)
	if result.Error != nil {
		t.Fatalf("testkit.CreateOrg clusters: %v", result.Error)
	}
	return id
}

func CreateRunner(t *testing.T, db *gorm.DB, orgID int64, nodeID string) (id int64) {
	t.Helper()
	var clusterID int64
	db.Raw(`SELECT id FROM execution_clusters WHERE organization_id = ? AND slug = 'online'`, orgID).Scan(&clusterID)
	result := db.Exec(
		`INSERT INTO runners (organization_id, cluster_id, node_id, status, max_concurrent_pods) VALUES (?, ?, ?, 'online', 5)`,
		orgID, clusterID, nodeID,
	)
	if result.Error != nil {
		t.Fatalf("testkit.CreateRunner: %v", result.Error)
	}
	db.Raw(`SELECT id FROM runners WHERE node_id = ?`, nodeID).Scan(&id)
	return id
}

func CreatePod(t *testing.T, db *gorm.DB, orgID, runnerID, userID int64) (podKey string) {
	t.Helper()
	podKey = fmt.Sprintf("pod-%d-%d", time.Now().UnixNano(), userID)
	var clusterID int64
	db.Raw(`SELECT cluster_id FROM runners WHERE id = ?`, runnerID).Scan(&clusterID)
	result := db.Exec(
		`INSERT INTO pods (organization_id, pod_key, runner_id, cluster_id, created_by_id, status) VALUES (?, ?, ?, ?, ?, 'initializing')`,
		orgID, podKey, runnerID, clusterID, userID,
	)
	if result.Error != nil {
		t.Fatalf("testkit.CreatePod: %v", result.Error)
	}
	return podKey
}

func CreateAgent(t *testing.T, db *gorm.DB, slug, name, agentfileSrc string) {
	t.Helper()
	CreateAgentWithLegacyFlag(t, db, slug, name, agentfileSrc, false)
}

func CreateAgentWithLegacyFlag(t *testing.T, db *gorm.DB, slug, name, agentfileSrc string, usesLegacy bool) {
	t.Helper()
	flag := 0
	if usesLegacy {
		flag = 1
	}
	result := db.Exec(
		`INSERT INTO agents (slug, name, launch_command, agentfile_source, supported_modes, uses_legacy_columns) VALUES (?, ?, ?, ?, 'pty', ?)`,
		slug, name, slug, agentfileSrc, flag,
	)
	if result.Error != nil {
		t.Fatalf("testkit.CreateAgent: %v", result.Error)
	}
}

func CreateChannel(t *testing.T, db *gorm.DB, orgID int64, name string) (id int64) {
	t.Helper()
	result := db.Exec(
		`INSERT INTO channels (organization_id, name) VALUES (?, ?)`, orgID, name,
	)
	if result.Error != nil {
		t.Fatalf("testkit.CreateChannel: %v", result.Error)
	}
	db.Raw(`SELECT id FROM channels WHERE organization_id = ? AND name = ?`, orgID, name).Scan(&id)
	return id
}

func CreateTicket(t *testing.T, db *gorm.DB, orgID, reporterID int64, title string) (id int64) {
	t.Helper()
	slug := fmt.Sprintf("T-%d", time.Now().UnixNano()%10000)
	result := db.Exec(
		`INSERT INTO tickets (organization_id, number, slug, title, reporter_id) VALUES (?, ?, ?, ?, ?)`,
		orgID, time.Now().UnixNano()%10000, slug, title, reporterID,
	)
	if result.Error != nil {
		t.Fatalf("testkit.CreateTicket: %v", result.Error)
	}
	db.Raw(`SELECT id FROM tickets WHERE slug = ?`, slug).Scan(&id)
	return id
}

func CreateRepo(t *testing.T, db *gorm.DB, orgID int64, slug, cloneURL string) (id int64) {
	t.Helper()
	result := db.Exec(
		`INSERT INTO repositories (organization_id, external_id, name, slug, http_clone_url) VALUES (?, ?, ?, ?, ?)`,
		orgID, "ext-"+slug, slug, slug, cloneURL,
	)
	if result.Error != nil {
		t.Fatalf("testkit.CreateRepo: %v", result.Error)
	}
	db.Raw(`SELECT id FROM repositories WHERE slug = ? AND organization_id = ?`, slug, orgID).Scan(&id)
	return id
}

func CreateWorkflow(t *testing.T, db *gorm.DB, orgID, userID int64, slug string) (id int64) {
	t.Helper()
	result := db.Exec(
		`INSERT INTO workflows (organization_id, name, slug, created_by_id, prompt_template) VALUES (?, ?, ?, ?, 'test prompt')`,
		orgID, "Workflow "+slug, slug, userID,
	)
	if result.Error != nil {
		t.Fatalf("testkit.CreateWorkflow: %v", result.Error)
	}
	db.Raw(`SELECT id FROM workflows WHERE slug = ? AND organization_id = ?`, slug, orgID).Scan(&id)
	return id
}

func SeedBillingPlans(t *testing.T, db *gorm.DB) {
	t.Helper()
	plans := []struct {
		name, display string
		maxPods       int
	}{
		{"free", "Free", 1},
		{"pro", "Pro", 10},
		{"enterprise", "Enterprise", 100},
	}
	for _, p := range plans {
		db.Exec(
			`INSERT INTO subscription_plans (name, display_name, max_concurrent_pods, is_active) VALUES (?, ?, ?, 1)`,
			p.name, p.display, p.maxPods,
		)
	}
}
