# Worker Integration Evidence Execution

Status: active

## Completion Standard

A Worker is eligible for a local verification claim only after all applicable
gates have durable evidence:

1. Its formal Definition, configuration schema, credential bindings, runtime
   image contract, and database projection agree.
2. A matching Runner is online and its Relay tunnel is connected.
3. Authenticated options and preflight APIs return the expected result.
4. The Web flow is exercised in a real browser with console, network, and
   screenshot evidence.
5. A disposable Worker is created, reaches its declared PTY or ACP lifecycle,
   and is terminated and verified as cleaned up.

An unavailable image, Runner, resource, credential, browser, or lifecycle
gate is a blocked result, not a fallback or support claim.

## Current Runtime Facts

Collected July 16, 2026 from the local development database and standard
Backend, Relay, and Runner processes.

| Worker type | Online matching Runner | Options API |
| --- | --- | --- |
| `codex-cli` | `dev-runner-codex` | selectable |
| `do-agent` | `dev-runner-do-agent` | selectable |
| `seedance-expert` | `dev-runner-do-agent` | selectable |
| `gemini-cli` | `dev-runner-gemini` | selectable |
| `minimax-cli` | `dev-runner-minimax` | selectable |
| `openclaw` | `dev-runner-openclaw` | selectable |
| `aider`, `claude-code`, `cursor-cli`, `grok-build`, `hermes`, `loopal`, `opencode` | no verified current path | unavailable |

The formal catalog and database projection contain all 13 Worker types.
`ListWorkerCreateOptions` correctly requires both an enabled runtime image and
an online matching Runner before a type is selectable.

## Executed Verification

The following credential and transport lifecycle was run with a temporary
Runner and no external model call:

1. Registered through `http://[::1]:10000` using the seeded development
   registration token.
2. Confirmed the Backend issued the Runner certificate, private key, CA, and
   configuration.
3. Started the Runner and observed its mTLS gRPC stream plus Relay tunnel.
4. Deleted the temporary Runner through the authenticated API.
5. Confirmed cleanup with `GetRunner` returning `404` and a zero-row database
   lookup.

The standard Backend now fails fast when its gRPC server cannot start. The
development seed and startup checks were also repaired so a stale Runner token,
fixture Runner quota, or IPv4 loopback shadow cannot silently create misleading
evidence.

## Metadata Preflight

The authenticated Connect `PreflightWorker` path was run against all six
selectable types with a minimal PTY draft. It performed no Pod creation or
provider request.

| Worker type | Result |
| --- | --- |
| `codex-cli`, `do-agent`, `minimax-cli`, `openclaw` | resolved with zero issues |
| `gemini-cli` | blocked because no compatible Gemini protocol resource exists |
| `seedance-expert` | blocked because model resource `2` has the required video capability but is disabled |

Preflight had incorrectly called `ResolveExact`, which decrypts provider
credentials. It now calls `ResolveMetadata`; the latter retains authorization,
status, modality, capability, protocol, and endpoint checks without
decryption. Actual Pod launch continues to use `ResolveExact`.

`pnpm run web:typecheck` and `pnpm run web:test` passed. The latter ran 282
test files and 2,183 tests. `pnpm run web:lint` had no errors and 183 existing
warnings.

## Not Yet Verified

- No external provider credential was read, sent, or used.
- No Worker was created, prompted, or cleaned up in this rebuild.
- No browser path was run in this rebuild. Browser and Chrome DevTools both
  returned `Transport closed`. The July 16 retry could not create
  `/dev-org/workers/new`, so no page, console, network, or screenshot
  evidence exists.
- Runner tracing reports an OpenTelemetry schema-URL conflict. This did not
  block mTLS or Relay connectivity, but remains a separate defect.
- Five orphaned Pods from July 14-15 are still active in the development
  database. They predate this preflight run and were not modified.

## Next Verification

Restore the browser path, resolve Gemini and Seedance resource gates, classify
the orphaned Pods, then obtain explicit authorization before any provider-backed
create, prompt, or lifecycle test.
