import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { CodeExamples } from "./code-examples";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("Pod API code examples", () => {
  it("uses source lineage without runtime overrides", () => {
    render(<CodeExamples />);

    const example = screen.getByText(/curl -X POST/).textContent ?? "";
    expect(example).toContain(
      "/api/v1/ext/orgs/my-org/pods",
    );
    expect(example).toContain('X-API-Key: amk_your_api_key_here');
    expect(example).toContain('"source_pod_key"');
    expect(example).toContain('"resume_agent_session"');
    expect(example).not.toContain('"agent_slug"');
    expect(example).not.toContain('"agentfile_layer"');
    expect(example).not.toContain('"runner_id"');
  });
});
