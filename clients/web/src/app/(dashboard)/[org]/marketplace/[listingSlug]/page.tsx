import { redirect } from "next/navigation";

export default async function MarketplaceDetailRoute({
  params,
}: {
  params: Promise<{ org: string; listingSlug: string }>;
}) {
  const { listingSlug } = await params;
  redirect(`https://market.l8ai.cn/apps/${encodeURIComponent(listingSlug)}`);
}
