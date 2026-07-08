import { useCallback, useEffect, useState } from "react";
import { readAuthToken } from "@/lib/auth-store";
import { listLiveExperts, type LiveExpert } from "@/lib/experts-api";

export function useLiveExperts() {
  const [items, setItems] = useState<LiveExpert[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    if (!readAuthToken()) {
      setItems([]);
      setLoading(false);
      return;
    }
    try {
      setError(null);
      setItems(await listLiveExperts());
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载专家失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { items, loading, error, refresh, isLive: Boolean(readAuthToken()) };
}
