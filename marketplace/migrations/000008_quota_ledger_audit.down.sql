DROP TRIGGER IF EXISTS marketplace_quota_ledger_immutable
    ON marketplace.marketplace_quota_ledger_entries;
DROP FUNCTION IF EXISTS marketplace.prevent_quota_ledger_mutation();

DROP TRIGGER IF EXISTS marketplace_quota_balance_guard
    ON marketplace.marketplace_quota_ledger_entries;
DROP FUNCTION IF EXISTS marketplace.enforce_quota_non_negative_balance();

DROP TABLE IF EXISTS marketplace.marketplace_audit_events;
DROP TABLE IF EXISTS marketplace.marketplace_quota_ledger_entries;
DROP TABLE IF EXISTS marketplace.marketplace_quota_reservations;
