import { Plus } from "lucide-react";
import { LOOP_BLOCK_TYPES } from "./loop-block-catalog";
import { customBlockType, type LoopCustomBlockDefinition } from "./loop-custom-block-types";
import type { LoopQuickInsertMessages } from "./loop-workbench-messages";

interface LoopQuickInsertProps {
  x: number;
  y: number;
  messages: LoopQuickInsertMessages;
  customDefinitions: readonly LoopCustomBlockDefinition[];
  onInsert: (type: string) => void;
  onCreateCustom: () => void;
  onClose: () => void;
}

export function LoopQuickInsert({
  x,
  y,
  messages,
  customDefinitions,
  onInsert,
  onCreateCustom,
  onClose,
}: LoopQuickInsertProps) {
  const options = [
    [messages.options.loop, LOOP_BLOCK_TYPES.loop],
    [messages.options.repeat, LOOP_BLOCK_TYPES.repeat],
    [messages.options.agent, LOOP_BLOCK_TYPES.agent],
    [messages.options.verifier, LOOP_BLOCK_TYPES.verifier],
    [messages.options.limits, LOOP_BLOCK_TYPES.limits],
    [messages.options.failure, LOOP_BLOCK_TYPES.failure],
  ] as const;
  return (
    <>
      <button
        aria-label={messages.close}
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
          {messages.title}
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
        <div className="border-t border-border px-3 py-2 text-xs font-medium text-muted-foreground">
          {messages.customTitle}
        </div>
        {customDefinitions.length === 0 && (
          <div className="px-3 py-2 text-xs text-muted-foreground">
            {messages.customEmpty}
          </div>
        )}
        {customDefinitions.map((definition) => (
          <button
            className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-accent"
            key={`${definition.slug}@${definition.version}`}
            onClick={() => onInsert(customBlockType(definition))}
            type="button"
          >
            <Plus className="h-3.5 w-3.5" />
            {definition.label}
          </button>
        ))}
        <button
          className="flex w-full items-center gap-2 border-t border-border px-3 py-2 text-left text-sm hover:bg-accent"
          onClick={onCreateCustom}
          type="button"
        >
          <Plus className="h-3.5 w-3.5" />
          {messages.createCustom}
        </button>
      </div>
    </>
  );
}
