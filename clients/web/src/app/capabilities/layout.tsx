import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "AI Expert Platform Capabilities",
  description:
    "Review implemented, composable, and planned Expert capabilities together with execution and permission controls.",
  alternates: { canonical: "https://agentsmesh.ai/capabilities" },
};

export default function CapabilitiesLayout({ children }: { children: React.ReactNode }) {
  return children;
}
