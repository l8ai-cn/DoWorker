# Agent Runtime Adapters

Canonical manifest: `loop.json`

Adapter outputs are generated artifacts. Do not edit them as the source of truth;
update `loop.json` and regenerate.

Unsupported capability behavior: `block_and_report`

Platform selection query: Which supported runtime should execute this contract loop?

| Target | Instruction Files | Supports Subagents | Hooks | Generated Files |
| --- | --- | --- | --- | --- |
| codex | AGENTS.md | true | not_assumed | AGENTS.md |
| claude_code | CLAUDE.md | true | supported | CLAUDE.md, .claude/settings.json |
| cursor | .cursor/rules/looper-creator.mdc | true | supported_when_configured | .cursor/rules/looper-creator.mdc |
