import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Request a Demo",
  description: "See how Agent Cloud can multiply your team's output with AI agent teams. Request a personalized demo today.",
  alternates: {
    canonical: "https://agentcloud.ai/demo",
  },
  openGraph: {
    title: "Request a Demo | Agent Cloud",
    description: "See how Agent Cloud can multiply your team's output with AI agent teams.",
    url: "https://agentcloud.ai/demo",
  },
};

export default function DemoLayout({ children }: { children: React.ReactNode }) {
  return children;
}
