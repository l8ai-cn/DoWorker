import { useQuery, useQueryClient, type UseQueryResult } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { authenticatedFetch } from "@/lib/identity";
import { readDoWorkerOrgSlug } from "@/lib/do-worker";

export interface PodPreviewInfo {
  preview_base_url: string;
  session_url: string;
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

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object";
}

function toPodPreviewInfo(payload: unknown): PodPreviewInfo {
  if (!isRecord(payload)) {
    throw new Error("Invalid preview response");
  }

  const {
    preview_base_url,
    session_url,
    expires_at,
  } = payload as Record<string, unknown>;

  if (
    typeof preview_base_url !== "string" ||
    typeof session_url !== "string" ||
    typeof expires_at !== "string"
  ) {
    throw new Error("Invalid preview response");
  }

  return {
    preview_base_url,
    session_url,
    expires_at,
  };
}

function podPreviewQueryKey(podKey: string) {
  return ["pod-preview", podKey] as const;
}

export function buildPreviewSrc(info: PodPreviewInfo): string {
  return info.session_url;
}

/** How long before expiry to proactively refetch the session URL. */
const REFRESH_MARGIN_MS = 60_000;

export interface UsePodPreviewOptions {
  /** Set false to suspend fetching (e.g. panel not visible). Defaults to true. */
  enabled?: boolean;
}

async function fetchPodPreview(podKey: string, orgSlug: string): Promise<PodPreviewInfo> {
  const res = await authenticatedFetch(
    `/api/v1/orgs/${encodeURIComponent(orgSlug)}/pods/${encodeURIComponent(podKey)}/preview`,
  );
  if (!res.ok) {
    throw new PodPreviewError(res.status, `${res.status} ${res.statusText}`.trim());
  }
  return toPodPreviewInfo(await res.json());
}

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
