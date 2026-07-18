# Agent Runtime Adapters

Canonical manifest: `loop.json`

Adapter outputs are generated artifacts. Update `loop.json` and regenerate.

Unsupported capability behavior: `block_and_report`

Platform selection query: Use Codex for this local evidence loop.

| Target | Instruction Files | Supports Subagents | Hooks | Generated Files |
| --- | --- | --- | --- | --- |
| codex | AGENTS.md | true | deterministic shell verifier | AGENTS.md |
