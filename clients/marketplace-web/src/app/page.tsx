import { CatalogPageContent } from "@/components/catalog-page-content";

export const dynamic = "force-dynamic";

export default function HomePage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  return <CatalogPageContent searchParams={searchParams} />;
}
