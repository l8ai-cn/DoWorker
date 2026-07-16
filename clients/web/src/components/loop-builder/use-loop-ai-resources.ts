"use client";

import { useEffect, useState } from "react";
import { listOrganizationEffectiveResources } from "@/lib/api/facade/aiResourceConnect";
import type { LoopAIResource } from "./loop-ai-assistant-types";

interface LoopAIResourceResult {
  requestKey: string;
  resources: LoopAIResource[];
  error?: string;
}

export function useLoopAIResources(orgSlug: string, resourceErrorMessage: string) {
  const [resourceAttempt, setResourceAttempt] = useState(0);
  const requestKey = `${orgSlug}:${resourceAttempt}`;
  const [result, setResult] = useState<LoopAIResourceResult>();

  useEffect(() => {
    let cancelled = false;
    listOrganizationEffectiveResources(orgSlug, ["chat"])
      .then((items) => {
        if (cancelled) return;
        setResult({
          requestKey,
          resources: items.flatMap((item) => {
            const resource = item.resource;
            const connection = item.connection;
            if (
              !item.selectable ||
              !resource ||
              !connection ||
              !resource.modalities.includes("chat") ||
              !resource.capabilities.includes("text-generation")
            ) {
              return [];
            }
            return [{
              id: String(resource.id),
              label: `${connection.name} · ${resource.displayName}`,
            }];
          }),
        });
      })
      .catch(() => {
        if (!cancelled) {
          setResult({ requestKey, resources: [], error: resourceErrorMessage });
        }
      });
    return () => {
      cancelled = true;
    };
  }, [orgSlug, requestKey, resourceErrorMessage]);

  const currentResult = result?.requestKey === requestKey ? result : undefined;

  return {
    resources: currentResult?.resources ?? [],
    resourcesLoading: !currentResult,
    resourceError: currentResult?.error,
    retryResources: () => setResourceAttempt((attempt) => attempt + 1),
  };
}
