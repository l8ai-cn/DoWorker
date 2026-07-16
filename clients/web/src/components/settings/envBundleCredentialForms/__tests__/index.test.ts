import { describe, expect, it } from "vitest";
import {
  findFieldByEnvKey,
  getCredentialFormSpecFromFields,
  getEnvKeysFromSpec,
} from "../index";

describe("credential form specs", () => {
  it("orders Claude Code fields from the API schema", () => {
    const spec = getCredentialFormSpecFromFields("claude-code", [
      { name: "ANTHROPIC_API_KEY", type: "secret", optional: true },
      { name: "ANTHROPIC_BASE_URL", type: "text", optional: true },
    ]);

    expect(spec.fields).toMatchObject([
      { envKey: "ANTHROPIC_BASE_URL" },
      { envKey: "ANTHROPIC_API_KEY" },
    ]);
  });

  it("uses a Cursor field supplied by the API schema", () => {
    const spec = getCredentialFormSpecFromFields("cursor-cli", [
      { name: "CURSOR_API_KEY", type: "secret", optional: true },
    ]);

    expect(spec.agentSlug).toBe("cursor-cli");
    expect(getEnvKeysFromSpec(spec)).toEqual(new Set(["CURSOR_API_KEY"]));
  });

  it("uses the Gemini API key label declared by the Worker definition", () => {
    const spec = getCredentialFormSpecFromFields("gemini-cli", [
      { name: "GEMINI_API_KEY", type: "secret", optional: true },
    ]);

    expect(spec.fields).toMatchObject([
      {
        envKey: "GEMINI_API_KEY",
        label: "settings.credentialForm.google.geminiApiKey",
      },
    ]);
  });

  it("returns an empty form when the API declares no credential fields", () => {
    expect(getCredentialFormSpecFromFields("custom-worker", [])).toEqual({
      agentSlug: "custom-worker",
      fields: [],
    });
  });
});

describe("credential field lookup", () => {
  it("returns declared simple fields", () => {
    const spec = getCredentialFormSpecFromFields("claude-code", [
      { name: "ANTHROPIC_BASE_URL", type: "text", optional: true },
    ]);

    expect(findFieldByEnvKey(spec, "ANTHROPIC_BASE_URL")?.kind).toBe("text");
    expect(findFieldByEnvKey(spec, "UNKNOWN_KEY")).toBeUndefined();
  });
});
