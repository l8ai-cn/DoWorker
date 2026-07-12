package migrations

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEntitlementInstallationQuotaMigrationCreatesMinimumTables(t *testing.T) {
	foundation := readMigration(t, "000006_quota_foundation.up.sql")
	installation := readMigration(t, "000007_entitlement_installation.up.sql")
	ledger := readMigration(t, "000008_quota_ledger_audit.up.sql")
	sql := foundation + installation + ledger

	for _, table := range []string{
		"marketplace_entitlements",
		"marketplace_installations",
		"marketplace_installation_operations",
		"marketplace_quota_plans",
		"marketplace_quota_accounts",
		"marketplace_quota_reservations",
		"marketplace_quota_ledger_entries",
		"marketplace_audit_events",
	} {
		require.Contains(t, sql, "CREATE TABLE marketplace."+table)
	}

	for _, fragment := range []string{
		"FOREIGN KEY (marketplace_id, listing_id)",
		"REFERENCES marketplace.marketplace_listings(marketplace_id, id)",
		"FOREIGN KEY (listing_id, listing_version_id)",
		"REFERENCES marketplace.marketplace_listing_versions(listing_id, id)",
		"FOREIGN KEY (marketplace_id, entitlement_id)",
		"REFERENCES marketplace.marketplace_entitlements(marketplace_id, id)",
		"FOREIGN KEY (marketplace_id, quota_plan_id)",
		"REFERENCES marketplace.marketplace_quota_plans(marketplace_id, id)",
		"FOREIGN KEY (marketplace_id, quota_account_id)",
		"REFERENCES marketplace.marketplace_quota_accounts(marketplace_id, id)",
		"FOREIGN KEY (marketplace_id, installation_id)",
		"REFERENCES marketplace.marketplace_installations(marketplace_id, id)",
	} {
		require.Contains(t, sql, fragment)
	}
	require.NotContains(t, sql, "REFERENCES public.")

	require.Contains(t, foundation, "CREATE TABLE marketplace.marketplace_quota_plans")
	require.Contains(t, foundation, "subject_ref BIGINT NOT NULL CHECK (subject_ref > 0)")
	require.Contains(t, foundation, "CREATE TABLE marketplace.marketplace_quota_accounts")
	require.Contains(t, installation, "CREATE TABLE marketplace.marketplace_entitlements")
	require.Contains(t, installation, "CREATE TABLE marketplace.marketplace_installations")
	require.Contains(t, installation, "CREATE TABLE marketplace.marketplace_installation_operations")
	require.Contains(t, ledger, "CREATE TABLE marketplace.marketplace_quota_reservations")
	require.Contains(t, ledger, "CREATE TABLE marketplace.marketplace_quota_ledger_entries")
	require.Contains(t, ledger, "CREATE TABLE marketplace.marketplace_audit_events")
}

func TestEntitlementInstallationQuotaMigrationEnforcesIdempotencyAndBalances(t *testing.T) {
	sql := readMigration(t, "000007_entitlement_installation.up.sql") +
		readMigration(t, "000008_quota_ledger_audit.up.sql")

	for _, fragment := range []string{
		"idx_marketplace_entitlements_active_direct",
		"WHERE source = 'direct' AND status = 'active'",
		"CONSTRAINT uq_marketplace_installation_operations_idempotency UNIQUE (idempotency_key)",
		"CONSTRAINT uq_marketplace_quota_reservations_idempotency UNIQUE (idempotency_key)",
		"CREATE FUNCTION marketplace.enforce_quota_non_negative_balance()",
		"FOR UPDATE",
		"available_balance < 0 OR reserved_balance < 0",
		"RAISE EXCEPTION 'quota balance cannot be negative'",
		"CREATE TRIGGER marketplace_quota_balance_guard",
	} {
		require.Contains(t, sql, fragment)
	}
}

func TestEntitlementInstallationQuotaDownDropsDependenciesFirst(t *testing.T) {
	sql := readMigration(t, "000008_quota_ledger_audit.down.sql") +
		readMigration(t, "000007_entitlement_installation.down.sql") +
		readMigration(t, "000006_quota_foundation.down.sql")

	orderedDrops := []string{
		"DROP TABLE IF EXISTS marketplace.marketplace_audit_events",
		"DROP TABLE IF EXISTS marketplace.marketplace_quota_ledger_entries",
		"DROP TABLE IF EXISTS marketplace.marketplace_quota_reservations",
		"DROP TABLE IF EXISTS marketplace.marketplace_installation_operations",
		"DROP TABLE IF EXISTS marketplace.marketplace_installations",
		"DROP TABLE IF EXISTS marketplace.marketplace_entitlements",
		"DROP TABLE IF EXISTS marketplace.marketplace_quota_accounts",
		"DROP TABLE IF EXISTS marketplace.marketplace_quota_plans",
	}
	last := -1
	for _, fragment := range orderedDrops {
		index := strings.Index(sql, fragment)
		require.Greater(t, index, last, fragment)
		last = index
	}

	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS fk_marketplace_installations_current_operation")
	require.Contains(t, sql, "DROP FUNCTION IF EXISTS marketplace.enforce_quota_non_negative_balance()")
}

func TestEntitlementInstallationQuotaMigrationsRespectFileLimit(t *testing.T) {
	for _, name := range []string{
		"000006_quota_foundation.up.sql",
		"000006_quota_foundation.down.sql",
		"000007_entitlement_installation.up.sql",
		"000007_entitlement_installation.down.sql",
		"000008_quota_ledger_audit.up.sql",
		"000008_quota_ledger_audit.down.sql",
	} {
		sql := readMigration(t, name)
		require.LessOrEqual(t, strings.Count(sql, "\n")+1, 200, name)
	}

	for _, name := range []string{
		"000006_entitlement_installation_quota.up.sql",
		"000006_entitlement_installation_quota.down.sql",
	} {
		_, err := os.Stat(name)
		require.ErrorIs(t, err, os.ErrNotExist, name)
	}
}

func readMigration(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(name)
	require.NoError(t, err)
	return string(content)
}
