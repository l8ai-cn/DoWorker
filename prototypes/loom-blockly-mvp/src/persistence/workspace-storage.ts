import * as Blockly from "blockly";

import {
  createCustomBlockDefinition,
  type CustomBlockDefinition,
} from "../custom-blocks/custom-block-definition";

const STORAGE_KEY = "loom-blockly-mvp.project.v1";

export type OutputTab = "diagnostics" | "json" | "evidence";

export interface LoomStoredProject {
  version: 1;
  workspaceState: Record<string, unknown>;
  customDefinitions: CustomBlockDefinition[];
  outputTab: OutputTab;
}

export interface LoomProjectLoadResult {
  project?: LoomStoredProject;
  error?: string;
}

export type StorageResult =
  | { ok: true }
  | { ok: false; error: string };

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function validOutputTab(value: unknown): value is OutputTab {
  return value === "diagnostics" || value === "json" || value === "evidence";
}

export function loadLoomProject(): LoomProjectLoadResult {
  try {
    const serialized = localStorage.getItem(STORAGE_KEY);
    if (!serialized) return {};
    const value: unknown = JSON.parse(serialized);
    if (
      !isRecord(value) ||
      value.version !== 1 ||
      !isRecord(value.workspaceState) ||
      !Array.isArray(value.customDefinitions) ||
      !validOutputTab(value.outputTab)
    ) {
      return { error: "本地项目结构无效，请清除后重新创建。" };
    }
    const customDefinitions: CustomBlockDefinition[] = [];
    const definitionIds = new Set<string>();
    for (const stored of value.customDefinitions) {
      if (!isRecord(stored)) {
        return { error: "本地自定义积木结构无效。" };
      }
      const result = createCustomBlockDefinition({
        id: String(stored.id ?? ""),
        name: String(stored.name ?? ""),
        template: String(stored.template ?? ""),
      });
      if (!result.definition) {
        return { error: result.errors.join(" ") };
      }
      if (definitionIds.has(result.definition.id)) {
        return { error: `自定义积木 ID 重复：${result.definition.id}。` };
      }
      definitionIds.add(result.definition.id);
      customDefinitions.push(result.definition);
    }
    return {
      project: {
        version: 1,
        workspaceState: value.workspaceState,
        customDefinitions,
        outputTab: value.outputTab,
      },
    };
  } catch (error) {
    return {
      error: error instanceof SyntaxError
        ? "本地项目不是有效 JSON，请清除后重新创建。"
        : `无法读取本地项目：${errorMessage(error)}`,
    };
  }
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "未知错误";
}

export function saveLoomProject(
  workspace: Blockly.Workspace,
  customDefinitions: CustomBlockDefinition[],
  outputTab: OutputTab,
): StorageResult {
  try {
    const project: LoomStoredProject = {
      version: 1,
      workspaceState: Blockly.serialization.workspaces.save(workspace),
      customDefinitions,
      outputTab,
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(project));
    return { ok: true };
  } catch (error) {
    return { ok: false, error: `无法保存本地项目：${errorMessage(error)}` };
  }
}

export function clearLoomProject(): StorageResult {
  try {
    localStorage.removeItem(STORAGE_KEY);
    return { ok: true };
  } catch (error) {
    return { ok: false, error: `无法清除本地项目：${errorMessage(error)}` };
  }
}
