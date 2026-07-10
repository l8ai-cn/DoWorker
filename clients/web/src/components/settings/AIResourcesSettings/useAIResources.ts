"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
  createOrganizationConnection,
  createPersonalConnection,
  createResource,
  deleteConnection,
  deleteResource,
  getCatalog,
  listOrganizationConnections,
  listOrganizationEffectiveResources,
  listPersonalConnections,
  listPersonalEffectiveResources,
  rotateConnectionCredentials,
  setConnectionEnabled,
  setDefaultResource,
  setResourceEnabled,
  updateConnection,
  updateResource,
  validateConnection,
  type ConnectionInput,
  type ResourceInput,
} from "@/lib/api";
import type { AIResourcesData, AIResourceScope } from "./types";

const emptyData: AIResourcesData = { catalog: [], connections: [], effectiveResources: [] };

export function useAIResources(scope: AIResourceScope, organizationSlug?: string) {
  const [data, setData] = useState<AIResourcesData>(emptyData);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [operationFailed, setOperationFailed] = useState(false);
  const latestRequest = useRef(0);
  const scopeKey = `${scope}:${organizationSlug ?? ""}`;
  const currentScopeKey = useRef(scopeKey);
  currentScopeKey.current = scopeKey;

  const requireOrganizationSlug = useCallback(() => {
    if (!organizationSlug) throw new Error("Organization scope requires an organization slug");
    return organizationSlug;
  }, [organizationSlug]);

  const reload = useCallback(async () => {
    const requestScopeKey = scopeKey;
    const request = ++latestRequest.current;
    const current = () => request === latestRequest.current && requestScopeKey === currentScopeKey.current;
    if (!current()) return false;
    setLoading(true);
    setError(null);
    try {
      const scoped = scope === "personal"
        ? [listPersonalConnections(), listPersonalEffectiveResources()] as const
        : [listOrganizationConnections(requireOrganizationSlug()), listOrganizationEffectiveResources(requireOrganizationSlug())] as const;
      const [catalog, connections, effectiveResources] = await Promise.all([getCatalog(), ...scoped]);
      if (!current()) return true;
      setData({ catalog, connections, effectiveResources });
      setOperationFailed(false);
      return true;
    } catch (cause) {
      if (!current()) return true;
      setError(cause instanceof Error ? cause.message : "Failed to load AI resources");
      return false;
    } finally {
      if (current()) setLoading(false);
    }
  }, [requireOrganizationSlug, scope, scopeKey]);

  useEffect(() => {
    void reload();
  }, [reload]);

  const runOperation = useCallback(async (operation: () => Promise<unknown>) => {
    const operationScopeKey = scopeKey;
    const current = () => operationScopeKey === currentScopeKey.current;
    if (current()) setOperationFailed(false);
    try {
      await operation();
      if (!current()) return false;
      if (!await reload()) {
        if (current()) setOperationFailed(true);
        return false;
      }
      return true;
    } catch {
      if (current()) setOperationFailed(true);
      return false;
    }
  }, [reload, scopeKey]);

  const createConnection = useCallback(async (input: ConnectionInput) => {
    return runOperation(() => scope === "personal"
      ? createPersonalConnection(input)
      : createOrganizationConnection({ ...input, orgSlug: requireOrganizationSlug() }));
  }, [requireOrganizationSlug, runOperation, scope]);

  const createModelResource = useCallback(async (connectionId: number, input: ResourceInput) => {
    return runOperation(() => createResource(connectionId, input));
  }, [runOperation]);

  const updateProviderConnection = useCallback(async (connectionId: number, input: { name: string; baseUrl: string }) => {
    return runOperation(() => updateConnection(connectionId, input));
  }, [runOperation]);

  const rotateCredentials = useCallback(async (connectionId: number, credentials: Record<string, string>) => {
    return runOperation(() => rotateConnectionCredentials(connectionId, credentials));
  }, [runOperation]);

  const updateModelResource = useCallback(async (resourceId: number, input: Omit<ResourceInput, "identifier">) => {
    return runOperation(() => updateResource(resourceId, input));
  }, [runOperation]);

  const changeConnectionEnabled = useCallback(async (connectionId: number, enabled: boolean) => {
    return runOperation(() => setConnectionEnabled(connectionId, enabled));
  }, [runOperation]);

  const checkConnection = useCallback(async (connectionId: number) => {
    return runOperation(() => validateConnection(connectionId));
  }, [runOperation]);

  const changeResourceEnabled = useCallback(async (resourceId: number, enabled: boolean) => {
    return runOperation(() => setResourceEnabled(resourceId, enabled));
  }, [runOperation]);

  const makeDefault = useCallback(async (resourceId: number, modality: string) => {
    return runOperation(() => setDefaultResource(resourceId, modality));
  }, [runOperation]);

  const removeConnection = useCallback((connectionId: number) => {
    return runOperation(() => deleteConnection(connectionId));
  }, [runOperation]);

  const removeResource = useCallback((resourceId: number) => {
    return runOperation(() => deleteResource(resourceId));
  }, [runOperation]);

  return {
    ...data,
    loading,
    error,
    operationFailed,
    reload,
    createConnection,
    createModelResource,
    updateProviderConnection,
    rotateCredentials,
    updateModelResource,
    changeConnectionEnabled,
    checkConnection,
    changeResourceEnabled,
    makeDefault,
    removeConnection,
    removeResource,
  };
}
