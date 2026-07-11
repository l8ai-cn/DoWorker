-- Register MiniMax CLI (mmx from npm mmx-cli) as a builtin PTY agent.
-- Slug keeps the -cli suffix; AGENT/EXECUTABLE must be the binary name `mmx`.

-- Bare `mmx` only prints help and exits; interactive Worker needs `text repl`.
INSERT INTO agents (slug, name, launch_command, executable, is_builtin, is_active, supported_modes, agentfile_source)
VALUES ('minimax-cli', 'MiniMax CLI', 'mmx', 'mmx', true, true, 'pty',
  E'AGENT mmx\nEXECUTABLE mmx\nMODE pty\nENV MINIMAX_API_KEY SECRET OPTIONAL\nENV MINIMAX_REGION TEXT OPTIONAL\nPROMPT_POSITION prepend\narg "text"\narg "repl"\n');
