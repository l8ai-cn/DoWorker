import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Careers",
  description:
    "Join the Do Worker team and help shape the future of AI-powered software development.",
  alternates: {
    canonical: "https://agentsmesh.ai/careers",
  },
  openGraph: {
    title: "Careers | Do Worker",
    description:
      "Join the Do Worker team and help shape the future of AI-powered software development.",
    url: "https://agentsmesh.ai/careers",
  },
};

export default function CareersLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
