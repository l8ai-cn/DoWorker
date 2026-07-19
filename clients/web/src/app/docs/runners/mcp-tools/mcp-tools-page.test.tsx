import { screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { render } from "@/test/test-utils";
import MCPToolsPage from "./page";

vi.mock("@/components/docs/DocNavigation", () => ({
  DocNavigation: () => <div data-testid="doc-navigation" />,
}));

describe("MCPToolsPage", () => {
  it("documents resource-native Worker and Workflow creation", () => {
    render(<MCPToolsPage />);

    const podSection = screen.getByRole("heading", { name: "Pod Tools" })
      .closest("section");
    const workflowSection = screen.getByRole("heading", {
      name: "Workflow Tools",
    }).closest("section");

    expect(podSection).not.toBeNull();
    expect(workflowSection).not.toBeNull();
    expect(within(podSection!).getByText("plan_id (required)"))
      .toBeInTheDocument();
    expect(within(workflowSection!).getByText("create_workflow"))
      .toBeInTheDocument();
    expect(
      within(workflowSection!).getByText(
        "resource (Workflow, required), enabled?",
      ),
    ).toBeInTheDocument();
  });
});
