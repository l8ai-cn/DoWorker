import * as Blockly from "blockly";
import { afterEach, describe, expect, it, vi } from "vitest";

import {
  clearLoomProject,
  loadLoomProject,
  saveLoomProject,
} from "./workspace-storage";

const originalStorage = globalThis.localStorage;

afterEach(() => {
  vi.restoreAllMocks();
  Object.defineProperty(globalThis, "localStorage", {
    configurable: true,
    value: originalStorage,
  });
});

function storage(overrides: Partial<Storage>): Storage {
  return {
    clear: vi.fn(),
    getItem: vi.fn(() => null),
    key: vi.fn(() => null),
    length: 0,
    removeItem: vi.fn(),
    setItem: vi.fn(),
    ...overrides,
  };
}

describe("workspace storage", () => {
  it("returns an error when reading localStorage throws", () => {
    Object.defineProperty(globalThis, "localStorage", {
      configurable: true,
      value: storage({
        getItem: vi.fn(() => {
          throw new Error("blocked");
        }),
      }),
    });

    expect(loadLoomProject()).toEqual({
      error: "无法读取本地项目：blocked",
    });
  });

  it("does not report save success when localStorage rejects the write", () => {
    Object.defineProperty(globalThis, "localStorage", {
      configurable: true,
      value: storage({
        setItem: vi.fn(() => {
          throw new Error("quota");
        }),
      }),
    });

    expect(saveLoomProject(new Blockly.Workspace(), [], "diagnostics")).toEqual({
      ok: false,
      error: "无法保存本地项目：quota",
    });
  });

  it("rejects duplicate custom block IDs on load", () => {
    const project = {
      version: 1,
      workspaceState: {},
      outputTab: "diagnostics",
      customDefinitions: [
        { id: "same-id", name: "A", template: "A" },
        { id: "same-id", name: "B", template: "B" },
      ],
    };
    Object.defineProperty(globalThis, "localStorage", {
      configurable: true,
      value: storage({ getItem: vi.fn(() => JSON.stringify(project)) }),
    });

    expect(loadLoomProject()).toEqual({
      error: "自定义积木 ID 重复：same-id。",
    });
  });

  it("returns an error when clearing localStorage throws", () => {
    Object.defineProperty(globalThis, "localStorage", {
      configurable: true,
      value: storage({
        removeItem: vi.fn(() => {
          throw new Error("blocked");
        }),
      }),
    });

    expect(clearLoomProject()).toEqual({
      ok: false,
      error: "无法清除本地项目：blocked",
    });
  });
});
