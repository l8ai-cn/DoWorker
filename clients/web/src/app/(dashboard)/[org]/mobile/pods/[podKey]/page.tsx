import { permanentRedirect } from "next/navigation";

export default async function LegacyMobilePodPage({
  params,
}: {
  params: Promise<{ org: string; podKey: string }>;
}) {
  const { org, podKey } = await params;
  permanentRedirect(
    `/${encodeURIComponent(org)}/mobile/workers/${encodeURIComponent(podKey)}`,
  );
}
