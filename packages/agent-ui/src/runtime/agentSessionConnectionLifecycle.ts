import { create } from "@bufbuild/protobuf";

import { SessionCursorSchema } from "@agent-cloud/proto/agent_workbench/v2/session_pb";

export interface InitialConnection {
  promise: Promise<void>;
  reject: (error: Error) => void;
  resolve: () => void;
  settled: boolean;
}

export function createInitialConnection(): InitialConnection {
  let resolve!: () => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<void>((accept, decline) => {
    resolve = accept;
    reject = decline;
  });
  return { promise, reject, resolve, settled: false };
}

export function resolveInitialConnection(
  initial: InitialConnection | undefined,
): void {
  if (!initial || initial.settled) return;
  initial.settled = true;
  initial.resolve();
}

export function rejectInitialConnection(
  initial: InitialConnection | undefined,
  error: Error,
): void {
  if (!initial || initial.settled) return;
  initial.settled = true;
  initial.reject(error);
}

export function sessionCursor(snapshot: {
  latestSequence: bigint;
  revision: bigint;
  sessionId: string;
  streamEpoch: string;
}) {
  return create(SessionCursorSchema, {
    sessionId: snapshot.sessionId,
    streamEpoch: snapshot.streamEpoch,
    revision: snapshot.revision,
    sequence: snapshot.latestSequence,
  });
}

export function cursorIdentity(snapshot: {
  latestSequence: bigint;
  revision: bigint;
  streamEpoch: string;
}): string {
  return [
    snapshot.streamEpoch,
    snapshot.revision.toString(),
    snapshot.latestSequence.toString(),
  ].join(":");
}

export function asConnectionError(cause: unknown): Error {
  return cause instanceof Error ? cause : new Error(String(cause));
}

export async function defaultAgentSessionRetryDelay(
  attempt: number,
  signal: AbortSignal,
): Promise<void> {
  const delay = Math.min(250 * 2 ** Math.max(0, attempt - 1), 4_000);
  await new Promise<void>((resolve) => {
    const timer = setTimeout(resolve, delay);
    signal.addEventListener(
      "abort",
      () => {
        clearTimeout(timer);
        resolve();
      },
      { once: true },
    );
  });
}
