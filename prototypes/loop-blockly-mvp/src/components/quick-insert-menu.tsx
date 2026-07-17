import { Blocks, Plus } from "lucide-react";
import { useEffect, useRef, type CSSProperties } from "react";

import { LOOP_BLOCK_TYPES } from "../blockly/block-catalog";
import {
  customBlockType,
  type CustomBlockDefinition,
} from "../custom-blocks/custom-block-definition";

interface QuickInsertMenuProps {
  customDefinitions: CustomBlockDefinition[];
  hasRoot: boolean;
  onClose: () => void;
  onCreateCustom: () => void;
  onInsert: (type: string) => void;
  point: { x: number; y: number };
}

const ITEMS = [
  ["Goal Loop", LOOP_BLOCK_TYPES.root],
  ["使用 Worker", LOOP_BLOCK_TYPES.worker],
  ["执行任务", LOOP_BLOCK_TYPES.instruction],
  ["验收条件", LOOP_BLOCK_TYPES.acceptance],
  ["运行验证命令", LOOP_BLOCK_TYPES.verifier],
  ["执行边界", LOOP_BLOCK_TYPES.limits],
  ["失败处理", LOOP_BLOCK_TYPES.escalation],
] as const;

export function QuickInsertMenu({
  customDefinitions,
  hasRoot,
  onClose,
  onCreateCustom,
  onInsert,
  point,
}: QuickInsertMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const close = (event: PointerEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) onClose();
    };
    window.addEventListener("pointerdown", close);
    return () => window.removeEventListener("pointerdown", close);
  }, [onClose]);

  return (
    <div
      className="quick-insert"
      ref={menuRef}
      style={{
        "--quick-x": `${point.x}px`,
        "--quick-y": `${point.y}px`,
      } as CSSProperties}
    >
      <div className="quick-insert-title">插入积木</div>
      {ITEMS.map(([label, type]) => (
        <button
          disabled={type === LOOP_BLOCK_TYPES.root && hasRoot}
          key={type}
          onClick={() => onInsert(type)}
          type="button"
        >
          <Blocks size={15} />
          <span>{label}</span>
        </button>
      ))}
      {customDefinitions.map((definition) => (
        <button
          key={definition.id}
          onClick={() => onInsert(customBlockType(definition.id))}
          type="button"
        >
          <Blocks size={15} />
          <span>{definition.name}</span>
        </button>
      ))}
      <button className="quick-create" onClick={onCreateCustom} type="button">
        <Plus size={15} />
        <span>创建自定义积木</span>
      </button>
    </div>
  );
}
