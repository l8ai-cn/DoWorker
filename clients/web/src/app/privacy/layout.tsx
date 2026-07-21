import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Privacy Policy",
  description:
    "Agent Cloud privacy policy — how we collect, use, and protect your data.",
  alternates: {
    canonical: "https://agentcloud.ai/privacy",
  },
  openGraph: {
    title: "Privacy Policy | Agent Cloud",
    description:
      "Agent Cloud privacy policy — how we collect, use, and protect your data.",
    url: "https://agentcloud.ai/privacy",
  },
};

export default function PrivacyLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
