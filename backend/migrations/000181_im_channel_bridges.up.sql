-- External IM channel bridges: Feishu / DingTalk / WeCom / Slack ↔ collaboration channels.
-- Design follows the OpenClaw/claw-connect provider registry pattern: one service,
-- many providers, shared binding + webhook pipeline.

CREATE TABLE im_channel_connections (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL,
    name VARCHAR(255) NOT NULL,
    channel_id BIGINT REFERENCES channels(id) ON DELETE SET NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    webhook_token VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'disabled',
    last_error TEXT,
    created_by_user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (organization_id, provider, name),
    CHECK (provider IN ('feishu', 'dingtalk', 'wecom', 'slack')),
    CHECK (status IN ('disabled', 'active', 'error'))
);

CREATE INDEX im_channel_connections_org ON im_channel_connections (organization_id);
CREATE INDEX im_channel_connections_provider ON im_channel_connections (provider);
CREATE UNIQUE INDEX im_channel_connections_webhook_token ON im_channel_connections (webhook_token);

CREATE TABLE im_thread_mappings (
    id BIGSERIAL PRIMARY KEY,
    connection_id BIGINT NOT NULL REFERENCES im_channel_connections(id) ON DELETE CASCADE,
    external_thread_id VARCHAR(512) NOT NULL,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (connection_id, external_thread_id)
);

CREATE INDEX im_thread_mappings_channel ON im_thread_mappings (channel_id);

CREATE TRIGGER update_im_channel_connections_updated_at
    BEFORE UPDATE ON im_channel_connections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE im_channel_connections IS 'Org-scoped IM bridge connections (Feishu/DingTalk/WeCom/Slack) mapped to internal collaboration channels.';
COMMENT ON TABLE im_thread_mappings IS 'Maps external IM thread/chat IDs to internal channel IDs per connection.';
