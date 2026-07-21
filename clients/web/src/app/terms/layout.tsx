import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Terms of Service",
  description:
    "Agent Cloud terms of service — the agreement governing your use of our platform.",
  alternates: {
    canonical: "https://agentcloud.ai/terms",
  },
  openGraph: {
    title: "Terms of Service | Agent Cloud",
    description:
      "Agent Cloud terms of service — the agreement governing your use of our platform.",
    url: "https://agentcloud.ai/terms",
  },
};

export default function TermsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
