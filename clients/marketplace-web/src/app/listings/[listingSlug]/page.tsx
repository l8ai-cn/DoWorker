import { permanentRedirect } from "next/navigation";

type ListingParams = Promise<{ listingSlug: string }>;

export default async function LegacyListingPage({
  params,
}: {
  params: ListingParams;
}) {
  permanentRedirect(`/apps/${encodeURIComponent((await params).listingSlug)}`);
}
