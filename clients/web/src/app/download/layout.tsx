import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Download Agent Cloud Runner",
  description:
    "Download the Agent Cloud self-hosted Runner CLI for macOS, Windows, and Linux.",
  alternates: {
    canonical: "https://agentcloud.ai/download",
  },
  openGraph: {
    title: "Download Agent Cloud",
    description:
      "Download the self-hosted Runner that keeps your agents on your infrastructure.",
    url: "https://agentcloud.ai/download",
  },
};

export default function DownloadLayout({ children }: { children: React.ReactNode }) {
  return children;
}
