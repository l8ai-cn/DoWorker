import { BUILTIN_CREDENTIAL_FALLBACK } from "./credentialBuiltinFallbacks";

BUILTIN_CREDENTIAL_FALLBACK["e2e-echo"] = [
  { name: "E2E_TEST_CRED_KEY", type: "secret", optional: true },
];
