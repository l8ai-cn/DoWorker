import { create } from "zustand";
import { persist } from "zustand/middleware";
import { afterAll, afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { legacyPersistStorage } from "./zustand-legacy-persist";

interface TestState {
  hydrated: boolean;
  setHydrated: (hydrated: boolean) => void;
}

function createTestStore() {
  return create<TestState>()(
    persist(
      (set) => ({
        hydrated: false,
        setHydrated: (hydrated) => set({ hydrated }),
      }),
      {
        name: "do-worker-ide",
        storage: legacyPersistStorage("agentsmesh-ide"),
        onRehydrateStorage: () => (state) => state?.setHydrated(true),
      },
    ),
  );
}

describe("legacyPersistStorage", () => {
  const warn = vi.spyOn(console, "warn").mockImplementation(() => {});

  beforeEach(() => window.localStorage.clear());
  afterEach(() => {
    window.localStorage.clear();
    warn.mockClear();
  });
  afterAll(() => warn.mockRestore());

  it("recovers IDE hydration when persisted JSON is malformed", async () => {
    window.localStorage.setItem("do-worker-ide", "{invalid");

    const store = createTestStore();
    await Promise.resolve();
    await Promise.resolve();

    expect(store.getState().hydrated).toBe(true);
    expect(() => JSON.parse(window.localStorage.getItem("do-worker-ide") ?? "")).not.toThrow();
    expect(warn).toHaveBeenCalledWith('Discarding malformed persisted UI state "do-worker-ide".');
  });
});
