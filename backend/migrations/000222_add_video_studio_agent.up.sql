INSERT INTO agents (
    slug,
    name,
    description,
    launch_command,
    executable,
    is_builtin,
    is_active,
    supported_modes,
    agentfile_source
)
SELECT
    'video-studio',
    'Video Studio',
    'Codex-based video production runtime with FFmpeg, Chromium, Remotion, Python, and CJK fonts.',
    launch_command,
    'video-studio-codex',
    true,
    true,
    supported_modes,
    agentfile_source
FROM agents
WHERE slug = 'codex-cli';

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM agents WHERE slug = 'video-studio') THEN
        RAISE EXCEPTION 'codex-cli must exist before video-studio is registered';
    END IF;
END
$$;
