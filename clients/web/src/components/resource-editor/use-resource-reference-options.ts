"use client";

import { useEffect, useState } from "react";
import type { ResourceReferenceCatalog } from "./resource-reference-options";
import { loadResourceReferenceCatalog } from "./resource-reference-catalog-loader";

export function useResourceReferenceOptions(
  orgSlug: string,
  workerType: string = "",
  modelProtocolAdapters: readonly string[] = [],
  credentialTargetNames: readonly string[] = [],
): ResourceReferenceCatalog {
  const modelAdaptersJSON = JSON.stringify(
    [...new Set(modelProtocolAdapters)].sort(),
  );
  const credentialTargetsJSON = JSON.stringify(
    [...new Set(credentialTargetNames)].sort(),
  );
  const identity = `${orgSlug}:${workerType}:${modelAdaptersJSON}:${credentialTargetsJSON}`;
  const [state, setState] = useState<{
    identity: string;
    catalog: ResourceReferenceCatalog;
  }>({
    identity: "",
    catalog: emptyLoadingCatalog(),
  });

  useEffect(() => {
    let active = true;
    const modelAdapters = JSON.parse(modelAdaptersJSON) as string[];
    const credentialTargets = JSON.parse(credentialTargetsJSON) as string[];
    void loadResourceReferenceCatalog(
      orgSlug,
      workerType,
      modelAdapters,
      credentialTargets,
    ).then((catalog) => {
      if (active) setState({ identity, catalog });
    });
    return () => {
      active = false;
    };
  }, [
    credentialTargetsJSON,
    identity,
    modelAdaptersJSON,
    orgSlug,
    workerType,
  ]);

  return state.identity === identity ? state.catalog : emptyLoadingCatalog();
}

function emptyLoadingCatalog(): ResourceReferenceCatalog {
  return { loading: true, error: null, errorsByKind: {}, byKind: {} };
}
