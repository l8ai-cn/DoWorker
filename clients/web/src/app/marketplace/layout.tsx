import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "专家应用市场",
  description: "浏览并启用已经装配好 Worker、提示词和 Skills 的 AI 专家应用。",
  alternates: {
    canonical: "https://agentsmesh.ai/marketplace",
  },
  openGraph: {
    title: "专家应用市场 | Do Worker",
    description: "选择可直接进入工作的 AI 专家应用。",
    url: "https://agentsmesh.ai/marketplace",
  },
};

export default function MarketplaceLayout({ children }: { children: React.ReactNode }) {
  return children;
}
