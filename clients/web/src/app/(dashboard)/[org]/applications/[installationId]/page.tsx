import { ApplicationsPage } from "@/components/applications/ApplicationsPage";

export default async function ApplicationFirstRunRoute({
  params,
}: {
  params: Promise<{ org: string; installationId: string }>;
}) {
  const { org, installationId } = await params;
  return <ApplicationsPage orgSlug={org} installationID={installationId} />;
}
