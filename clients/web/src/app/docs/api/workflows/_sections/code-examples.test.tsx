import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { CodeExamples } from "./code-examples";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("Workflow API code examples", () => {
  it("uses the external API route and API key authentication", () => {
    render(<CodeExamples />);

    const example = screen.getByText(/curl -X POST/).textContent ?? "";
    expect(example).toContain(
      "/api/v1/ext/orgs/{org}/workflows/{slug}/trigger",
    );
    expect(example).toContain('X-API-Key: {api_key}');
    expect(example).not.toContain("Authorization: Bearer");
  });
});
