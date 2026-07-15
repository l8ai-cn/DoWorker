import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "How Do Worker Experts Work",
  description:
    "See how one Expert combines Workers, models, Skills, organizational knowledge, tools, and operating rules.",
  alternates: { canonical: "https://agentsmesh.ai/how-it-works" },
};

export default function HowItWorksLayout({ children }: { children: React.ReactNode }) {
  return children;
}
