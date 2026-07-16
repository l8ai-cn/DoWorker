import { useCallback, useEffect, useRef, useState } from "react";
import { authenticatedFetch } from "@/lib/identity";

type SessionFileState =
  | { status: "idle"; url: null; mimeType: null }
  | { status: "loading"; url: null; mimeType: null }
  | { status: "ready"; url: string; mimeType: string | null }
  | { status: "error"; url: null; mimeType: null };

type SessionFileObjectUrl = SessionFileState & {
  load: () => void;
};

const IDLE_STATE: SessionFileState = {
  status: "idle",
  url: null,
  mimeType: null,
};

export function useSessionFileObjectUrl(path: string | null): SessionFileObjectUrl {
  const [state, setState] = useState<SessionFileState>(IDLE_STATE);
  const controllerRef = useRef<AbortController | null>(null);
  const objectUrlRef = useRef<string | null>(null);

  const clearActiveRequest = useCallback(() => {
    controllerRef.current?.abort();
    controllerRef.current = null;
    if (objectUrlRef.current) {
      URL.revokeObjectURL(objectUrlRef.current);
      objectUrlRef.current = null;
    }
  }, []);

  useEffect(() => {
    clearActiveRequest();
    setState(IDLE_STATE);
    return clearActiveRequest;
  }, [clearActiveRequest, path]);

  const load = useCallback(() => {
    if (!path) return;
    clearActiveRequest();
    const controller = new AbortController();
    controllerRef.current = controller;
    setState({ status: "loading", url: null, mimeType: null });

    void authenticatedFetch(path, { signal: controller.signal })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`file request failed: ${response.status}`);
        }
        const contentType = response.headers.get("Content-Type");
        return response.blob().then((blob) => ({ blob, contentType }));
      })
      .then(({ blob, contentType }) => {
        if (controller.signal.aborted) return;
        const url = URL.createObjectURL(blob);
        objectUrlRef.current = url;
        setState({
          status: "ready",
          url,
          mimeType: contentType || blob.type || null,
        });
      })
      .catch(() => {
        if (!controller.signal.aborted) {
          setState({ status: "error", url: null, mimeType: null });
        }
      });
  }, [clearActiveRequest, path]);

  return { ...state, load };
}
