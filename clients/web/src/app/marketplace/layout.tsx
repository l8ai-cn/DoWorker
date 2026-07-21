import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Agent Market",
  description:
    "Discover approved, versioned Agents with visible capabilities, permissions, and delivery expectations.",
  alternates: {
    canonical: "https://market.l8ai.cn",
  },
  openGraph: {
    title: "Agent Market | Agent Cloud",
    description:
      "Discover approved Agents and inspect their capabilities, permissions, and versions before installation.",
    url: "https://market.l8ai.cn",
  },
};

export default function MarketplaceLayout({ children }: { children: React.ReactNode }) {
  return children;
}
