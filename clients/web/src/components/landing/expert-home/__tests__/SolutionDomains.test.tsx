import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { SolutionDomains } from "../SolutionDomains";

const labels = vi.hoisted(() => ({
  "landing.workforce.expertHome.solutions.eyebrow": "Solution domains",
  "landing.workforce.expertHome.solutions.title": "Four entrances",
  "landing.workforce.expertHome.solutions.description": "One operating system",
  "landing.workforce.expertHome.solutions.workflowLabel": "Delivery path",
  "landing.workforce.expertHome.solutions.deliverablesLabel": "Inspectable result",
}));

const items = vi.hoisted(() => [
  { id: "cross-border-commerce", title: "Commerce", description: "Commerce work", chain: "Research to review", outcome: "Assets", action: "Explore commerce" },
  { id: "ai-education", title: "AI education", description: "Education work", chain: "Goals to assessment", outcome: "Lessons", action: "Explore education" },
  { id: "digital-employees", title: "AI partners", description: "Shared-context collaborators", chain: "Goal to evidence", outcome: "Inspectable results", action: "Create a partner" },
  { id: "marketplace", title: "Marketplace", description: "Verified apps", chain: "Discover to govern", outcome: "Installed apps", action: "Open marketplace" },
]);

vi.mock("next-intl", () => ({
  useTranslations: () =>
    Object.assign(
      (key: keyof typeof labels) => labels[key] ?? key,
      { raw: () => items },
    ),
}));

describe("SolutionDomains", () => {
  it("keeps a section heading when the page hero owns the visible introduction", () => {
    render(<SolutionDomains showIntro={false} />);

    expect(
      screen.getByRole("heading", { level: 2, name: "Four entrances" }),
    ).toBeInTheDocument();
  });

  it("switches between the four approved business entrances", () => {
    render(<SolutionDomains />);

    expect(screen.getAllByRole("tab")).toHaveLength(4);
    fireEvent.click(screen.getByRole("tab", { name: "AI education" }));

    expect(screen.getByRole("tab", { name: "AI education" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    const panel = screen.getByRole("tabpanel");
    expect(within(panel).getByText("Delivery path")).toBeVisible();
    expect(within(panel).getByText("Inspectable result")).toBeVisible();
    expect(within(panel).getByText("Goals to assessment")).toBeVisible();
    expect(within(panel).getByText("Lessons")).toBeVisible();
  });
});
