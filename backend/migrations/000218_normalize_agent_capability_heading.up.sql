UPDATE agents
SET agentfile_source = replace(
  agentfile_source,
  '# === Hive Capabilities ===',
  '# === Agent Capabilities ==='
),
updated_at = NOW()
WHERE agentfile_source LIKE '%# === Hive Capabilities ===%';
