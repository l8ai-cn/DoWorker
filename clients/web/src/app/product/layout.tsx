import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Enterprise Agent Supply Platform",
  description:
    "Build, verify, release, install, run, and evolve governed Agents across your organization.",
  alternates: { canonical: "https://agentsmesh.ai/product" },
};

export default function ProductLayout({ children }: { children: React.ReactNode }) {
  return children;
}
