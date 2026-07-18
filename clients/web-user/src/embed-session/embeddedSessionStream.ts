import type { EmbedSessionClient } from "@/embed-session-api";
import { BlockStream } from "@/lib/blockStream";
import type { AnyBlock } from "@/lib/blocks";
import type { StreamEvent } from "@/lib/events";
import { parseSseStream } from "@/lib/sse";
import type { SessionStatus } from "@/lib/types";

interface EmbeddedStreamCallbacks {
  onBlock(block: AnyBlock): void;
  onConnection(
    status: "connecting" | "connected" | "reconnecting" | "disconnected",
  ): void;
  onError(error: unknown): void;
  onReconnect(): Promise<void>;
  onStatus(status: SessionStatus): void;
}

const streamIdleTimeoutMs = 35_000;

export async function consumeEmbeddedSessionStream(
  client: EmbedSessionClient,
  controller: AbortController,
  callbacks: EmbeddedStreamCallbacks,
): Promise<void> {
  let connectedOnce = false;
  let needsHydration = false;
  while (!controller.signal.aborted) {
    const attemptController = new AbortController();
    const abortAttempt = () => attemptController.abort();
    controller.signal.addEventListener("abort", abortAttempt, { once: true });
    const reconnecting = connectedOnce || needsHydration;
    callbacks.onConnection(reconnecting ? "reconnecting" : "connecting");
    try {
      const response = await client.openStream(attemptController.signal);
      if (!response.ok) {
        throw new EmbeddedStreamOpenError(
          response.status,
          isRetryableStatus(response.status),
        );
      }
      if (!response.body) {
        throw new EmbeddedStreamOpenError(response.status, false);
      }
      connectedOnce = true;
      callbacks.onConnection("connected");
      if (needsHydration) await callbacks.onReconnect();
      needsHydration = false;
      const reducer = new BlockStream();
      for await (const block of reducer.reduce(
        trackSessionEvents(
          withStreamIdleTimeout(response.body, attemptController),
          callbacks.onStatus,
        ),
      )) {
        if (controller.signal.aborted) return;
        callbacks.onBlock(block);
      }
      needsHydration = true;
    } catch (cause) {
      if (controller.signal.aborted) return;
      callbacks.onError(cause);
      if (cause instanceof EmbeddedStreamOpenError && !cause.retryable) {
        callbacks.onConnection("disconnected");
        return;
      }
      needsHydration = true;
    } finally {
      controller.signal.removeEventListener("abort", abortAttempt);
      attemptController.abort();
    }
    await pause(500, controller.signal);
  }
}

class EmbeddedStreamOpenError extends Error {
  readonly retryable: boolean;

  constructor(status: number, retryable: boolean) {
    super(`Embedded session stream failed (${status})`);
    this.retryable = retryable;
  }
}

function isRetryableStatus(status: number): boolean {
  return status === 408 || status === 429 || status >= 500;
}

function withStreamIdleTimeout(
  body: ReadableStream<Uint8Array>,
  controller: AbortController,
): ReadableStream<Uint8Array> {
  const reader = body.getReader();
  return new ReadableStream<Uint8Array>({
    async pull(streamController) {
      try {
        const result = await readStreamChunk(reader, controller);
        if (result.done) {
          streamController.close();
          return;
        }
        streamController.enqueue(result.value);
      } catch (cause) {
        streamController.error(cause);
      }
    },
    cancel(reason) {
      controller.abort();
      return reader.cancel(reason);
    },
  });
}

function readStreamChunk(
  reader: ReadableStreamDefaultReader<Uint8Array>,
  controller: AbortController,
): Promise<ReadableStreamReadResult<Uint8Array>> {
  return new Promise((resolve, reject) => {
    const timer = window.setTimeout(() => {
      controller.abort();
      reject(new Error("Embedded session stream stalled"));
    }, streamIdleTimeoutMs);
    reader.read().then(
      (result) => {
        window.clearTimeout(timer);
        resolve(result);
      },
      (cause) => {
        window.clearTimeout(timer);
        reject(cause);
      },
    );
  });
}

async function* trackSessionEvents(
  body: ReadableStream<Uint8Array>,
  onStatus: (status: SessionStatus) => void,
): AsyncIterable<StreamEvent> {
  for await (const event of parseSseStream(body)) {
    if (event.type === "session_status") onStatus(event.status);
    yield event;
  }
}

function pause(ms: number, signal: AbortSignal): Promise<void> {
  if (signal.aborted) return Promise.resolve();
  return new Promise((resolve) => {
    const timer = window.setTimeout(resolve, ms);
    signal.addEventListener(
      "abort",
      () => {
        window.clearTimeout(timer);
        resolve();
      },
      { once: true },
    );
  });
}
