INSERT INTO agents (
  slug,
  name,
  description,
  launch_command,
  executable,
  adapter_id,
  is_builtin,
  is_active,
  supported_modes,
  agentfile_source
)
VALUES (
  'seedance-expert',
  'Seedance Expert',
  'Plans, generates, resumes, and reviews Seedance video tasks.',
  'do-agent',
  'do-agent',
  'do-agent-acp',
  true,
  true,
  'pty,acp',
  E'AGENT do-agent\nEXECUTABLE do-agent\n\nMODE pty\nMODE acp "acp"\n\nCONFIG model STRING = ""\n\nENV DO_AGENT_HOME = sandbox.root + "/seedance-expert-home"\nENV DO_AGENT_SETTINGS = sandbox.root + "/seedance-expert-home/settings.json"\nENV DO_AGENT_LOG_DIR = sandbox.root + "/seedance-expert-home/logs"\nENV OPENAI_API_KEY SECRET OPTIONAL\nENV ANTHROPIC_API_KEY SECRET OPTIONAL\nENV SEEDANCE_API_KEY SECRET OPTIONAL\nENV SEEDANCE_BASE_URL TEXT OPTIONAL\nENV SEEDANCE_MODEL TEXT OPTIONAL\n\nPROMPT_POSITION prepend\nMCP ON\n\narg "--model" config.model when config.model != ""\nmkdir sandbox.root + "/seedance-expert-home"\n\nif config_json {\n  file sandbox.root + "/seedance-expert-home/settings.json" json(config_json)\n}\n\nif mcp.enabled {\n  mkdir sandbox.work_dir + "/.agent"\n  file sandbox.work_dir + "/.agent/config.json" json({ mcpServers: mcp.servers })\n}\n\nCAPABILITY resume acp\nCAPABILITY permission notification\nCAPABILITY usage live\nCAPABILITY control set_model,set_execution_mode\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY subagents false\nCAPABILITY model_family multi\n'
);
