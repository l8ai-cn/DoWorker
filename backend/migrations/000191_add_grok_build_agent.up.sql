INSERT INTO agents (
    slug, name, launch_command, executable,
    is_builtin, is_active, supported_modes, agentfile_source
)
VALUES (
    'grok-build',
    'Grok Build',
    'grok',
    'grok',
    true,
    true,
    'pty,acp',
    E'AGENT grok\nEXECUTABLE grok\nMODE pty\nMODE pty "--no-auto-update"\nMODE acp "--no-auto-update" "agent" "stdio"\nCONFIG model STRING = ""\nCONFIG effort SELECT("", "low", "medium", "high") = ""\nENV XAI_API_KEY SECRET\nENV GROK_HOME = sandbox.root + "/grok-home"\nPROMPT_POSITION prepend\nMCP ON\nSKILLS am-delegate, am-channel\nCAPABILITY resume none\nCAPABILITY permission none\nCAPABILITY usage none\nCAPABILITY interrupt true\nCAPABILITY streaming true\nCAPABILITY subagents true\nCAPABILITY model_family multi\narg "--model" config.model when config.model != ""\narg "--effort" config.effort when config.effort != ""\nmkdir sandbox.root + "/grok-home"\n'
);
