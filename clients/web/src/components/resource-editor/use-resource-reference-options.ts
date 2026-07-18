"use client";

import { useEffect, useState } from "react";
import type { ResourceReferenceCatalog } from "./resource-reference-options";
import { loadResourceReferenceCatalog } from "./resource-reference-catalog-loader";

export function useResourceReferenceOptions(
  orgSlug: string,
  workerType: string = "",
  credentialTargetNames: readonly string[] = [],
): ResourceReferenceCatalog {
  const credentialTargetsJSON = JSON.stringify(
    [...new Set(credentialTargetNames)].sort(),
  );
  const identity = `${orgSlug}:${workerType}:${credentialTargetsJSON}`;
  const [state, setState] = useState<{
    identity: string;
    catalog: ResourceReferenceCatalog;
  }>({
    identity: "",
    catalog: emptyLoadingCatalog(),
  });

  useEffect(() => {
    let active = true;
    const credentialTargets = JSON.parse(credentialTargetsJSON) as string[];
    void loadResourceReferenceCatalog(
      orgSlug,
      workerType,
      credentialTargets,
    ).then((catalog) => {
      if (active) setState({ identity, catalog });
    });
    return () => {
      active = false;
    };
  }, [credentialTargetsJSON, identity, orgSlug, workerType]);

  return state.identity === identity ? state.catalog : emptyLoadingCatalog();
}

function emptyLoadingCatalog(): ResourceReferenceCatalog {
  return { loading: true, error: null, errorsByKind: {}, byKind: {} };
}
