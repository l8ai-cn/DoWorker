import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ExpertHero } from "../ExpertHero";

const labels = vi.hoisted(() => ({
  "landing.workforce.expertHome.hero.eyebrow": "Enterprise Agent supply",
  "landing.workforce.expertHome.hero.title": "Make Agents reusable organizational capacity",
  "landing.workforce.expertHome.hero.description": "Build and operate governed Agents.",
  "landing.workforce.expertHome.hero.primaryAction": "Enter the Agent market",
  "landing.workforce.expertHome.hero.secondaryAction": "Explore the supply system",
  "landing.workforce.expertHome.hero.proof": "Human checkpoints",
  "landing.workforce.expertHome.console.title": "Expert control surface",
  "landing.workforce.expertHome.console.status": "Running",
  "landing.workforce.expertHome.console.live": "Live",
  "landing.workforce.expertHome.console.paused": "Paused",
  "landing.workforce.expertHome.console.complete": "Complete",
  "landing.workforce.expertHome.console.goalLabel": "Goal",
  "landing.workforce.expertHome.console.goal": "Launch a campaign",
  "landing.workforce.expertHome.console.expertLabel": "Active expert",
  "landing.workforce.expertHome.console.expert": "Growth expert",
  "landing.workforce.expertHome.console.formulaLabel": "Assembly",
  "landing.workforce.expertHome.console.formula": "Worker + Skills + Workflow",
  "landing.workforce.expertHome.console.workflowLabel": "Delivery chain",
  "landing.workforce.expertHome.console.checkpointLabel": "Human checkpoint",
  "landing.workforce.expertHome.console.checkpoint": "Approve positioning",
  "landing.workforce.expertHome.console.deliverableLabel": "Inspectable delivery",
  "landing.workforce.expertHome.console.deliverable": "Brief and evidence",
  "landing.workforce.expertHome.console.controls.pause": "Pause",
  "landing.workforce.expertHome.console.controls.resume": "Resume",
  "landing.workforce.expertHome.console.controls.next": "Next step",
  "landing.workforce.expertHome.console.controls.replay": "Replay",
}));

const rawLabels = vi.hoisted(() => ({
  "landing.workforce.expertHome.console.steps": [
    "Understand demand",
    "Match capabilities",
    "Confirm access",
    "Coordinate work",
    "Human review",
    "Deliver evidence",
  ],
}));

vi.mock("next-intl", () => ({
  useTranslations: () =>
    Object.assign(
      (key: keyof typeof labels) => labels[key] ?? key,
      { raw: (key: keyof typeof rawLabels) => rawLabels[key] },
    ),
}));

describe("Agent supply hero", () => {
  it("presents the supply proposition and a controllable delivery chain", () => {
    render(<ExpertHero />);

    expect(screen.getByRole("heading", { name: "Make Agents reusable organizational capacity" })).toBeVisible();
    expect(screen.getByRole("link", { name: "Enter the Agent market" })).toHaveAttribute(
      "href",
      "/marketplace",
    );
    expect(screen.getByRole("link", { name: "Explore the supply system" })).toHaveAttribute(
      "href",
      "/product",
    );
    expect(screen.getByRole("region", { name: "Expert control surface" })).toBeVisible();
    expect(screen.getAllByRole("listitem")).toHaveLength(6);
    expect(screen.getByRole("button", { name: "Pause" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "Next step" })).toBeEnabled();
  });

  it("reflects paused and completed states in the control surface", () => {
    render(<ExpertHero />);

    const nextButton = screen.getByRole("button", { name: "Next step" });
    fireEvent.click(screen.getByRole("button", { name: "Pause" }));

    expect(screen.getByText("Paused")).toBeVisible();
    expect(screen.getByRole("button", { name: "Resume" })).toBeVisible();
    expect(nextButton).toBeDisabled();

    fireEvent.click(screen.getByRole("button", { name: "Resume" }));
    fireEvent.click(nextButton);
    fireEvent.click(nextButton);
    fireEvent.click(nextButton);

    expect(screen.getByText("Complete")).toBeVisible();
    expect(nextButton).toBeDisabled();
  });
});
