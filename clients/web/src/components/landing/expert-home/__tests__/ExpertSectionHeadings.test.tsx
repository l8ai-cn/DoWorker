import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { CapabilitySpectrum } from "../CapabilitySpectrum";
import { ExpertOperatingModel } from "../ExpertOperatingModel";

const labels = vi.hoisted(() => ({
  "landing.workforce.expertHome.operating.title": "How an Expert operates",
  "landing.workforce.expertHome.operating.formulaTitle": "Expert formula",
  "landing.workforce.expertHome.operating.humanTitle": "Human checkpoints",
  "landing.workforce.expertHome.operating.humanDescription": "Review critical decisions",
  "landing.workforce.expertHome.capabilities.title": "Capability spectrum",
  "landing.workforce.expertHome.capabilities.levels.implemented": "Implemented",
  "landing.workforce.expertHome.capabilities.levels.composable": "Composable",
  "landing.workforce.expertHome.capabilities.levels.planned": "Planned",
}));

const parts = vi.hoisted(() =>
  Array.from({ length: 6 }, (_, index) => ({
    id: `part-${index}`,
    title: `Part ${index}`,
    description: `Part description ${index}`,
  })),
);
const humanItems = vi.hoisted(() => [
  { id: "review", title: "Review", description: "Review decisions" },
]);
const capabilities = vi.hoisted(() =>
  Array.from({ length: 12 }, (_, index) => ({
    id: `capability-${index}`,
    title: `Capability ${index}`,
    description: `Capability description ${index}`,
    level: "implemented",
  })),
);

vi.mock("next-intl", () => ({
  useTranslations: () =>
    Object.assign(
      (key: keyof typeof labels) => labels[key] ?? key,
      {
        raw: (key: string) => {
          if (key.endsWith("operating.parts")) return parts;
          if (key.endsWith("operating.humanItems")) return humanItems;
          return capabilities;
        },
      },
    ),
}));

describe("expert page section headings", () => {
  it("keeps the operating model below the page h1", () => {
    render(<ExpertOperatingModel showIntro={false} />);

    expect(
      screen.getByRole("heading", { level: 2, name: "How an Expert operates" }),
    ).toBeInTheDocument();
  });

  it("keeps the capability spectrum below the page h1", () => {
    render(<CapabilitySpectrum showIntro={false} />);

    expect(
      screen.getByRole("heading", { level: 2, name: "Capability spectrum" }),
    ).toBeInTheDocument();
  });
});
