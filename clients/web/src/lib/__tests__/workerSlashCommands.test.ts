import { describe, expect, it } from "vitest";
import {
  buildWorkerSlashCommands,
  filterWorkerSlashCommands,
  isWorkerSlashCommandText,
  parseWorkerSlashInput,
} from "@/lib/workerSlashCommands";
import { getSlashQuery } from "@/lib/slashQuery";

const t = (key: string) => key;

describe("workerSlashCommands", () => {
  it("detects slash query at input start", () => {
    expect(getSlashQuery("/comp", 5)).toEqual({ query: "comp", startIndex: 0 });
    expect(getSlashQuery("/goal ship", 10)).toBeNull();
  });

  it("filters commands by prefix", () => {
    const cmds = buildWorkerSlashCommands(t);
    expect(cmds.find((command) => command.id === "compact")?.hint).toBe(
      "workerSlash.commands.compact",
    );
    expect(filterWorkerSlashCommands(cmds, "comp").map((c) => c.id)).toEqual(["compact"]);
  });

  it("parses slash input with optional args", () => {
    const cmds = buildWorkerSlashCommands(t);
    expect(parseWorkerSlashInput("/compact", cmds)?.command.id).toBe("compact");
    expect(parseWorkerSlashInput("/model gpt-5", cmds)?.arg).toBe("gpt-5");
  });

  it("recognizes slash command text", () => {
    expect(isWorkerSlashCommandText("/compact")).toBe(true);
    expect(isWorkerSlashCommandText("/etc/hosts")).toBe(false);
  });
});
