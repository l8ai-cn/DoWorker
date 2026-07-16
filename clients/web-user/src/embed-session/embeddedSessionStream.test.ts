import { expect, it, vi } from "vitest";

import type { EmbedSessionClient } from "@/embed-session-api";
import { consumeEmbeddedSessionStream } from "./embeddedSessionStream";

it("stops reconnecting when the embedded session token is rejected", async () => {
  const client = {
    openStream: vi.fn().mockResolvedValue(new Response(null, { status: 401 })),
  } as unknown as EmbedSessionClient;
  const controller = new AbortController();
  const onError = vi.fn();
  const onConnection = vi.fn();

  const result = await Promise.race([
    consumeEmbeddedSessionStream(client, controller, {
      onBlock: vi.fn(),
      onConnection,
      onError,
      onReconnect: vi.fn().mockResolvedValue(undefined),
      onStatus: vi.fn(),
    }).then(() => "stopped"),
    new Promise<string>((resolve) =>
      window.setTimeout(() => resolve("still-retrying"), 100),
    ),
  ]);

  controller.abort();
  expect(result).toBe("stopped");
  expect(client.openStream).toHaveBeenCalledOnce();
  expect(onError).toHaveBeenCalledOnce();
  expect(onConnection.mock.calls.map(([status]) => status)).toEqual([
    "connecting",
    "disconnected",
  ]);
});

it("rehydrates after the first stream attempt fails and the retry connects", async () => {
  vi.useFakeTimers();
  const controller = new AbortController();
  const liveStream = new ReadableStream<Uint8Array>({
    start(streamController) {
      controller.signal.addEventListener(
        "abort",
        () => streamController.close(),
        { once: true },
      );
    },
  });
  const client = {
    openStream: vi
      .fn()
      .mockResolvedValueOnce(new Response(null, { status: 500 }))
      .mockResolvedValueOnce(new Response(liveStream, { status: 200 })),
  } as unknown as EmbedSessionClient;
  const onReconnect = vi.fn(async () => {
    controller.abort();
  });
  const onConnection = vi.fn();

  const running = consumeEmbeddedSessionStream(client, controller, {
    onBlock: vi.fn(),
    onConnection,
    onError: vi.fn(),
    onReconnect,
    onStatus: vi.fn(),
  });
  await vi.advanceTimersByTimeAsync(500);
  await running;

  expect(client.openStream).toHaveBeenCalledTimes(2);
  expect(onReconnect).toHaveBeenCalledOnce();
  expect(onConnection.mock.calls.map(([status]) => status)).toEqual([
    "connecting",
    "reconnecting",
    "connected",
  ]);
  vi.useRealTimers();
});

it("reconnects when an open stream stops delivering bytes", async () => {
  vi.useFakeTimers();
  const controller = new AbortController();
  const streams = [
    (signal: AbortSignal) =>
      new ReadableStream<Uint8Array>({
        start(streamController) {
          signal.addEventListener(
            "abort",
            () => streamController.error(new DOMException("Aborted", "AbortError")),
            { once: true },
          );
        },
      }),
    (signal: AbortSignal) =>
      new ReadableStream<Uint8Array>({
        start(streamController) {
          signal.addEventListener(
            "abort",
            () => streamController.close(),
            { once: true },
          );
        },
      }),
  ];
  const client = {
    openStream: vi.fn(async (signal: AbortSignal) => {
      const stream = streams.shift();
      if (!stream) throw new Error("unexpected stream attempt");
      return new Response(stream(signal), { status: 200 });
    }),
  } as unknown as EmbedSessionClient;
  const onReconnect = vi.fn(async () => {
    controller.abort();
  });

  const running = consumeEmbeddedSessionStream(client, controller, {
    onBlock: vi.fn(),
    onConnection: vi.fn(),
    onError: vi.fn(),
    onReconnect,
    onStatus: vi.fn(),
  });
  await vi.advanceTimersByTimeAsync(36_000);

  expect(client.openStream).toHaveBeenCalledTimes(2);
  expect(onReconnect).toHaveBeenCalledOnce();
  await running;
  vi.useRealTimers();
});
