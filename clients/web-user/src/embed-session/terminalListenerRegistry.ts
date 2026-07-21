import type { AgentConnectionStatus } from "@agent-cloud/agent-ui";

type OutputListener = (bytes: Uint8Array) => void;
type StatusListener = (status: AgentConnectionStatus) => void;

export class TerminalListenerRegistry {
  private readonly outputs = new Map<string, Set<OutputListener>>();
  private readonly statuses = new Map<string, Set<StatusListener>>();

  subscribeOutput(resourceId: string, listener: OutputListener): () => void {
    return this.add(this.outputs, resourceId, listener);
  }

  subscribeStatus(resourceId: string, listener: StatusListener): () => void {
    return this.add(this.statuses, resourceId, listener);
  }

  publishOutput(resourceId: string, bytes: Uint8Array): void {
    this.outputs.get(resourceId)?.forEach((listener) => listener(bytes));
  }

  publishStatus(resourceId: string, status: AgentConnectionStatus): void {
    this.statuses.get(resourceId)?.forEach((listener) => listener(status));
  }

  private add<T>(
    target: Map<string, Set<T>>,
    resourceId: string,
    listener: T,
  ): () => void {
    const listeners = target.get(resourceId) ?? new Set<T>();
    listeners.add(listener);
    target.set(resourceId, listeners);
    return () => listeners.delete(listener);
  }
}
