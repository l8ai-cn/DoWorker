import type { Metadata } from "next";
import DocsShell from "@/components/docs/DocsShell";

export const metadata: Metadata = {
  title: {
    template: "%s | Agent Cloud Docs",
    default: "Documentation",
  },
  description:
    "Agent Cloud documentation — orchestrate AI coding agents at scale.",
};

export default function DocsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <DocsShell>{children}</DocsShell>;
}
