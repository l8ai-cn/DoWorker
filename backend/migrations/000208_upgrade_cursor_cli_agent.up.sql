UPDATE agents
SET
  launch_command = 'agent',
  executable = 'agent',
  adapter_id = 'cursor-acp',
  supported_modes = 'pty,acp',
  agentfile_source = E'# === Identity ===\nAGENT agent\nEXECUTABLE agent\n\n# === Mode ===\nMODE pty\nMODE acp "acp"\n\n# === Environment ===\nENV CURSOR_API_KEY SECRET OPTIONAL\n\n# === Prompt ===\nPROMPT_POSITION prepend\n\nCAPABILITY resume none\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family multi\n',
  updated_at = NOW()
WHERE slug = 'cursor-cli';
