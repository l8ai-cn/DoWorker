import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Changelog",
  description:
    "Stay up to date with the latest Agent Cloud features, improvements, and bug fixes.",
  alternates: {
    canonical: "https://agentcloud.ai/changelog",
  },
  openGraph: {
    title: "Changelog | Agent Cloud",
    description:
      "Stay up to date with the latest Agent Cloud features, improvements, and bug fixes.",
    url: "https://agentcloud.ai/changelog",
  },
};

export default function ChangelogLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
