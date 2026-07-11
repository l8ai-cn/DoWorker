import { createDocsMetadata } from "@/lib/docs-metadata";

export const metadata = createDocsMetadata("/docs/concepts/loop-and-workflow");

export default function Layout({ children }: { children: React.ReactNode }) {
  return children;
}
