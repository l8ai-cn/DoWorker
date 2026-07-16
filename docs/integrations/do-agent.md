# Do Agent Worker Status

Do Agent is a formal Worker type with an explicit ACP adapter, but it is not
currently selectable because no immutable runtime image digest has been
published. This page describes the contract, not a completed integration.

| Area | Contract |
| --- | --- |
| Worker slug | `do-agent` |
| Executable | `do-agent` |
| Adapter | `do-agent-acp` |
| Interaction modes | PTY and ACP |
| Model resources | OpenAI-compatible or Anthropic |
| Credential injection | `OPENAI_API_KEY` or `ANTHROPIC_API_KEY` |
| Configuration document | JSON `settings` written through `DO_AGENT_SETTINGS` |
| Runtime lock | Missing |
| Browser and product-path evidence | Missing |

The authoritative Definition and AgentFile are:

```text
config/worker-types/do-agent/definition.json
config/worker-types/do-agent/AgentFile
```

The Worker creation API, Rust Core, and Web form read this contract. Model
credentials are selected through `model_resource_id`; they are not accepted as
plaintext form fields or arbitrary environment bundles.

## Completion Gates

Do Agent can be promoted only after all of the following are recorded:

1. A real Do Agent artifact builds into a Runner image and passes
   `do-agent --version`.
2. The image is published and added to the immutable runtime catalog.
3. Runner starts the exact `do-agent-acp` transport for ACP mode without
   executable-name inference.
4. The product path preflights a compatible model resource, creates a Worker,
   exchanges a real ACP prompt and response, and terminates cleanly.

Until then, the UI must keep Do Agent unavailable with the runtime-image reason.
