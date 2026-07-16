import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ExpertHome } from "../ExpertHome";

vi.mock("../ExpertHero", () => ({
  ExpertHero: () => <section data-testid="hero" />,
}));
vi.mock("../ExpertOperatingModel", () => ({
  ExpertOperatingModel: () => <section data-testid="operating" />,
}));
vi.mock("../SolutionDomains", () => ({
  SolutionDomains: () => <section data-testid="solutions" />,
}));
vi.mock("../CapabilitySpectrum", () => ({
  CapabilitySpectrum: () => <section data-testid="capabilities" />,
}));
vi.mock("../ExpertMarketplace", () => ({
  ExpertMarketplace: () => <section data-testid="marketplace" />,
}));
vi.mock("../ExpertGovernance", () => ({
  ExpertGovernance: () => <section data-testid="governance" />,
}));

describe("ExpertHome", () => {
  it("explains the supply lifecycle before presenting solution directions", () => {
    render(<ExpertHome />);

    expect(
      screen.getAllByTestId(
        /^(hero|solutions|operating|capabilities|marketplace|governance)$/,
      ).map((node) => node.dataset.testid),
    ).toEqual([
      "hero",
      "operating",
      "solutions",
      "capabilities",
      "marketplace",
      "governance",
    ]);
  });
});
