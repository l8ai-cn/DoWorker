-- WeChat (Weixin/iLink) support for IM channel bridges.
-- OpenClaw uses @tencent-weixin/openclaw-weixin with iLink long-polling, not webhooks.

ALTER TABLE im_channel_connections DROP CONSTRAINT IF EXISTS im_channel_connections_provider_check;
ALTER TABLE im_channel_connections ADD CONSTRAINT im_channel_connections_provider_check
    CHECK (provider IN ('feishu', 'dingtalk', 'wecom', 'slack', 'weixin', 'wechat'));

ALTER TABLE im_thread_mappings ADD COLUMN IF NOT EXISTS context_token VARCHAR(512);

COMMENT ON COLUMN im_thread_mappings.context_token IS 'Weixin iLink context_token for outbound replies per peer thread.';
