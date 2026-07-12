CREATE TABLE marketplace.marketplace_quota_reservations (
    id UUID PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    quota_account_id UUID NOT NULL,
    reservation_type VARCHAR(20) NOT NULL
        CHECK (reservation_type IN ('installation', 'runtime_execution')),
    subject_ref VARCHAR(100) NOT NULL,
    idempotency_key UUID NOT NULL,
    reserved_credits NUMERIC(20,6) NOT NULL CHECK (reserved_credits > 0),
    status VARCHAR(16) NOT NULL DEFAULT 'held'
        CHECK (status IN ('held', 'settled', 'released', 'expired')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_marketplace_quota_reservations_idempotency UNIQUE (idempotency_key),
    UNIQUE (quota_account_id, id),
    FOREIGN KEY (marketplace_id, quota_account_id)
        REFERENCES marketplace.marketplace_quota_accounts(marketplace_id, id)
);

CREATE TABLE marketplace.marketplace_quota_ledger_entries (
    id UUID PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    quota_account_id UUID NOT NULL,
    entry_type VARCHAR(20) NOT NULL
        CHECK (entry_type IN ('grant', 'reserve', 'debit', 'release', 'adjust', 'grant_expire')),
    available_delta NUMERIC(20,6) NOT NULL DEFAULT 0,
    reserved_delta NUMERIC(20,6) NOT NULL DEFAULT 0,
    consumed_delta NUMERIC(20,6) NOT NULL DEFAULT 0 CHECK (consumed_delta >= 0),
    shortfall_delta NUMERIC(20,6) NOT NULL DEFAULT 0 CHECK (shortfall_delta >= 0),
    reservation_id UUID,
    usage_event_id UUID,
    operation_id UUID,
    period_start TIMESTAMPTZ,
    reason VARCHAR(240) NOT NULL,
    created_by_platform_user_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (entry_type <> 'grant' OR period_start IS NOT NULL),
    FOREIGN KEY (marketplace_id, quota_account_id)
        REFERENCES marketplace.marketplace_quota_accounts(marketplace_id, id),
    FOREIGN KEY (quota_account_id, reservation_id)
        REFERENCES marketplace.marketplace_quota_reservations(quota_account_id, id),
    FOREIGN KEY (marketplace_id, operation_id)
        REFERENCES marketplace.marketplace_installation_operations(marketplace_id, id)
);

CREATE UNIQUE INDEX idx_marketplace_quota_grants_period
    ON marketplace.marketplace_quota_ledger_entries (quota_account_id, period_start)
    WHERE entry_type = 'grant';

CREATE FUNCTION marketplace.enforce_quota_non_negative_balance() RETURNS TRIGGER AS $$
DECLARE
    available_balance NUMERIC(20,6);
    reserved_balance NUMERIC(20,6);
BEGIN
    PERFORM 1
    FROM marketplace.marketplace_quota_accounts
    WHERE marketplace_id = NEW.marketplace_id AND id = NEW.quota_account_id
    FOR UPDATE;

    SELECT
        COALESCE(SUM(available_delta), 0) + NEW.available_delta,
        COALESCE(SUM(reserved_delta), 0) + NEW.reserved_delta
    INTO available_balance, reserved_balance
    FROM marketplace.marketplace_quota_ledger_entries
    WHERE quota_account_id = NEW.quota_account_id;

    IF available_balance < 0 OR reserved_balance < 0 THEN
        RAISE EXCEPTION 'quota balance cannot be negative';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER marketplace_quota_balance_guard
BEFORE INSERT ON marketplace.marketplace_quota_ledger_entries
FOR EACH ROW EXECUTE FUNCTION marketplace.enforce_quota_non_negative_balance();

CREATE FUNCTION marketplace.prevent_quota_ledger_mutation() RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'quota ledger entries are immutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER marketplace_quota_ledger_immutable
BEFORE UPDATE OR DELETE ON marketplace.marketplace_quota_ledger_entries
FOR EACH ROW EXECUTE FUNCTION marketplace.prevent_quota_ledger_mutation();

CREATE TABLE marketplace.marketplace_audit_events (
    id UUID PRIMARY KEY,
    marketplace_id BIGINT NOT NULL REFERENCES marketplace.marketplaces(id),
    actor_platform_user_id BIGINT,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(100) NOT NULL,
    target_ref VARCHAR(100) NOT NULL,
    old_data JSONB,
    new_data JSONB,
    ip_address INET,
    user_agent VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
