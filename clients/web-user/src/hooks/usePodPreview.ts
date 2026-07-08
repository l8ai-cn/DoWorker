import { useQuery, useQueryClient, type UseQueryResult } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { authenticatedFetch } from "@/lib/identity";
import { readDoWorkerOrgSlug } from "@/lib/do-worker";

/**
 * Wire response from `GET /api/v1/orgs/:slug/pods/:key/preview`
 * (backend/internal/api/rest/v1/pod_preview.go). Field names match the JSON
 * the server emits (snake_case), not a camelCased client convention.
 */
export interface PodPreviewInfo {
  preview_base_url: string;
  session_url: string;
  token: string;
  /** RFC3339 UTC timestamp. */
  expires_at: string;
}

export class PodPreviewError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = "PodPreviewError";
    this.status = status;
  }
}

function podPreviewQueryKey(podKey: string) {
  return ["pod-preview", podKey] as const;
}

async function fetchPodPreview(podKey: string, orgSlug: string): Promise<PodPreviewInfo> {
  const res = await authenticatedFetch(
    `/api/v1/orgs/${encodeURIComponent(orgSlug)}/pods/${encodeURIComponent(podKey)}/preview`,
  );
  if (!res.ok) {
    throw new PodPreviewError(res.status, `${res.status} ${res.statusText}`.trim());
  }
  return (await res.json()) as PodPreviewInfo;
}

/**
 * Builds the iframe `src` for a pod preview.
 *
 * This is deliberately the *session* URL, not the base preview URL with a raw
 * token query param: the browser navigates the iframe to `__session?token=`,
 * the gateway exchanges the one-shot token for an HttpOnly cookie and 302s to
 * the base, and the iframe's persisted `src` (what devtools/history show)
 * never carries the long-lived-looking raw token.
 */
export function buildPreviewSrc(info: PodPreviewInfo): string {
  return info.session_url;
}

/** How long before `expires_at` to proactively refetch a fresh token. */
const REFRESH_MARGIN_MS = 60_000;

export interface UsePodPreviewOptions {
  /** Set false to suspend fetching (e.g. panel not visible). Defaults to true. */
  enabled?: boolean;
}

/**
 * Fetches (and keeps fresh) the preview token + URLs for a pod. Refetches
 * shortly before the token's `expires_at` so a long-open preview panel never
 * gets stuck with a dead session after the iframe's initial load.
 */
export function usePodPreview(
  podKey: string,
  options: UsePodPreviewOptions = {},
): UseQueryResult<PodPreviewInfo, Error> {
  const orgSlug = readDoWorkerOrgSlug();
  const enabled = (options.enabled ?? true) && !!podKey && !!orgSlug;
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: podPreviewQueryKey(podKey),
    queryFn: () => fetchPodPreview(podKey, orgSlug as string),
    enabled,
    retry: false,
    // Freshness is governed by expires_at (via the refresh timer below), not
    // react-query's generic staleTime heuristics.
    staleTime: Infinity,
  });

  const timerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  useEffect(() => {
    clearTimeout(timerRef.current);
    const expiresAt = query.data?.expires_at ? Date.parse(query.data.expires_at) : NaN;
    if (Number.isNaN(expiresAt)) return;
    const delay = Math.max(expiresAt - Date.now() - REFRESH_MARGIN_MS, 0);
    timerRef.current = setTimeout(() => {
      queryClient.invalidateQueries({ queryKey: podPreviewQueryKey(podKey) });
    }, delay);
    return () => clearTimeout(timerRef.current);
  }, [query.data?.expires_at, podKey, queryClient]);

  return query;
}
