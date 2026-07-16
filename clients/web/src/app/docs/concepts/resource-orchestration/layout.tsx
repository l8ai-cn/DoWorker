import { createDocsMetadata } from "@/lib/docs-metadata";

export const metadata = createDocsMetadata(
  "/docs/concepts/resource-orchestration",
);

export default function Layout({ children }: { children: React.ReactNode }) {
  return children;
}
