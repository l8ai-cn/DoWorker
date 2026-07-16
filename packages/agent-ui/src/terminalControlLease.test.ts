import { vi } from "vitest";

import { startTerminalLeaseRenewal } from "./terminalControlLease";

it("renews a granted terminal lease before expiry and stops after cleanup", async () => {
  vi.useFakeTimers();
  const renew = vi.fn(async () => undefined);
  const stop = startTerminalLeaseRenewal({
    expiresAt: 60_000,
    leaseId: "lease-1",
    now: () => 0,
    renew,
  });

  await vi.advanceTimersByTimeAsync(49_999);
  expect(renew).not.toHaveBeenCalled();
  await vi.advanceTimersByTimeAsync(1);
  expect(renew).toHaveBeenCalledWith("lease-1");

  stop();
  await vi.advanceTimersByTimeAsync(60_000);
  expect(renew).toHaveBeenCalledTimes(1);
  vi.useRealTimers();
});

it("reports renewal failures to the terminal surface", async () => {
  vi.useFakeTimers();
  const error = new Error("lease expired");
  const onError = vi.fn();
  const stop = startTerminalLeaseRenewal({
    expiresAt: 20_000,
    leaseId: "lease-1",
    now: () => 0,
    onError,
    renew: vi.fn(async () => {
      throw error;
    }),
  });

  await vi.advanceTimersByTimeAsync(10_000);

  expect(onError).toHaveBeenCalledWith(error);
  stop();
  vi.useRealTimers();
});
