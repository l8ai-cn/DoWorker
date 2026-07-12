import { ApplicationsPage } from "@/components/applications/ApplicationsPage";

export default async function ApplicationsRoute({
  params,
}: {
  params: Promise<{ org: string }>;
}) {
  const { org } = await params;
  return <ApplicationsPage orgSlug={org} />;
}
