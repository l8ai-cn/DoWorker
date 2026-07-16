import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { SolutionDomains } from "../SolutionDomains";

const labels = vi.hoisted(() => ({
  "landing.workforce.expertHome.solutions.eyebrow": "Solution domains",
  "landing.workforce.expertHome.solutions.title": "Three organization models",
  "landing.workforce.expertHome.solutions.description": "One operating system",
  "landing.workforce.expertHome.solutions.workflowLabel": "Delivery path",
  "landing.workforce.expertHome.solutions.deliverablesLabel": "Inspectable result",
}));

const items = vi.hoisted(() => [
  { id: "enterprise-agent-supply", title: "Enterprise supply", description: "Governed internal supply", chain: "Build to improve", outcome: "Reusable Agents", action: "Explore enterprise supply" },
  { id: "opc-incubation", title: "OPC incubation", description: "One human with an AI operating team", chain: "Goal to delivery", outcome: "An operating loop", action: "Explore OPC incubation" },
  { id: "higher-education-digital-employees", title: "University digital employees", description: "Teaching and operations", chain: "Design to govern", outcome: "Reusable digital employees", action: "Explore universities" },
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
      screen.getByRole("heading", { level: 2, name: "Three organization models" }),
    ).toBeInTheDocument();
  });

  it("switches between the three approved solution directions", () => {
    render(<SolutionDomains />);

    expect(screen.getAllByRole("tab")).toHaveLength(3);
    fireEvent.click(screen.getByRole("tab", { name: "University digital employees" }));

    expect(screen.getByRole("tab", { name: "University digital employees" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    const panel = screen.getByRole("tabpanel");
    expect(within(panel).getByText("Delivery path")).toBeVisible();
    expect(within(panel).getByText("Inspectable result")).toBeVisible();
    expect(within(panel).getByText("Design to govern")).toBeVisible();
    expect(within(panel).getByText("Reusable digital employees")).toBeVisible();
  });
});
