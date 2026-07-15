import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { render } from "@/test/test-utils";
import ResourceOrchestrationPage from "./page";

vi.mock("@/components/docs/DocNavigation", () => ({
  DocNavigation: () => <div data-testid="doc-navigation" />,
}));

describe("ResourceOrchestrationPage", () => {
  it("documents the resource lifecycle, YAML draft, and GoalLoop Apply boundary", () => {
    render(<ResourceOrchestrationPage />);

    expect(
      screen.getByRole("heading", { name: "Resource-native orchestration" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Validate → Plan → Apply")).toBeInTheDocument();
    expect(screen.getByText("One draft, two views")).toBeInTheDocument();
    expect(screen.getByText(/GoalLoop Apply creates a draft/)).toBeInTheDocument();
  });
});
