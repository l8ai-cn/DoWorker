import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { CredentialFormFields } from "../CredentialFormFields";
import { getCredentialFormSpecFromFields } from "../envBundleCredentialForms";

const mockT = (key: string) => {
  const translations: Record<string, string> = {
    "common.optional": "Optional",
    "settings.agentCredentials.secretPlaceholder": "Leave empty to keep existing",
    "settings.agentCredentials.secretEditHint": "Leave empty to keep the existing value",
    "settings.credentialForm.anthropic.apiKey": "Anthropic API Key",
    "settings.credentialForm.anthropic.baseUrl": "Anthropic Base URL",
    "settings.credentialForm.openai.apiKey": "OpenAI API Key",
    "settings.credentialForm.google.apiKey": "Google API Key",
  };
  return translations[key] ?? key;
};

describe("CredentialFormFields", () => {
  it("renders the Claude fields returned by the API", () => {
    const spec = getCredentialFormSpecFromFields("claude-code", [
      { name: "ANTHROPIC_API_KEY", type: "secret", optional: true },
      { name: "ANTHROPIC_BASE_URL", type: "text", optional: true },
    ]);

    render(
      <CredentialFormFields
        spec={spec}
        values={{}}
        onValueChange={vi.fn()}
        selectedOneOf={{}}
        onOneOfChange={vi.fn()}
        isEditing={false}
        t={mockT}
      />,
    );

    expect(screen.getByLabelText("Anthropic Base URL")).toBeInTheDocument();
    expect(screen.getByLabelText("Anthropic API Key")).toBeInTheDocument();
  });

  it("renders only API-declared Loopal credential fields", () => {
    const spec = getCredentialFormSpecFromFields("loopal", [
      { name: "ANTHROPIC_API_KEY", type: "secret", optional: true },
      { name: "OPENAI_API_KEY", type: "secret", optional: true },
      { name: "GOOGLE_API_KEY", type: "secret", optional: true },
    ]);
    render(
      <CredentialFormFields
        spec={spec}
        values={{}}
        onValueChange={vi.fn()}
        selectedOneOf={{}}
        onOneOfChange={vi.fn()}
        isEditing={false}
        t={mockT}
      />,
    );

    expect(screen.getByLabelText("Anthropic API Key")).toBeInTheDocument();
    expect(screen.getByLabelText("OpenAI API Key")).toBeInTheDocument();
    expect(screen.getByLabelText("Google API Key")).toBeInTheDocument();
  });
});
