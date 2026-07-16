import { lightFetch } from "@/lib/light-auth";

export interface MarketplaceInstallResponse {
  expert: { slug: string };
  already_installed: boolean;
}

export function installMarketplaceApplication(
  orgSlug: string,
  applicationSlug: string,
  modelResourceID: number,
  toolModelResourceIDs: Record<string, number>,
): Promise<MarketplaceInstallResponse> {
  return lightFetch<MarketplaceInstallResponse>(
    `/api/v1/orgs/${orgSlug}/marketplace/experts/${applicationSlug}/install`,
    {
      method: "POST",
      authenticated: true,
      body: {
        model_resource_id: modelResourceID,
        tool_model_resource_ids: toolModelResourceIDs,
      },
    },
  );
}
