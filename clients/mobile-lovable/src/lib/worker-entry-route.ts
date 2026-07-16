export type WorkerEntryDescriptor = {
  interactionMode: string;
  consoleAvailable: boolean;
};

export function resolveWorkerEntryRoute(
  descriptor: WorkerEntryDescriptor,
): "chat" | "terminal" | null {
  if (!descriptor.consoleAvailable) return null;
  if (descriptor.interactionMode === "acp") return "chat";
  if (descriptor.interactionMode === "pty") return "terminal";
  return null;
}
