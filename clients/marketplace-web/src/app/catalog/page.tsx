import type { Metadata } from "next";

import { CatalogPageContent } from "@/components/catalog-page-content";

export const metadata: Metadata = {
  title: "全部市场内容",
};

export const dynamic = "force-dynamic";

export default function CatalogPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  return <CatalogPageContent searchParams={searchParams} catalogOnly />;
}
