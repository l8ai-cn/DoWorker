import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Enterprise Agent Supply Platform",
  description:
    "Build, validate, publish, install, operate, and improve governed Agents across your organization.",
  alternates: { canonical: "https://agentsmesh.ai/product" },
};

export default function ProductLayout({ children }: { children: React.ReactNode }) {
  return children;
}
