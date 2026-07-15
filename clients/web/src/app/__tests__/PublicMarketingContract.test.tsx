import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import DownloadPage from "../download/page";
import EnterprisePage from "../enterprise/page";
import { FAQ_SECTIONS } from "../docs/faq/faq-sections";
import sitemap from "../sitemap";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));
vi.mock("@/components/common", () => ({
  PageHeader: () => <header>Header</header>,
  PageFooter: () => <footer>Footer</footer>,
}));
vi.mock("@/components/landing", () => ({
  EnterpriseFeatures: () => <section>Enterprise capabilities</section>,
  SelfHostedCTA: () => <section>Deployment consultation</section>,
  Navbar: () => <nav>Navigation</nav>,
  Footer: () => <footer>Footer</footer>,
}));
vi.mock("@/components/download", () => ({
  DownloadHero: () => <section>Download</section>,
  FallbackHero: () => <section>Download unavailable</section>,
  RunnerSection: () => <section>Runner</section>,
  ResourcesSection: () => <section>Resources</section>,
}));
vi.mock("@/lib/download/github-release", () => ({
  fetchLatestRelease: vi.fn().mockResolvedValue(null),
}));
vi.mock("@/lib/blog", () => ({
  getAllPosts: vi.fn().mockResolvedValue([]),
}));

describe("public marketing contract", () => {
  it("publishes every primary marketing destination in the sitemap", async () => {
    const urls = new Set((await sitemap()).map(({ url }) => url));

    for (const path of [
      "/solutions",
      "/how-it-works",
      "/capabilities",
      "/marketplace",
      "/docs",
    ]) {
      expect(urls).toContain(`https://agentsmesh.ai${path}`);
    }
  });

  it("keeps the enterprise page focused on capabilities without pricing copy", () => {
    render(<EnterprisePage />);

    expect(screen.getByText("Enterprise capabilities")).toBeVisible();
    expect(screen.getByText("Deployment consultation")).toBeVisible();
    expect(screen.queryByText(/pricing/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/landing\.pricing/i)).not.toBeInTheDocument();
  });

  it("keeps public FAQ navigation free of billing and pricing sections", () => {
    expect(
      FAQ_SECTIONS.some(({ categoryKey }) => categoryKey.includes("billing")),
    ).toBe(false);
  });

  it("does not publish a price through download page structured data", async () => {
    const page = await DownloadPage();
    const { container } = render(page);
    const script = container.querySelector('script[type="application/ld+json"]');
    const structuredData = JSON.parse(script?.textContent ?? "{}");

    expect(structuredData).not.toHaveProperty("offers");
  });
});
