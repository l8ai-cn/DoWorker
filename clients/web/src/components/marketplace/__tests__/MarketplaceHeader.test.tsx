import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { MarketplaceHeader } from "../MarketplaceHeader";

vi.mock("@/components/landing/Navbar", () => ({
  Navbar: () => <nav data-testid="marketing-navigation">Marketing navigation</nav>,
}));
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: vi.fn() }),
}));
vi.mock("@/hooks/useLightSession", () => ({
  useLightSession: () => ({ hydrated: false, session: null }),
}));
vi.mock("@/components/common", () => ({
  Logo: () => <span>Logo</span>,
}));

describe("MarketplaceHeader", () => {
  it("uses the shared five-page marketing navigation", () => {
    render(<MarketplaceHeader />);

    expect(screen.getByTestId("marketing-navigation")).toBeVisible();
  });
});
