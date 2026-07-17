"use client";

import { useEffect, useState } from "react";

import { fetchSessionByPodKey } from "@/lib/api/sessionImportApi";

interface AgentSessionLinkState {
  error: string | null;
  requestKey: string | null;
  sessionId: string | null;
}

const LINK_RETRY_DELAY_MS = 500;
const MAX_LINK_ATTEMPTS = 10;

export function useAgentSessionLink(podKey: string, enabled: boolean) {
  const [state, setState] = useState<AgentSessionLinkState>({
    error: null,
    requestKey: null,
    sessionId: null,
  });

  useEffect(() => {
    let cancelled = false;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;

    if (!enabled) {
      return;
    }

    const resolveLink = async (attempt: number) => {
      try {
        const session = await fetchSessionByPodKey(podKey);
        if (cancelled) return;
        if (session) {
          setState({
            error: null,
            requestKey: podKey,
            sessionId: session.id,
          });
          return;
        }
        if (attempt >= MAX_LINK_ATTEMPTS) {
          setState({
            error: "Agent session association was not created",
            requestKey: podKey,
            sessionId: null,
          });
          return;
        }
        retryTimer = setTimeout(
          () => void resolveLink(attempt + 1),
          LINK_RETRY_DELAY_MS,
        );
      } catch (error) {
        if (cancelled) return;
        setState({
          error: error instanceof Error ? error.message : "Session lookup failed",
          requestKey: podKey,
          sessionId: null,
        });
      }
    };

    void resolveLink(1);
    return () => {
      cancelled = true;
      if (retryTimer) clearTimeout(retryTimer);
    };
  }, [enabled, podKey]);

  if (!enabled) {
    return { error: null, loading: false, sessionId: null };
  }
  if (state.requestKey !== podKey) {
    return { error: null, loading: true, sessionId: null };
  }
  return {
    error: state.error,
    loading: false,
    sessionId: state.sessionId,
  };
}
