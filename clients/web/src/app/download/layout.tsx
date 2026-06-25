import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Download AgentsMesh Runner",
  description:
    "Download the AgentsMesh self-hosted Runner CLI for macOS, Windows, and Linux.",
  alternates: {
    canonical: "https://agentsmesh.ai/download",
  },
  openGraph: {
    title: "Download AgentsMesh",
    description:
      "Download the self-hosted Runner that keeps your agents on your infrastructure.",
    url: "https://agentsmesh.ai/download",
  },
};

export default function DownloadLayout({ children }: { children: React.ReactNode }) {
  return children;
}
