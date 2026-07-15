# Worker Runtime Evidence Audit

This document is a readable snapshot of the Worker integration contract. It
does not grant support status. The machine-readable sources are:

- `config/worker-types/catalog.json`
- `config/worker-types/<slug>/definition.json` and `AgentFile`
- `backend/internal/domain/workerruntime/runtime_catalog.lock.json`
- `tools/loops/worker-onboarding/catalog-loop/evidence/runtime-lock-probes.json`
- `tools/loops/worker-onboarding/catalog-loop/catalog/inventory.json`

## Support rule

A Worker is not supported merely because its command exists locally or because
a digest appears in a lock file. It needs all of the following:

1. A versioned Definition and matching AgentFile bundle hash.
2. An explicit `adapter_id`, executable, interaction modes, credential binding,
   and configuration-document contract.
3. A published immutable runtime reference whose exact digest can be pulled
   from the configured registry.
4. A non-mock runtime version probe and Runner create evidence using the
   declared adapter.
5. Backend, Rust Core, Web preflight, browser, ACP or PTY, and cleanup
   evidence.

`verified_local_dev` is local-development evidence, not a deployment claim. A
missing image, unpullable digest, missing authorized model resource, upstream
artifact, or failed build is an explicit blocker, never a fallback.

## Current release state

No Worker type is formally deployable at this time.

| Worker | Executable | Adapter | Credential source | Release runtime state | Local evidence |
| --- | --- | --- | --- | --- | --- |
| Aider | `aider` | `aider-pty` | Aider credential bundle | No published digest; upstream build is blocked | No Runner/product proof |
| Claude Code | `claude` | `claude-stream-json` | Anthropic model resource | Configured digest is not pullable | Runtime and guard-path proof only |
| Codex CLI | `codex` | `codex-app-server` | OpenAI-compatible model resource | Configured digest is not pullable | Local create, ACP prompt, and cleanup verified |
| Cursor CLI | `agent` | `cursor-acp` | Cursor credential bundle | No published digest | Local runtime and adapter unit proof only |
| Do Agent | `do-agent` | `do-agent-acp` | OpenAI-compatible or Anthropic model resource | No published digest | Local runtime and adapter unit proof only |
| Gemini CLI | `gemini` | `gemini-acp` | Gemini model resource | Configured digest is not pullable | Runtime and missing-model guard proof only |
| Grok Build | `grok` | `grok-build-acp` | Grok Build credential bundle | No published digest | Local runtime and adapter unit proof only |
| Hermes | `hermes` | `hermes-pty` | OpenAI-compatible model resource | No published digest; upstream build is blocked | No Runner/product proof |
| Loopal | `loopal` | `loopal-acp` | Loopal credential bundle | No published digest; no accepted real artifact | No Runner/product proof |
| MiniMax CLI | `mmx` | `minimax-pty` | MiniMax model resource | No published digest | Local runtime proof only |
| OpenClaw | `openclaw` | `openclaw-pty` | OpenAI-compatible model resource | No published digest | Local runtime proof only |
| OpenCode | `opencode` | `opencode-acp` | None | No published digest | Local runtime and adapter unit proof only |

The configured Codex CLI, Claude Code, and Gemini CLI references returned
`not found` during the exact-digest pull probe. Their lock entries are
therefore not pullable release artifacts. Codex's local browser/Runner path is
valuable evidence, but it cannot promote the type while the configured
published runtime is unavailable.

## Build contract

Each Worker definition belongs under `config/worker-types/<slug>/`:

```text
definition.json  identity, executable, adapter, modes, credentials, config documents
AgentFile        runtime workspace and process configuration
```

The catalog hash covers both files:

```text
sha256(definition.json + NUL + AgentFile)
```

The Runner must receive the exact `adapter_id` declared by the Definition.
Unknown adapters fail; the Runner must not infer ACP from an executable name.
Model credentials are injected from a selected model resource. Credential
bundles only contain the Definition-declared bundle fields.

## Runtime image template

Use a real upstream artifact and fail if it is absent. Do not substitute a mock
binary or register a mutable tag:

```dockerfile
RUN install-real-worker-cli --version "${WORKER_VERSION:?required}" \
  && worker-cli --version
```

```json
{
  "reference": "registry.example/do-worker/runner-worker@sha256:<64-lowercase-hex>",
  "digest": "sha256:<64-lowercase-hex>",
  "worker_type_slugs": ["worker"]
}
```

An image becomes selectable only after its build, exact-digest pull probe,
Runner path, and product-path evidence are recorded. A local `:latest` tag is
never an immutable runtime lock.

## Verification commands

```bash
node scripts/probe-local-worker-images.mjs
node scripts/probe-worker-runtime-locks.mjs
node scripts/generate-worker-loop-inventory.mjs
pnpm run worker-docs:sync

bash tools/loops/worker-onboarding/catalog-loop/scripts/verify-runtime-lock-probes.sh
bash tools/loops/worker-onboarding/catalog-loop/scripts/verify-inventory.sh
bash tools/loops/worker-onboarding/catalog-loop/scripts/verify-catalog-contract.sh
bash tools/loops/worker-onboarding/catalog-loop/scripts/verify-definition-chain.sh
```

After a Worker reaches a terminal result, use:

```bash
bash tools/loops/worker-onboarding/catalog-loop/scripts/verify-worker-run.sh <slug>
bash tools/loops/worker-onboarding/catalog-loop/scripts/verify-worker-runs.sh --processed
```

`--all-verified` is stricter: it fails whenever a Worker is blocked or lacks
the full end-to-end evidence chain.
