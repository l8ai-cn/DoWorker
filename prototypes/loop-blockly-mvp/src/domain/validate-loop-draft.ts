import {
  isPositiveInteger,
  isNonBlankString,
  isRecord,
  issue,
  validLimits,
  validateAdapterDiagnostics,
  validateTextBlocks,
} from "./loop-draft-validation-primitives";
import type { Diagnostic, LoopDraft } from "./loop-types";

export function validateLoopDraft(draft: LoopDraft): Diagnostic[] {
  const value = draft as unknown;
  if (!isRecord(value)) {
    return [issue("invalid-draft", "Loop 草稿必须是对象。")];
  }
  const rootBlockId = isNonBlankString(value.rootBlockId)
    ? value.rootBlockId
    : "";
  const diagnosticRootId = rootBlockId || "workspace";
  const diagnostics: Diagnostic[] = [];
  diagnostics.push(...validateAdapterDiagnostics(
    value.adapterDiagnostics,
    diagnosticRootId,
  ));

  if (!rootBlockId) {
    diagnostics.push(issue(
      "invalid-root-block-id",
      "根积木 ID 不能为空。",
    ));
  }
  if (typeof value.name !== "string" || value.name.trim() === "") {
    diagnostics.push(issue("missing-name", "需要填写 Loop 名称。", rootBlockId, "name"));
  }

  if (!isRecord(value.worker)) {
    diagnostics.push(issue("missing-worker", "需要连接一个 Worker。", rootBlockId, "worker"));
  } else if (
    !isNonBlankString(value.worker.blockId) ||
    !isRecord(value.worker.value) ||
    !isPositiveInteger(value.worker.value.snapshotId) ||
    typeof value.worker.value.label !== "string" ||
    value.worker.value.label.trim() === ""
  ) {
    const blockId = isNonBlankString(value.worker.blockId)
      ? value.worker.blockId
      : diagnosticRootId;
    diagnostics.push(issue("invalid-worker-snapshot",
      "Worker 快照 ID 必须是正整数，名称不能为空。", blockId));
  }

  diagnostics.push(...validateTextBlocks(
    value.instructions,
    rootBlockId,
    "instruction",
    "instructions",
  ));
  diagnostics.push(...validateTextBlocks(
    value.acceptanceCriteria,
    rootBlockId,
    "acceptance-criterion",
    "acceptance-criteria",
  ));

  if (!isRecord(value.verification)) {
    diagnostics.push(issue(
      "missing-verification",
      "需要连接验证命令积木。",
      rootBlockId,
      "verification",
    ));
  } else if (
    !isNonBlankString(value.verification.blockId) ||
    typeof value.verification.value !== "string" ||
    value.verification.value.trim() === ""
  ) {
    const blockId = isNonBlankString(value.verification.blockId)
      ? value.verification.blockId
      : diagnosticRootId;
    diagnostics.push(issue(
      isNonBlankString(value.verification.blockId) &&
        typeof value.verification.value === "string"
        ? "empty-verification"
        : "invalid-verification",
      "验证命令不能为空。",
      blockId,
    ));
  }

  if (!isRecord(value.limits)) {
    diagnostics.push(issue("missing-limits", "需要连接执行边界积木。", rootBlockId, "limits"));
  } else if (
    !isNonBlankString(value.limits.blockId) ||
    !validLimits(value.limits.value)
  ) {
    const blockId = isNonBlankString(value.limits.blockId)
      ? value.limits.blockId
      : diagnosticRootId;
    diagnostics.push(issue("invalid-limits",
      "边界值必须是安全正整数，最大迭代不能超过 100。",
      blockId));
  }

  if (!isRecord(value.escalationPolicy)) {
    diagnostics.push(issue(
      "missing-escalation-policy",
      "需要连接失败处理积木。",
      rootBlockId,
      "escalation",
    ));
  } else if (
    !isNonBlankString(value.escalationPolicy.blockId) ||
    (value.escalationPolicy.value !== "pause" &&
      value.escalationPolicy.value !== "fail")
  ) {
    const blockId = isNonBlankString(value.escalationPolicy.blockId)
      ? value.escalationPolicy.blockId
      : diagnosticRootId;
    diagnostics.push(issue("invalid-escalation-policy",
      "失败策略只能是暂停或失败。", blockId));
  }

  if (!Array.isArray(value.looseBlockIds)) {
    diagnostics.push(issue(
      "invalid-loose-block-list",
      "散落积木元数据必须是数组。",
      rootBlockId,
    ));
  } else {
    for (const blockId of value.looseBlockIds) {
      if (typeof blockId !== "string" || blockId.trim() === "") {
        diagnostics.push(issue(
          "invalid-loose-block-id",
          "散落积木 ID 不能为空。",
          rootBlockId,
        ));
        continue;
      }
      diagnostics.push(issue(
        "loose-block",
        "所有积木都必须连接到 Goal Loop 根积木。",
        blockId,
      ));
    }
  }
  if (!Array.isArray(value.unknownBlockTypes)) {
    diagnostics.push(issue(
      "invalid-unknown-block-list",
      "未知积木元数据必须是数组。",
      rootBlockId,
    ));
  } else {
    for (const unknown of value.unknownBlockTypes) {
      if (
        !isRecord(unknown) ||
        !isNonBlankString(unknown.blockId) ||
        !isNonBlankString(unknown.type)
      ) {
        diagnostics.push(issue(
          "invalid-unknown-block-entry",
          "未知积木必须包含非空 blockId 和 type。",
          rootBlockId,
        ));
        continue;
      }
      diagnostics.push(issue(
        "unknown-block-type",
        `不支持的积木类型：${unknown.type}。`,
        unknown.blockId,
      ));
    }
  }

  return diagnostics;
}
