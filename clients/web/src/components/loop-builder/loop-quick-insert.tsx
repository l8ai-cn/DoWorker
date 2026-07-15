import { Plus } from "lucide-react";
import { LOOP_BLOCK_TYPES } from "./loop-block-catalog";

const options = [
  ["Loop", LOOP_BLOCK_TYPES.loop],
  ["Worker", LOOP_BLOCK_TYPES.worker],
  ["Repeat", LOOP_BLOCK_TYPES.repeat],
  ["Agent", LOOP_BLOCK_TYPES.agent],
  ["Verifier", LOOP_BLOCK_TYPES.verifier],
  ["Limits", LOOP_BLOCK_TYPES.limits],
  ["Failure", LOOP_BLOCK_TYPES.failure],
] as const;

interface LoopQuickInsertProps {
  x: number;
  y: number;
  onInsert: (type: string) => void;
  onClose: () => void;
}

export function LoopQuickInsert({ x, y, onInsert, onClose }: LoopQuickInsertProps) {
  return (
    <>
      <button
        aria-label="关闭快速插入"
        className="absolute inset-0 z-10 cursor-default"
        onClick={onClose}
        type="button"
      />
      <div
        className="absolute z-20 w-44 overflow-hidden rounded-md border border-border bg-popover py-1 shadow-lg"
        style={{
          left: `clamp(0.5rem, ${x}px, calc(100% - 11.5rem))`,
          top: `clamp(0.5rem, ${y}px, calc(100% - 20rem))`,
        }}
      >
        <div className="border-b border-border px-3 py-2 text-xs font-medium text-muted-foreground">
          插入积木
        </div>
        {options.map(([label, type]) => (
          <button
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-accent"
            key={type}
            onClick={() => onInsert(type)}
            type="button"
          >
            <Plus className="h-3.5 w-3.5" />
            {label}
          </button>
        ))}
      </div>
    </>
  );
}
