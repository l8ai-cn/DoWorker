"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { listOrganizationEffectiveResources } from "@/lib/api";
import { readCurrentOrg } from "@/stores/auth";
import { VirtualKeysPanel } from "./VirtualKeysPanel";
import { TokenQuotaPanel } from "./TokenQuotaPanel";
import { QuotaReportPanel } from "./QuotaReportPanel";

export type TokenModelResource = {
  id: number;
  name: string;
  model: string;
};

export function ModelQuotasSettings() {
  const [models, setModels] = useState<TokenModelResource[]>([]);

  useEffect(() => {
    const orgSlug = readCurrentOrg()?.slug;
    if (!orgSlug) return;
    listOrganizationEffectiveResources(orgSlug)
      .then((resources) => {
        setModels(resources.flatMap((entry) => (
          entry.selectable && entry.resource
            ? [{ id: entry.resource.id, name: entry.resource.displayName, model: entry.resource.modelId }]
            : []
        )));
      })
      .catch((e) => toast.error(e instanceof Error ? e.message : "Failed to load model resources"));
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold">Quota &amp; Billing</h1>
        <p className="text-sm text-muted-foreground">
          Manage virtual API keys, token quotas, and view usage-vs-quota across the organization.
        </p>
      </div>
      <QuotaReportPanel />
      <VirtualKeysPanel models={models} />
      <TokenQuotaPanel />
    </div>
  );
}
