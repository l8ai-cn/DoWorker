import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";

vi.mock("./ResourceEditorShell", () => ({
  ResourceEditorShell: ({ kind }: { kind: string }) => (
    <div data-testid="selected-resource-kind">{kind}</div>
  ),
}));

import { ResourceDependencyEditor } from "./ResourceDependencyEditor";

describe("ResourceDependencyEditor", () => {
  it("lets users choose every reference resource kind", async () => {
    const user = userEvent.setup();
    render(<ResourceDependencyEditor orgSlug="acme" />);

    expect(screen.getByTestId("selected-resource-kind")).toHaveTextContent(
      "Prompt",
    );
    await user.selectOptions(
      screen.getByLabelText("Resource kind"),
      "ToolBinding",
    );
    expect(screen.getByTestId("selected-resource-kind")).toHaveTextContent(
      "ToolBinding",
    );
    expect(screen.getAllByRole("option").map((option) => option.textContent))
      .toEqual([
        "Prompt",
        "ModelBinding",
        "ToolBinding",
        "Repository",
        "Skill",
        "KnowledgeBase",
        "EnvironmentBundle",
        "ComputeTarget",
        "ResourceProfile",
      ]);
  });
});
