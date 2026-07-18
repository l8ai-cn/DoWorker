import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import Home from "../page";

const { replace, useLightSession } = vi.hoisted(() => ({
  replace: vi.fn(),
  useLightSession: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
}));
vi.mock("@/hooks/useLightSession", () => ({
  useLightSession,
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
  beforeEach(() => {
    replace.mockReset();
    useLightSession.mockReturnValue({ hydrated: false, session: null });
  });

  it("server-renders the Agent supply overview without public pricing", () => {
    const { container } = render(<Home />);

    expect(screen.getByText("Expert overview")).toBeVisible();
    expect(screen.queryByText("Pricing section")).not.toBeInTheDocument();

    const script = container.querySelector('script[type="application/ld+json"]');
    const structuredData = JSON.parse(script?.textContent ?? "{}");
    expect(structuredData.description).toMatch(/higher-education digital employee pilot/i);
  });

  it("redirects an authenticated direct visit after session hydration", async () => {
    useLightSession.mockReturnValue({
      hydrated: true,
      session: {
        isAuthenticated: true,
        currentOrgSlug: "acme",
      },
    });

    render(<Home />);

    await waitFor(() => expect(replace).toHaveBeenCalledWith("/acme/workspace"));
  });
});
