import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Careers",
  description:
    "Join the Agent Cloud team and help shape the future of AI-powered software development.",
  alternates: {
    canonical: "https://agentcloud.ai/careers",
  },
  openGraph: {
    title: "Careers | Agent Cloud",
    description:
      "Join the Agent Cloud team and help shape the future of AI-powered software development.",
    url: "https://agentcloud.ai/careers",
  },
};

export default function CareersLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
