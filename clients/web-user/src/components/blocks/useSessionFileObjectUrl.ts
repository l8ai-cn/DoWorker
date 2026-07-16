import { useEffect, useState } from "react";
import { authenticatedFetch } from "@/lib/identity";

interface SessionFileObjectUrl {
  url: string | null;
  loading: boolean;
  error: boolean;
}

export function useSessionFileObjectUrl(path: string | null, enabled = true): SessionFileObjectUrl {
  const [state, setState] = useState<SessionFileObjectUrl>({
    url: null,
    loading: Boolean(path && enabled),
    error: false,
  });

  useEffect(() => {
    if (!path || !enabled) {
      setState({ url: null, loading: false, error: false });
      return;
    }

    let cancelled = false;
    let objectUrl: string | null = null;
    const controller = new AbortController();
    setState({ url: null, loading: true, error: false });

    void authenticatedFetch(path, { signal: controller.signal })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`file request failed: ${response.status}`);
        }
        return response.blob();
      })
      .then((blob) => {
        if (cancelled) return;
        objectUrl = URL.createObjectURL(blob);
        setState({ url: objectUrl, loading: false, error: false });
      })
      .catch(() => {
        if (!cancelled) {
          setState({ url: null, loading: false, error: true });
        }
      });

    return () => {
      cancelled = true;
      controller.abort();
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [enabled, path]);

  return state;
}
