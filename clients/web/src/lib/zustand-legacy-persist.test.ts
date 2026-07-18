import { afterEach, describe, expect, it, vi } from "vitest";
import { legacyPersistStorage } from "./zustand-legacy-persist";

describe("legacyPersistStorage", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("does not access browser storage during server rendering", () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.stubGlobal("localStorage", undefined);
    const storage = legacyPersistStorage("legacy-workspace");

    expect(storage.getItem("workspace")).toBeNull();
    expect(() => storage.setItem("workspace", '{"state":{}}')).not.toThrow();
    expect(() => storage.removeItem("workspace")).not.toThrow();
    expect(warn).not.toHaveBeenCalled();
  });
});
