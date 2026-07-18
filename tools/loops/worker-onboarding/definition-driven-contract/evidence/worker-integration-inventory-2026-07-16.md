# Worker Integration Inventory

Source: `evidence/current-worker-evidence-matrix.json`.

- Formal Worker types: 13
- Runtime evidence only: 12
- Runtime blocked: 1
- Definition-owned config documents: 3
- Lifecycle passed: 1
- Lifecycle failed: 2
- Formal support: none; local lifecycle proof is not a full release claim.

| Worker | Runtime and transport | Model or credential source | Config document | Current release blockers |
| --- | --- | --- | --- | --- |
| aider | aider; pty; runtime evidence only | credential_bundle:aider->OPENAI_API_KEY (optional)<br>credential_bundle:aider->ANTHROPIC_API_KEY (optional) | none | not_verified<br>browser_not_verified<br>browser_not_verified<br>not_run |
| claude-code | claude; pty/acp; runtime evidence only | model_resource:anthropic->ANTHROPIC_API_KEY<br>model_resource:anthropic->ANTHROPIC_BASE_URL | none | not_verified<br>browser_not_verified<br>browser_not_verified<br>not_run |
| codex-cli | codex; pty/acp; runtime evidence only | model_resource:openai-compatible->OPENAI_API_KEY | none | not_verified<br>browser_not_verified<br>browser_not_verified<br>not_run |
| cursor-cli | agent; pty/acp; runtime evidence only | credential_bundle:cursor->CURSOR_API_KEY (optional) | none | not_verified<br>browser_flow_passed<br>browser_flow_passed<br>failed |
| do-agent | do-agent; pty/acp; runtime evidence only | model_resource:openai-compatible->OPENAI_API_KEY<br>model_resource:anthropic->ANTHROPIC_API_KEY | settings:json->DO_AGENT_SETTINGS | not_verified<br>public_contract_pending<br>browser_flow_passed<br>browser_flow_passed<br>failed |
| gemini-cli | gemini; pty/acp; runtime evidence only | model_resource:gemini->GEMINI_API_KEY | none | not_verified<br>browser_not_verified<br>browser_not_verified<br>not_run |
| grok-build | grok; pty/acp; runtime evidence only | credential_bundle:grok-build->XAI_API_KEY (required) | none | not_verified<br>browser_not_verified<br>browser_not_verified<br>not_run |
| hermes | hermes; pty; runtime evidence only | model_resource:openai-compatible->OPENAI_API_KEY | none | not_verified<br>browser_not_verified<br>browser_not_verified<br>not_run |
| loopal | loopal; pty/acp; No runtime image is available for this worker type | credential_bundle:loopal->ANTHROPIC_API_KEY (optional)<br>credential_bundle:loopal->OPENAI_API_KEY (optional)<br>credential_bundle:loopal->GOOGLE_API_KEY (optional) | none | not_verified<br>browser_not_verified<br>browser_not_verified<br>not_run |
| minimax-cli | mmx; pty; runtime evidence only | model_resource:minimax->MINIMAX_API_KEY | none | not_verified<br>browser_not_verified<br>browser_not_verified<br>not_run |
| openclaw | openclaw; pty; runtime evidence only | model_resource:openai-compatible->OPENAI_API_KEY | openclaw-json:json->openclaw-home/.openclaw/openclaw.json | not_verified<br>public_contract_pending<br>browser_not_verified<br>browser_not_verified<br>not_run |
| opencode | opencode; pty/acp; runtime evidence only | none | none | not_verified<br>browser_flow_passed<br>browser_flow_passed<br>passed |
| seedance-expert | do-agent; pty/acp; runtime evidence only | model_resource:openai-compatible->OPENAI_API_KEY<br>model_resource:anthropic->ANTHROPIC_API_KEY | settings:json->DO_AGENT_SETTINGS | not_verified<br>public_contract_pending<br>browser_not_verified<br>browser_not_verified<br>not_run |

## Gate Semantics

- `runtime evidence only`: a local image probe, selectable create option, and
  online Runner report exist. It is not a successful Worker lifecycle.
- `credential reference`: only proves the API projects a reference field;
  it does not prove key injection or provider authentication.
- `config document`: Do Agent, OpenClaw, and Seedance require an explicit
  `{document_id, config_bundle_id}` binding before materialization can be
  tested. Anonymous config bundles are insufficient.
- `lifecycle`: `failed` records an actual attempt that did not satisfy every
  release phase; it is not a support claim. `not_run` has no lifecycle evidence.
