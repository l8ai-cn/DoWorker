import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Skill Marketplace",
  description:
    "Browse public Do Worker Skills that can be installed into repositories and mounted into AI workers.",
  alternates: {
    canonical: "https://agentsmesh.ai/marketplace",
  },
  openGraph: {
    title: "Skill Marketplace | Do Worker",
    description:
      "Browse public Do Worker Skills for Codex, Claude Code, DoAgent, and other AI workers.",
    url: "https://agentsmesh.ai/marketplace",
  },
};

export default function MarketplaceLayout({ children }: { children: React.ReactNode }) {
  return children;
}
