UPDATE agents
SET
  launch_command = 'cursor-agent',
  executable = 'cursor-agent',
  adapter_id = 'cursor-pty',
  supported_modes = 'pty',
  agentfile_source = E'# === Identity ===\nAGENT cursor-agent\nEXECUTABLE cursor-agent\n\n# === Mode ===\nMODE pty\n\n# === Environment ===\nENV CURSOR_API_KEY SECRET OPTIONAL\n\n# === Prompt ===\nPROMPT_POSITION prepend\n\nCAPABILITY resume none\nCAPABILITY permission acp\nCAPABILITY usage live\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY model_family multi\n',
  updated_at = NOW()
WHERE slug = 'cursor-cli';
