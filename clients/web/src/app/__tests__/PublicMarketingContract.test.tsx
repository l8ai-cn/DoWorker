import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("next/font/google", () => ({
  Geist: () => ({ variable: "--font-geist-sans" }),
  Geist_Mono: () => ({ variable: "--font-geist-mono" }),
  Space_Grotesk: () => ({ variable: "--font-space-grotesk" }),
}));
vi.mock("geist/font/sans", () => ({
  GeistSans: { variable: "--font-geist-sans" },
}));
vi.mock("geist/font/mono", () => ({
  GeistMono: { variable: "--font-geist-mono" },
}));

import DownloadPage from "../download/page";
import EnterprisePage from "../enterprise/page";
import { FAQ_SECTIONS } from "../docs/faq/faq-sections";
import { metadata as rootMetadata, viewport } from "../layout";
import { metadata as marketplaceMetadata } from "../marketplace/layout";
import { metadata as solutionsMetadata } from "../solutions/layout";
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
      "/product",
      "/marketplace",
      "/docs",
    ]) {
      expect(urls).toContain(`https://agentsmesh.ai${path}`);
    }

    expect(urls).not.toContain("https://agentsmesh.ai/how-it-works");
    expect(urls).not.toContain("https://agentsmesh.ai/capabilities");
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

  it("keeps the Agent Market metadata aligned with its external destination", () => {
    expect(marketplaceMetadata).toMatchObject({
      title: "Agent Market",
      alternates: { canonical: "https://market.l8ai.cn" },
      openGraph: {
        title: "Agent Market | Do Worker",
        url: "https://market.l8ai.cn",
      },
    });
    expect(marketplaceMetadata.description).not.toMatch(/Worker|Expert|专家/i);
  });

  it("keeps higher-education positioning limited to pilots in public metadata", () => {
    expect(rootMetadata.description).toMatch(/higher-education digital employee pilots/i);
    expect(rootMetadata.keywords).toContain("higher-education digital employee pilots");
    expect(solutionsMetadata.description).toMatch(/higher-education digital employee pilots/i);
  });

  it("allows browser zoom for public pages", () => {
    expect(viewport).not.toMatchObject({
      maximumScale: 1,
      userScalable: false,
    });
  });
});
