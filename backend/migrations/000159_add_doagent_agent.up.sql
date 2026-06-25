-- Register DoAgent (do-agent) as a builtin agent.
--
-- Binary on disk is `do-agent` (Rust CLI from AgentForge/doagent). Slug matches
-- the executable — unlike cursor-cli/codex-cli there is no separate marketing
-- name suffix because the upstream project brands itself as do-agent.
--
-- PTY mode launches `do-agent` in a TTY (interactive REPL when stdin is a
-- terminal). ACP mode appends the `acp` subcommand for stdio JSON-RPC.
--
-- DO_AGENT_SETTINGS isolates provider config per pod under the sandbox root.
-- MCP servers land in {work_dir}/.agent/config.json — do-agent's project overlay.

INSERT INTO agents (slug, name, launch_command, executable, is_builtin, is_active, supported_modes, agentfile_source)
VALUES ('do-agent', 'DoAgent', 'do-agent', 'do-agent', true, true, 'pty,acp',
  E'# === Identity ===\nAGENT do-agent\nEXECUTABLE do-agent\n\n# === Mode ===\nMODE pty\nMODE acp "acp"\n\n# === Configuration ===\nCONFIG model STRING = ""\n\n# === Environment ===\nENV DO_AGENT_HOME = sandbox.root + "/do-agent-home"\nENV DO_AGENT_SETTINGS = sandbox.root + "/do-agent-home/settings.json"\nENV DO_AGENT_LOG_DIR = sandbox.root + "/do-agent-home/logs"\nENV OPENAI_API_KEY SECRET OPTIONAL\nENV ANTHROPIC_API_KEY SECRET OPTIONAL\n\n# === Prompt ===\nPROMPT_POSITION prepend\n\n# === Capabilities ===\nMCP ON\n\n# === Build Logic ===\narg "--model" config.model when config.model != ""\nmkdir sandbox.root + "/do-agent-home"\n\nif config_json {\n  file sandbox.root + "/do-agent-home/settings.json" json(config_json)\n}\n\nif mcp.enabled {\n  mkdir sandbox.work_dir + "/.agent"\n  file sandbox.work_dir + "/.agent/config.json" json({ mcpServers: mcp.servers })\n}\n');
