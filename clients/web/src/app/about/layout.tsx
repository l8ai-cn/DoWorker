import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "About",
  description:
    "Software is built by teams. Now teams are built differently. Agent Cloud helps organizations scale beyond headcount with AI agent teams.",
  alternates: {
    canonical: "https://agentcloud.ai/about",
  },
  openGraph: {
    title: "About | Agent Cloud",
    description:
      "Software is built by teams. Now teams are built differently. Agent Cloud helps organizations scale beyond headcount with AI agent teams.",
    url: "https://agentcloud.ai/about",
  },
};

export default function AboutLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
