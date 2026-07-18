import { render, screen } from "@/test/test-utils";
import { describe, expect, it, vi } from "vitest";
import { ResourceYamlPanel } from "./ResourceYamlPanel";

describe("ResourceYamlPanel", () => {
  it("shows both source limits and the user manual", () => {
    render(
      <ResourceYamlPanel
        kind="Prompt"
        value="kind: Prompt"
        error={null}
        onChange={vi.fn()}
      />,
    );

    expect(screen.getByText(
      "Maximum 256 KiB per document and 64 KiB per physical line.",
    )).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "YAML reference" }))
      .toHaveAttribute(
        "href",
        "/docs/concepts/resource-orchestration#yaml",
      );
  });
});
