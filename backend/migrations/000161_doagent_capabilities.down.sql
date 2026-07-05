UPDATE agents
SET agentfile_source = regexp_replace(
  agentfile_source,
  E'\n\n# === Hive Capabilities ===\nCAPABILITY resume none\nCAPABILITY permission notification\nCAPABILITY usage live\nCAPABILITY control set_model,set_execution_mode\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY subagents false\nCAPABILITY model_family multi\n',
  '',
  'g'
),
updated_at = NOW()
WHERE slug = 'do-agent' AND is_builtin = true;
