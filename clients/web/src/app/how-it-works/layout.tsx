import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Enterprise Agent Supply Platform",
  description:
    "Build, validate, publish, install, operate, and improve governed Agents across your organization.",
  alternates: { canonical: "https://agentcloud.ai/product" },
};

export default function HowItWorksLayout({ children }: { children: React.ReactNode }) {
  return children;
}
