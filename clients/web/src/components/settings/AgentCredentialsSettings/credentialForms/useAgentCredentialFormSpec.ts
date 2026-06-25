"use client";

import { useEffect, useMemo, useState } from "react";
import { getAgentConfigSchema } from "@/lib/api/facade/agentConnect";
import { useCurrentOrg } from "@/stores/auth";
import type { CredentialField } from "@/lib/viewModels/agent";
import type { CredentialFormSpec } from "./types";
import { buildCredentialFormSpec } from "./buildCredentialFormSpec";
import { getBuiltinCredentialFallback } from "./credentialBuiltinFallbacks";

export function useAgentCredentialFormSpec(agentSlug: string): {
  spec: CredentialFormSpec;
  loading: boolean;
} {
  const currentOrg = useCurrentOrg();
  const canFetchRemote = Boolean(currentOrg?.slug && agentSlug);
  const fetchKey = `${currentOrg?.slug ?? ""}:${agentSlug}`;

  const [remoteCache, setRemoteCache] = useState<{
    key: string;
    fields: CredentialField[];
  } | null>(null);

  useEffect(() => {
    if (!canFetchRemote) return;

    let cancelled = false;
    getAgentConfigSchema(currentOrg!.slug, agentSlug)
      .then((schema) => {
        if (!cancelled) {
          setRemoteCache({ key: fetchKey, fields: schema.credential_fields ?? [] });
        }
      })
      .catch(() => {
        if (!cancelled) {
          setRemoteCache({ key: fetchKey, fields: getBuiltinCredentialFallback(agentSlug) });
        }
      });

    return () => {
      cancelled = true;
    };
  }, [canFetchRemote, currentOrg, agentSlug, fetchKey]);

  const credentialFields = useMemo(() => {
    if (!canFetchRemote) return getBuiltinCredentialFallback(agentSlug);
    if (remoteCache?.key === fetchKey) return remoteCache.fields;
    return getBuiltinCredentialFallback(agentSlug);
  }, [canFetchRemote, agentSlug, fetchKey, remoteCache]);

  const spec = useMemo(
    () => buildCredentialFormSpec(agentSlug, credentialFields),
    [agentSlug, credentialFields],
  );

  const loading = canFetchRemote && remoteCache?.key !== fetchKey;

  return { spec, loading };
}
