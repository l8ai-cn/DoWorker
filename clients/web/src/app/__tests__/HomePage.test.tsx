import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import Home from "../page";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: vi.fn() }),
}));
vi.mock("@/hooks/useLightSession", () => ({
  useLightSession: () => ({ hydrated: true, session: null }),
}));
vi.mock("@/components/landing", () => ({
  Navbar: () => <div>Navigation</div>,
  PricingSection: () => <section>Pricing section</section>,
  FinalCTA: () => <section>Final action</section>,
  Footer: () => <footer>Footer</footer>,
}));
vi.mock("@/components/landing/expert-home/ExpertHome", () => ({
  ExpertHome: () => <section>Expert overview</section>,
}));

describe("Home", () => {
  it("keeps the Expert overview without public pricing", () => {
    render(<Home />);

    expect(screen.getByText("Expert overview")).toBeVisible();
    expect(screen.queryByText("Pricing section")).not.toBeInTheDocument();
  });
});
