import type { Diagnostic, LoopLimits } from "./loop-types";

export function issue(
  code: string,
  message: string,
  blockId?: string,
  slot?: string,
): Diagnostic {
  return {
    code,
    message,
    ...(blockId ? { blockId } : {}),
    ...(slot ? { slot } : {}),
  };
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export function isNonBlankString(value: unknown): value is string {
  return typeof value === "string" && value.trim() !== "";
}

export function isPositiveInteger(value: unknown): value is number {
  return Number.isSafeInteger(value) && Number(value) > 0;
}

export function validLimits(value: unknown): value is LoopLimits {
  if (!isRecord(value)) return false;
  return (
    isPositiveInteger(value.maxIterations) &&
    value.maxIterations <= 100 &&
    isPositiveInteger(value.tokenBudget) &&
    isPositiveInteger(value.timeoutMinutes) &&
    isPositiveInteger(value.noProgressLimit) &&
    isPositiveInteger(value.sameErrorLimit)
  );
}

export function validateTextBlocks(
  value: unknown,
  rootBlockId: string,
  kind: "instruction" | "acceptance-criterion",
  slot: string,
): Diagnostic[] {
  if (!Array.isArray(value) || value.length === 0) {
    const code = kind === "instruction"
      ? "missing-instructions"
      : "missing-acceptance-criteria";
    return [issue(
      code,
      kind === "instruction" ? "至少需要一个任务积木。" : "至少需要一个验收条件积木。",
      rootBlockId,
      slot,
    )];
  }

  const diagnostics: Diagnostic[] = [];
  for (const item of value) {
    if (!isRecord(item)) {
      diagnostics.push(issue(`invalid-${kind}`, "积木结构无效。"));
      continue;
    }
    const blockId = isNonBlankString(item.blockId) ? item.blockId : rootBlockId;
    if (!isNonBlankString(item.blockId)) {
      diagnostics.push(issue(
        `invalid-${kind}-block-id`,
        "积木 ID 不能为空。",
        rootBlockId,
      ));
    }
    if (typeof item.value !== "string") {
      diagnostics.push(issue(`invalid-${kind}`, "积木文本必须是字符串。", blockId));
    } else if (item.value.trim() === "") {
      diagnostics.push(issue(`empty-${kind}`, "积木文本不能为空。", blockId));
    }
  }
  return diagnostics;
}

export function validateAdapterDiagnostics(
  value: unknown,
  rootBlockId: string,
): Diagnostic[] {
  if (!Array.isArray(value)) {
    return [issue(
      "invalid-adapter-diagnostic-list",
      "工作区适配诊断必须是数组。",
      rootBlockId,
    )];
  }
  const diagnostics: Diagnostic[] = [];
  for (const entry of value) {
    if (
      !isRecord(entry) ||
      !isNonBlankString(entry.code) ||
      !isNonBlankString(entry.message) ||
      (entry.blockId !== undefined && !isNonBlankString(entry.blockId)) ||
      (entry.slot !== undefined && !isNonBlankString(entry.slot))
    ) {
      diagnostics.push(issue(
        "invalid-adapter-diagnostic",
        "工作区适配诊断结构无效。",
        rootBlockId,
      ));
      continue;
    }
    diagnostics.push({
      code: entry.code,
      message: entry.message,
      ...(entry.blockId ? { blockId: entry.blockId } : {}),
      ...(entry.slot ? { slot: entry.slot } : {}),
    });
  }
  return diagnostics;
}
