import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, vi } from "vitest";

import { TerminalSurface } from "./TerminalSurface";
import type { TerminalRuntime } from "./contracts";

const terminalMock = vi.hoisted(() => ({
  onData: null as null | ((data: string) => void),
  write: vi.fn(),
  dispose: vi.fn(),
  fit: vi.fn(),
}));

vi.mock("@xterm/xterm", () => ({
  Terminal: class {
    cols = 120;
    rows = 36;
    loadAddon() {}
    open() {}
    write = terminalMock.write;
    dispose = terminalMock.dispose;
    onData(listener: (data: string) => void) {
      terminalMock.onData = listener;
      return { dispose: vi.fn() };
    }
  },
}));

vi.mock("@xterm/addon-fit", () => ({
  FitAddon: class {
    fit = terminalMock.fit;
  },
}));

class ResizeObserverMock {
  static callback: ResizeObserverCallback | null = null;
  constructor(callback: ResizeObserverCallback) {
    ResizeObserverMock.callback = callback;
  }
  observe() {}
  disconnect() {}
}

vi.stubGlobal("ResizeObserver", ResizeObserverMock);

beforeEach(() => {
  vi.clearAllMocks();
  terminalMock.onData = null;
  ResizeObserverMock.callback = null;
});

it("keeps one terminal connection while control changes and writes controlled input", async () => {
  let output: ((bytes: Uint8Array) => void) | null = null;
  const runtime: TerminalRuntime = {
    connect: vi.fn(async () => undefined),
    disconnect: vi.fn(),
    subscribeOutput: vi.fn((_id, listener) => {
      output = listener;
      return vi.fn();
    }),
    subscribeStatus: vi.fn(() => vi.fn()),
    write: vi.fn(async () => undefined),
    resize: vi.fn(async () => undefined),
    acquireControl: vi.fn(async () => ({
      leaseId: "lease-1",
      expiresAt: Date.now() + 60_000,
    })),
    renewControl: vi.fn(async () => undefined),
    releaseControl: vi.fn(async () => undefined),
  };

  const { unmount } = render(
    <TerminalSurface
      clientLabel="iframe"
      resource={{
        id: "terminal-1",
        label: "Agent terminal",
        status: "connected",
        writable: true,
      }}
      runtime={runtime}
    />,
  );

  await waitFor(() => expect(runtime.connect).toHaveBeenCalledTimes(1));
  (output as ((bytes: Uint8Array) => void) | null)?.(
    new TextEncoder().encode("ready"),
  );
  expect(terminalMock.write).toHaveBeenCalled();
  ResizeObserverMock.callback?.([], {} as ResizeObserver);
  expect(runtime.resize).not.toHaveBeenCalled();

  fireEvent.click(screen.getByRole("button", { name: "Take control" }));
  await waitFor(() => expect(runtime.acquireControl).toHaveBeenCalledTimes(1));
  await waitFor(() =>
    expect(
      screen.getByRole("button", { name: "Release control" }),
    ).toBeVisible(),
  );
  expect(runtime.resize).toHaveBeenCalledWith("terminal-1", 120, 36);

  terminalMock.onData?.("ls\r");
  await waitFor(() =>
    expect(runtime.write).toHaveBeenCalledWith(
      "terminal-1",
      new TextEncoder().encode("ls\r"),
    ),
  );
  expect(runtime.connect).toHaveBeenCalledTimes(1);

  unmount();
  await waitFor(() =>
    expect(runtime.releaseControl).toHaveBeenCalledWith(
      "terminal-1",
      "lease-1",
    ),
  );
  await waitFor(() =>
    expect(runtime.disconnect).toHaveBeenCalledWith("terminal-1"),
  );
  expect(
    vi.mocked(runtime.releaseControl).mock.invocationCallOrder[0],
  ).toBeLessThan(vi.mocked(runtime.disconnect).mock.invocationCallOrder[0]);
});

it("shows terminal connection errors instead of leaving a blank surface", async () => {
  const runtime = {
    connect: vi.fn(async () => {
      throw new Error("Relay unavailable");
    }),
    disconnect: vi.fn(),
    subscribeOutput: vi.fn(() => vi.fn()),
    subscribeStatus: vi.fn(() => vi.fn()),
    write: vi.fn(async () => undefined),
    resize: vi.fn(async () => undefined),
    acquireControl: vi.fn(),
    renewControl: vi.fn(),
    releaseControl: vi.fn(),
  } satisfies TerminalRuntime;

  render(
    <TerminalSurface
      clientLabel="iframe"
      resource={{
        id: "terminal-1",
        label: "Agent terminal",
        status: "connecting",
        writable: false,
      }}
      runtime={runtime}
    />,
  );

  expect(await screen.findByText("Relay unavailable")).toBeVisible();
});

it("tears down the terminal when releasing control throws synchronously", async () => {
  const runtime = {
    connect: vi.fn(async () => undefined),
    disconnect: vi.fn(),
    subscribeOutput: vi.fn(() => vi.fn()),
    subscribeStatus: vi.fn(() => vi.fn()),
    write: vi.fn(async () => undefined),
    resize: vi.fn(async () => undefined),
    acquireControl: vi.fn(async () => ({
      leaseId: "lease-1",
      expiresAt: Date.now() + 60_000,
    })),
    renewControl: vi.fn(async () => undefined),
    releaseControl: vi.fn(() => {
      throw new Error("release failed");
    }),
  } as unknown as TerminalRuntime;

  const { unmount } = render(
    <TerminalSurface
      clientLabel="iframe"
      resource={{
        id: "terminal-1",
        label: "Agent terminal",
        status: "connected",
        writable: true,
      }}
      runtime={runtime}
    />,
  );
  await waitFor(() => expect(runtime.connect).toHaveBeenCalledTimes(1));
  fireEvent.click(screen.getByRole("button", { name: "Take control" }));
  await screen.findByRole("button", { name: "Release control" });

  expect(() => unmount()).not.toThrow();
  await waitFor(() =>
    expect(runtime.disconnect).toHaveBeenCalledWith("terminal-1"),
  );
  expect(terminalMock.dispose).toHaveBeenCalledTimes(1);
});
