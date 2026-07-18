# Worker Lifecycle Templates

Generated from `config/worker-types/*/definition.json` and the current Worker evidence matrix.

## Common Release Sequence

1. Build or locate the exact runtime image, then run its declared version probe.
2. Fetch create options, select only Definition-declared model and credential references, and preflight the draft.
3. In the browser, verify required fields, validation errors, and type-switch reset behavior without exposing a secret value.
4. With a named disposable target, create one Worker, prove the declared PTY or ACP connection, then run one harmless prompt.
5. Verify termination, Runner cleanup, browser state, and console/network errors before changing that Worker to supported.

## Per-Worker Templates

### aider

- Build: AGENT_RUNTIME=aider installs aider-chat with pip.
- Definition: executable `aider`; adapter `aider-pty`; modes `pty`.
- Form: No primary model resource is required. Credential bindings: `credential_bundle:aider->OPENAI_API_KEY, credential_bundle:aider->ANTHROPIC_API_KEY`. No Definition-owned config document. No tool model.
- Runner: PTY only; the Runner registers the aider process and usage parser.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: PTY attach, harmless prompt, terminal output, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### claude-code

- Build: AGENT_RUNTIME=claude-code installs @anthropic-ai/claude-code.
- Definition: executable `claude`; adapter `claude-stream-json`; modes `pty/acp`.
- Form: Primary model adapters: `anthropic`. Credential bindings: `model_resource:anthropic->ANTHROPIC_API_KEY, model_resource:anthropic->ANTHROPIC_BASE_URL`. No Definition-owned config document. No tool model.
- Runner: Custom claude-stream-json transport with Claude streaming and control handling.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: ACP initialize, session creation, harmless prompt, expected event, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### codex-cli

- Build: AGENT_RUNTIME=codex-cli installs @openai/codex.
- Definition: executable `codex`; adapter `codex-app-server`; modes `pty/acp`.
- Form: Primary model adapters: `openai-compatible`. Credential bindings: `model_resource:openai-compatible->OPENAI_API_KEY`. No Definition-owned config document. No tool model.
- Runner: Custom codex-app-server transport; isolated CODEX_HOME and auth/config materialization.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: ACP initialize, session creation, harmless prompt, expected event, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### cursor-cli

- Build: AGENT_RUNTIME=cursor-cli installs Cursor and exposes the agent binary.
- Definition: executable `agent`; adapter `cursor-acp`; modes `pty/acp`.
- Form: No primary model resource is required. Credential bindings: `credential_bundle:cursor->CURSOR_API_KEY`. No Definition-owned config document. No tool model.
- Runner: Standard ACP transport registered as cursor-acp.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: ACP initialize, session creation, harmless prompt, expected event, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### do-agent

- Build: AGENT_RUNTIME=do-agent stages the real do-agent sidecar binary.
- Definition: executable `do-agent`; adapter `do-agent-acp`; modes `pty/acp`.
- Form: Primary model adapters: `openai-compatible, anthropic`. Credential bindings: `model_resource:openai-compatible->OPENAI_API_KEY, model_resource:anthropic->ANTHROPIC_API_KEY`. Named config document bindings required: `settings:json->DO_AGENT_SETTINGS`. No tool model.
- Runner: Custom do-agent-acp transport with allow/restricted permission modes.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: ACP initialize, session creation, harmless prompt, expected event, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### gemini-cli

- Build: AGENT_RUNTIME=gemini-cli installs @google/gemini-cli.
- Definition: executable `gemini`; adapter `gemini-acp`; modes `pty/acp`.
- Form: Primary model adapters: `gemini`. Credential bindings: `model_resource:gemini->GEMINI_API_KEY`. No Definition-owned config document. No tool model.
- Runner: Standard ACP transport registered as gemini-acp; model launch argument is required.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: ACP initialize, session creation, harmless prompt, expected event, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### grok-build

- Build: AGENT_RUNTIME=grok-build installs @xai-official/grok.
- Definition: executable `grok`; adapter `grok-build-acp`; modes `pty/acp`.
- Form: No primary model resource is required. Credential bindings: `credential_bundle:grok-build->XAI_API_KEY`. No Definition-owned config document. No tool model.
- Runner: ACP transport performs xai.api_key headless authentication during initialize.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: ACP initialize, session creation, harmless prompt, expected event, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### hermes

- Build: AGENT_RUNTIME=hermes installs hermes-agent on the Python runtime base.
- Definition: executable `hermes`; adapter `hermes-pty`; modes `pty`.
- Form: Primary model adapters: `openai-compatible`. Credential bindings: `model_resource:openai-compatible->OPENAI_API_KEY`. No Definition-owned config document. No tool model.
- Runner: PTY only; HERMES_HOME is isolated per pod.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: PTY attach, harmless prompt, terminal output, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### loopal

- Build: AGENT_RUNTIME=loopal requires a real LOOPAL_BINARY and rejects E2E mock artifacts.
- Definition: executable `loopal`; adapter `loopal-acp`; modes `pty/acp`.
- Form: No primary model resource is required. Credential bindings: `credential_bundle:loopal->ANTHROPIC_API_KEY, credential_bundle:loopal->OPENAI_API_KEY, credential_bundle:loopal->GOOGLE_API_KEY`. No Definition-owned config document. No tool model.
- Runner: ACP transport forwards Loopal control-panel events and control requests.
- Current state: blocked: No runtime image is available for this worker type.
- Lifecycle proof: ACP initialize, session creation, harmless prompt, expected event, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### minimax-cli

- Build: AGENT_RUNTIME=minimax-cli installs mmx-cli behind the MMX config wrapper.
- Definition: executable `mmx`; adapter `minimax-pty`; modes `pty`.
- Form: Primary model adapters: `minimax`. Credential bindings: `model_resource:minimax->MINIMAX_API_KEY`. No Definition-owned config document. No tool model.
- Runner: PTY only; the wrapper writes MINIMAX_API_KEY to MMX_CONFIG_DIR.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: PTY attach, harmless prompt, terminal output, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### openclaw

- Build: AGENT_RUNTIME=openclaw installs OpenClaw with the pinned Node runtime.
- Definition: executable `openclaw`; adapter `openclaw-pty`; modes `pty`.
- Form: Primary model adapters: `openai-compatible`. Credential bindings: `model_resource:openai-compatible->OPENAI_API_KEY`. Named config document bindings required: `openclaw-json:json->openclaw-home/.openclaw/openclaw.json`. No tool model.
- Runner: PTY only; OPENCLAW_HOME config is merged and receives OpenAI provider settings.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: PTY attach, harmless prompt, terminal output, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### opencode

- Build: AGENT_RUNTIME=opencode installs opencode-ai.
- Definition: executable `opencode`; adapter `opencode-acp`; modes `pty/acp`.
- Form: No primary model resource is required. No credential binding. No Definition-owned config document. No tool model.
- Runner: Standard ACP transport registered as opencode-acp.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: ACP initialize, session creation, harmless prompt, expected event, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.

### seedance-expert

- Build: Uses AGENT_RUNTIME=do-agent and the real do-agent sidecar binary.
- Definition: executable `do-agent`; adapter `do-agent-acp`; modes `pty/acp`.
- Form: Primary model adapters: `openai-compatible, anthropic`. Credential bindings: `model_resource:openai-compatible->OPENAI_API_KEY, model_resource:anthropic->ANTHROPIC_API_KEY`. Named config document bindings required: `settings:json->DO_AGENT_SETTINGS`. Tool model: `seedance-video:doubao/video/video-generation`.
- Runner: Uses do-agent-acp plus a required Seedance video tool-model environment.
- Current state: runtime evidence only; no successful lifecycle is recorded.
- Lifecycle proof: ACP initialize, session creation, harmless prompt, expected event, terminate, and cleanup.
- Negative proof: reject undeclared secrets, incompatible model resources, malformed declared JSON, and stale resource revisions before dispatch.
