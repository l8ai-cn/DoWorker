"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { type ModelConfig, listModelConfigs } from "@/lib/api/quotaApi";
import { VirtualKeysPanel } from "./VirtualKeysPanel";
import { TokenQuotaPanel } from "./TokenQuotaPanel";
import { QuotaReportPanel } from "./QuotaReportPanel";

export function ModelQuotasSettings() {
  const [models, setModels] = useState<ModelConfig[]>([]);

  useEffect(() => {
    listModelConfigs()
      .then(setModels)
      .catch((e) => toast.error(e instanceof Error ? e.message : "Failed to load models"));
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
