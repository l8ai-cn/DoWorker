export interface CredentialOneOfOverride {
  group: string;
  label: string;
  description?: string;
  envKeys: string[];
}

export interface CredentialUxOverride {
  fieldOrder?: string[];
  oneofGroups?: CredentialOneOfOverride[];
  labels?: Record<string, string>;
  descriptions?: Record<string, string>;
  placeholders?: Record<string, string>;
  securityHints?: Record<string, string>;
  allowCustomEnv?: boolean;
  customEnvHint?: string;
}

const OVERRIDES: Record<string, CredentialUxOverride> = {
  "claude-code": {
    fieldOrder: ["ANTHROPIC_BASE_URL", "ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN"],
    oneofGroups: [
      {
        group: "anthropic_auth",
        label: "settings.credentialForm.anthropic.authMethod",
        description: "settings.credentialForm.anthropic.authMethodHint",
        envKeys: ["ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN"],
      },
    ],
    labels: {
      ANTHROPIC_BASE_URL: "settings.credentialForm.anthropic.baseUrl",
      ANTHROPIC_API_KEY: "settings.credentialForm.anthropic.apiKey",
      ANTHROPIC_AUTH_TOKEN: "settings.credentialForm.anthropic.authToken",
    },
    placeholders: {
      ANTHROPIC_BASE_URL: "https://api.anthropic.com",
      ANTHROPIC_API_KEY: "sk-ant-...",
    },
    securityHints: {
      ANTHROPIC_BASE_URL: "settings.agentCredentials.baseUrlSecurityHint",
    },
    allowCustomEnv: false,
  },
  "codex-cli": {
    labels: { OPENAI_API_KEY: "settings.credentialForm.openai.apiKey" },
    placeholders: { OPENAI_API_KEY: "sk-..." },
    allowCustomEnv: true,
    customEnvHint: "settings.credentialForm.codex.customEnvHint",
  },
  loopal: {
    labels: {
      ANTHROPIC_API_KEY: "settings.credentialForm.anthropic.apiKey",
      OPENAI_API_KEY: "settings.credentialForm.openai.apiKey",
      GOOGLE_API_KEY: "settings.credentialForm.google.apiKey",
    },
    placeholders: {
      ANTHROPIC_API_KEY: "sk-ant-...",
      OPENAI_API_KEY: "sk-...",
    },
    allowCustomEnv: true,
    customEnvHint: "settings.credentialForm.loopal.customEnvHint",
  },
  "gemini-cli": {
    labels: { GOOGLE_API_KEY: "settings.credentialForm.google.apiKey" },
    allowCustomEnv: true,
    customEnvHint: "settings.credentialForm.gemini.customEnvHint",
  },
  aider: {
    labels: {
      OPENAI_API_KEY: "settings.credentialForm.openai.apiKey",
      ANTHROPIC_API_KEY: "settings.credentialForm.anthropic.apiKey",
    },
    allowCustomEnv: true,
    customEnvHint: "settings.credentialForm.aider.customEnvHint",
  },
  opencode: {
    allowCustomEnv: true,
    customEnvHint: "settings.credentialForm.opencode.customEnvHint",
  },
  "cursor-cli": {
    labels: { CURSOR_API_KEY: "settings.credentialForm.cursor.apiKey" },
    descriptions: { CURSOR_API_KEY: "settings.credentialForm.cursor.apiKeyHint" },
    allowCustomEnv: true,
    customEnvHint: "settings.credentialForm.cursor.customEnvHint",
  },
  "do-agent": {
    labels: {
      OPENAI_API_KEY: "settings.credentialForm.openai.apiKey",
      ANTHROPIC_API_KEY: "settings.credentialForm.anthropic.apiKey",
    },
    placeholders: {
      OPENAI_API_KEY: "sk-...",
      ANTHROPIC_API_KEY: "sk-ant-...",
    },
    allowCustomEnv: true,
    customEnvHint: "settings.credentialForm.doAgent.customEnvHint",
  },
  "grok-build": {
    labels: { XAI_API_KEY: "settings.credentialForm.xai.apiKey" },
    placeholders: { XAI_API_KEY: "xai-..." },
    allowCustomEnv: true,
    customEnvHint: "settings.credentialForm.grok.customEnvHint",
  },
  openclaw: {
    labels: {
      OPENAI_API_KEY: "settings.credentialForm.openai.apiKey",
      ANTHROPIC_API_KEY: "settings.credentialForm.anthropic.apiKey",
      XAI_API_KEY: "settings.credentialForm.xai.apiKey",
      GOOGLE_API_KEY: "settings.credentialForm.google.apiKey",
      GEMINI_API_KEY: "settings.credentialForm.google.geminiApiKey",
    },
    placeholders: {
      OPENAI_API_KEY: "sk-...",
      ANTHROPIC_API_KEY: "sk-ant-...",
      XAI_API_KEY: "xai-...",
      GEMINI_API_KEY: "AIza...",
    },
    allowCustomEnv: true,
    customEnvHint: "settings.credentialForm.openclaw.customEnvHint",
  },
  hermes: {
    labels: {
      OPENAI_API_KEY: "settings.credentialForm.openai.apiKey",
      ANTHROPIC_API_KEY: "settings.credentialForm.anthropic.apiKey",
      XAI_API_KEY: "settings.credentialForm.xai.apiKey",
      GOOGLE_API_KEY: "settings.credentialForm.google.apiKey",
      GEMINI_API_KEY: "settings.credentialForm.google.geminiApiKey",
    },
    placeholders: {
      OPENAI_API_KEY: "sk-...",
      ANTHROPIC_API_KEY: "sk-ant-...",
      XAI_API_KEY: "xai-...",
      GEMINI_API_KEY: "AIza...",
    },
    allowCustomEnv: true,
    customEnvHint: "settings.credentialForm.hermes.customEnvHint",
  },
  "e2e-echo": {
    labels: { E2E_TEST_CRED_KEY: "E2E Test Credential Key" },
    descriptions: {
      E2E_TEST_CRED_KEY: "Internal stub — used only by EnvBundle end-to-end tests.",
    },
    placeholders: { E2E_TEST_CRED_KEY: "Test value (not a real credential)" },
    allowCustomEnv: true,
    customEnvHint: "Internal stub agent — values are echoed by the test pod.",
  },
};

export function getCredentialUxOverride(agentSlug: string): CredentialUxOverride | undefined {
  return OVERRIDES[agentSlug];
}
