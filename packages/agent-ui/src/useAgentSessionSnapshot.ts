import { useEffect, useSyncExternalStore } from "react";

import type { AgentSessionRuntime } from "./contracts";

export function useAgentSessionSnapshot(
  runtime: AgentSessionRuntime,
  sessionId: string,
  lifecycleRuntime: AgentSessionRuntime = runtime,
) {
  useEffect(() => {
    void lifecycleRuntime.open(sessionId);
    return () => lifecycleRuntime.close(sessionId);
  }, [lifecycleRuntime, sessionId]);

  return useSyncExternalStore(
    (listener) => runtime.subscribe(sessionId, listener),
    () => runtime.getSnapshot(sessionId),
    () => runtime.getSnapshot(sessionId),
  );
}
