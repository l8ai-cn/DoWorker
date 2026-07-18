import { describe, expect, it } from "vitest";
import deMessages from "@/messages/de/app.json";
import enMessages from "@/messages/en/app.json";
import esMessages from "@/messages/es/app.json";
import frMessages from "@/messages/fr/app.json";
import jaMessages from "@/messages/ja/app.json";
import koMessages from "@/messages/ko/app.json";
import ptMessages from "@/messages/pt/app.json";
import zhMessages from "@/messages/zh/app.json";
import {
  createLoopWorkbenchMessages,
  type LoopMessageTranslator,
} from "../loop-workbench-messages";

function translator(messages: Record<string, unknown>): LoopMessageTranslator {
  return (key, values) => {
    const value = key.split(".").reduce<unknown>(
      (current, segment) => (current as Record<string, unknown>)[segment],
      messages,
    );
    return Object.entries(values ?? {}).reduce(
      (text, [name, replacement]) => text.replace(`{${name}}`, String(replacement)),
      String(value),
    );
  };
}

function paths(value: unknown, prefix = ""): string[] {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return [prefix];
  }
  return Object.entries(value).flatMap(([key, child]) =>
    paths(child, prefix ? `${prefix}.${key}` : key));
}

describe("Loop workbench messages", () => {
  it.each([
    ["English", enMessages.loopWorkbench],
    ["Chinese", zhMessages.loopWorkbench],
    ["German", deMessages.loopWorkbench],
    ["Spanish", esMessages.loopWorkbench],
    ["French", frMessages.loopWorkbench],
    ["Japanese", jaMessages.loopWorkbench],
    ["Korean", koMessages.loopWorkbench],
    ["Portuguese", ptMessages.loopWorkbench],
  ])("keeps %s Loop keys aligned with English", (_, messages) => {
    expect(paths(messages).sort()).toEqual(paths(enMessages.loopWorkbench).sort());
  });

  it("preserves the current Chinese labels", () => {
    const messages = createLoopWorkbenchMessages(
      translator(zhMessages.loopWorkbench),
    );

    expect(messages.toolbar.title).toBe("循环工作台");
    expect(messages.shell.canvasTitle).toBe("积木画布");
    expect(messages.blockly.loop.message0).toBe("循环 %1");
    expect(messages.quickInsert.title).toBe("插入积木");
    expect(messages.status.parseStatus.valid).toBe("有效");
    expect(messages.runtime.title).toBe("选择运行环境");
    expect(messages.ai.generateMode).toBe("生成 / 修改");
    expect(messages.ai.explainMode).toBe("解释当前循环");
    expect(messages.ai.projection.limits).toBe("执行预算");
    expect(messages.shell.editorMetadata(7, 4)).toBe("编辑版本 7 · 语义版本 4");
  });

  it("builds a complete English label set", () => {
    const messages = createLoopWorkbenchMessages(
      translator(enMessages.loopWorkbench),
    );

    expect(messages.toolbar.title).toBe("Loop workbench");
    expect(messages.shell.canvasTitle).toBe("Block canvas");
    expect(messages.blockly.loop.message0).toBe("Loop %1");
    expect(messages.quickInsert.title).toBe("Insert block");
    expect(messages.status.parseStatus.valid).toBe("Valid");
    expect(messages.runtime.title).toBe("Select runtime");
    expect(messages.ai.generateMode).toBe("Generate / modify");
    expect(messages.ai.explainMode).toBe("Explain current loop");
    expect(messages.ai.projection.limits).toBe("Execution budget");
    expect(messages.runtime.snapshotLabel("Review", "codex", "42")).toBe(
      "Review · codex · Snapshot 42",
    );
  });
});
