ALTER TABLE im_thread_mappings DROP COLUMN IF EXISTS context_token;

ALTER TABLE im_channel_connections DROP CONSTRAINT IF EXISTS im_channel_connections_provider_check;
ALTER TABLE im_channel_connections ADD CONSTRAINT im_channel_connections_provider_check
    CHECK (provider IN ('feishu', 'dingtalk', 'wecom', 'slack'));

DELETE FROM im_channel_connections WHERE provider IN ('weixin', 'wechat');
