import { useCallback, useEffect, useState } from "react";
import {
  listSessionTerminals,
  pickAgentTerminal,
  type TerminalInfo,
} from "@/lib/terminals-api";
import { readAuthToken } from "@/lib/auth-store";

export function useSessionTerminal(sessionId: string) {
  const [terminal, setTerminal] = useState<TerminalInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    if (!readAuthToken()) {
      setLoading(false);
      setError("未登录");
      return;
    }
    try {
      setError(null);
      const rows = await listSessionTerminals(sessionId);
      setTerminal(pickAgentTerminal(rows));
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载终端失败");
    } finally {
      setLoading(false);
    }
  }, [sessionId]);

  useEffect(() => {
    void refresh();
    const t = setInterval(() => void refresh(), 4000);
    return () => clearInterval(t);
  }, [refresh]);

  return { terminal, loading, error, refresh };
}
