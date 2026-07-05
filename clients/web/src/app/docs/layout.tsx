import type { Metadata } from "next";
import DocsShell from "@/components/docs/DocsShell";

export const metadata: Metadata = {
  title: {
    template: "%s | Do Worker Docs",
    default: "Documentation",
  },
  description:
    "Do Worker documentation — orchestrate AI coding agents at scale.",
};

export default function DocsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <DocsShell>{children}</DocsShell>;
}
