import type * as Blockly from "blockly";

import { LOOP_BLOCK_TYPES } from "../blockly/block-catalog";
import {
  customBlockType,
  type CustomBlockDefinition,
} from "../custom-blocks/custom-block-definition";
import type { Diagnostic } from "../domain/loop-types";

interface FieldSpec {
  name: string;
  label: string;
  kind: "text" | "textarea" | "number" | "select";
  options?: { label: string; value: string }[];
}

const BUILT_IN_FIELDS: Record<string, FieldSpec[]> = {
  [LOOP_BLOCK_TYPES.root]: [
    { name: "NAME", label: "Loop 名称", kind: "text" },
  ],
  [LOOP_BLOCK_TYPES.worker]: [
    { name: "SNAPSHOT_ID", label: "Worker 快照 ID", kind: "number" },
    { name: "LABEL", label: "显示名称", kind: "text" },
  ],
  [LOOP_BLOCK_TYPES.instruction]: [
    { name: "TEXT", label: "任务指令", kind: "textarea" },
  ],
  [LOOP_BLOCK_TYPES.acceptance]: [
    { name: "TEXT", label: "验收条件", kind: "textarea" },
  ],
  [LOOP_BLOCK_TYPES.verifier]: [
    { name: "COMMAND", label: "验证命令", kind: "textarea" },
  ],
  [LOOP_BLOCK_TYPES.limits]: [
    { name: "MAX_ITERATIONS", label: "最大迭代", kind: "number" },
    { name: "TOKEN_BUDGET", label: "Token 预算", kind: "number" },
    { name: "TIMEOUT_MINUTES", label: "超时分钟", kind: "number" },
    { name: "NO_PROGRESS_LIMIT", label: "无进展上限", kind: "number" },
    { name: "SAME_ERROR_LIMIT", label: "同错上限", kind: "number" },
  ],
  [LOOP_BLOCK_TYPES.escalation]: [{
    name: "POLICY",
    label: "失败策略",
    kind: "select",
    options: [
      { label: "暂停并等待人工", value: "pause" },
      { label: "直接失败", value: "fail" },
    ],
  }],
};

function fieldsFor(
  block: Blockly.BlockSvg,
  definitions: CustomBlockDefinition[],
): FieldSpec[] {
  const builtIn = BUILT_IN_FIELDS[block.type];
  if (builtIn) return builtIn;
  const custom = definitions.find(
    ({ id }) => customBlockType(id) === block.type,
  );
  return custom?.parameters.map((parameter) => ({
    name: parameter,
    label: parameter,
    kind: "text" as const,
  })) ?? [];
}

interface BlockInspectorProps {
  block: Blockly.BlockSvg | null;
  customDefinitions: CustomBlockDefinition[];
  diagnostics: Diagnostic[];
  disabled?: boolean;
}

export function BlockInspector({
  block,
  customDefinitions,
  diagnostics,
  disabled = false,
}: BlockInspectorProps) {
  const blockDiagnostics = block
    ? diagnostics.filter(({ blockId }) => blockId === block.id)
    : diagnostics.filter(({ blockId }) => !blockId);

  return (
    <aside className="inspector-panel" aria-label="积木参数">
      <div className="panel-heading">
        <span>参数</span>
        {block && <code>{block.type}</code>}
      </div>
      {!block ? (
        <div className="panel-empty">未选择积木</div>
      ) : (
        <div className="inspector-fields">
          {fieldsFor(block, customDefinitions).map((field) => (
            <label className="field-control" key={field.name}>
              <span>{field.label}</span>
              {field.kind === "textarea" ? (
                <textarea
                  disabled={disabled}
                  rows={4}
                  value={String(block.getFieldValue(field.name) ?? "")}
                  onChange={(event) => block.setFieldValue(
                    event.target.value,
                    field.name,
                  )}
                />
              ) : field.kind === "select" ? (
                <select
                  disabled={disabled}
                  value={String(block.getFieldValue(field.name) ?? "")}
                  onChange={(event) => block.setFieldValue(
                    event.target.value,
                    field.name,
                  )}
                >
                  {field.options?.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              ) : (
                <input
                  disabled={disabled}
                  type={field.kind}
                  min={field.kind === "number" ? 1 : undefined}
                  value={String(block.getFieldValue(field.name) ?? "")}
                  onChange={(event) => block.setFieldValue(
                    event.target.value,
                    field.name,
                  )}
                />
              )}
            </label>
          ))}
          {fieldsFor(block, customDefinitions).length === 0 && (
            <div className="panel-empty">此积木没有可编辑参数</div>
          )}
        </div>
      )}
      {blockDiagnostics.length > 0 && (
        <div className="inspector-errors">
          {blockDiagnostics.map((diagnostic) => (
            <div key={`${diagnostic.code}-${diagnostic.slot ?? ""}`}>
              {diagnostic.message}
            </div>
          ))}
        </div>
      )}
    </aside>
  );
}
