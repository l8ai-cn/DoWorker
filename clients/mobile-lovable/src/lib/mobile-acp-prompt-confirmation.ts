type PendingPrompt = {
  reject: (reason: Error) => void;
  resolve: () => void;
  timer: number;
};

type RelayAcpEvent = Record<string, unknown>;

export function createMobileAcpPromptConfirmation(timeoutMs = 15_000) {
  const pending = new Map<string, PendingPrompt>();

  function waitFor(requestId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const timer = window.setTimeout(() => {
        pending.delete(requestId);
        reject(new Error("Worker 未确认接收消息，请重试"));
      }, timeoutMs);
      pending.set(requestId, { resolve, reject, timer });
    });
  }

  function consume(payload: unknown): boolean {
    if (!payload || typeof payload !== "object") return false;
    const event = payload as RelayAcpEvent;
    const requestId = typeof event.requestId === "string" ? event.requestId : "";
    const item = pending.get(requestId);
    if (!item) return false;
    if (event.type === "contentChunk" && event.role === "user") {
      finish(requestId, item.resolve);
      return true;
    }
    if (event.type === "commandFailed") {
      const message =
        typeof event.message === "string" && event.message
          ? event.message
          : "Worker 未能接收消息";
      finish(requestId, () => item.reject(new Error(message)));
      return true;
    }
    return false;
  }

  function rejectAll(message: string) {
    for (const [requestId, item] of pending) {
      finish(requestId, () => item.reject(new Error(message)));
    }
  }

  function finish(requestId: string, complete: () => void) {
    const item = pending.get(requestId);
    if (!item) return;
    pending.delete(requestId);
    window.clearTimeout(item.timer);
    complete();
  }

  return { consume, rejectAll, waitFor };
}
