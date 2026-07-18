import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Agent Supply Solutions",
  description:
    "Apply one governed Agent supply platform to enterprise internal supply, OPC incubation, and higher-education digital employee pilots.",
  alternates: { canonical: "https://agentsmesh.ai/solutions" },
};

export default function SolutionsLayout({ children }: { children: React.ReactNode }) {
  return children;
}
