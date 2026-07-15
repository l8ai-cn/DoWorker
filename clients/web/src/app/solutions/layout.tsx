import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "AI Expert Solutions",
  description:
    "Explore outcome-first AI partners for cross-border growth, course development, and cross-functional collaboration.",
  alternates: { canonical: "https://agentsmesh.ai/solutions" },
};

export default function SolutionsLayout({ children }: { children: React.ReactNode }) {
  return children;
}
