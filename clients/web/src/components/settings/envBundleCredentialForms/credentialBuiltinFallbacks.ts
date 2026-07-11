import type { CredentialField } from "@/lib/viewModels/agent";

// Offline/test fallback when GetAgentConfigSchema is unavailable. Field sets
// mirror builtin AgentFile ENV SECRET/TEXT declarations; prefer API at runtime.
export const BUILTIN_CREDENTIAL_FALLBACK: Record<string, CredentialField[]> = {
  "claude-code": [
    { name: "ANTHROPIC_API_KEY", type: "secret", optional: true },
    { name: "ANTHROPIC_AUTH_TOKEN", type: "secret", optional: true },
    { name: "ANTHROPIC_BASE_URL", type: "text", optional: true },
  ],
  "codex-cli": [{ name: "OPENAI_API_KEY", type: "secret", optional: true }],
  loopal: [
    { name: "ANTHROPIC_API_KEY", type: "secret", optional: true },
    { name: "OPENAI_API_KEY", type: "secret", optional: true },
    { name: "GOOGLE_API_KEY", type: "secret", optional: true },
  ],
  "gemini-cli": [{ name: "GOOGLE_API_KEY", type: "secret", optional: true }],
  aider: [
    { name: "OPENAI_API_KEY", type: "secret", optional: true },
    { name: "ANTHROPIC_API_KEY", type: "secret", optional: true },
  ],
  "cursor-cli": [{ name: "CURSOR_API_KEY", type: "secret", optional: true }],
  "do-agent": [
    { name: "OPENAI_API_KEY", type: "secret", optional: true },
    { name: "ANTHROPIC_API_KEY", type: "secret", optional: true },
  ],
  "grok-build": [{ name: "XAI_API_KEY", type: "secret", optional: false }],
  openclaw: [
    { name: "OPENAI_API_KEY", type: "secret", optional: true },
    { name: "ANTHROPIC_API_KEY", type: "secret", optional: true },
    { name: "XAI_API_KEY", type: "secret", optional: true },
    { name: "GOOGLE_API_KEY", type: "secret", optional: true },
    { name: "GEMINI_API_KEY", type: "secret", optional: true },
  ],
  harn: [
    { name: "OPENAI_API_KEY", type: "secret", optional: true },
    { name: "ANTHROPIC_API_KEY", type: "secret", optional: true },
    { name: "XAI_API_KEY", type: "secret", optional: true },
    { name: "GOOGLE_API_KEY", type: "secret", optional: true },
    { name: "GEMINI_API_KEY", type: "secret", optional: true },
  ],
};

export function getBuiltinCredentialFallback(agentSlug: string): CredentialField[] {
  return BUILTIN_CREDENTIAL_FALLBACK[agentSlug] ?? [];
}
