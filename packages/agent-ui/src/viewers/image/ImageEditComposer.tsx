import { Send } from "lucide-react";
import { useState } from "react";

import type { AgentArtifactActionCommand } from "../../contracts";
import type { ImageDimensions } from "./imageGeometry";

export interface NormalizedImageRegion {
  height: number;
  width: number;
  x: number;
  y: number;
}

export interface EditImagePayload {
  instruction: string;
  normalizedRegion?: NormalizedImageRegion;
  sourceDimensions: ImageDimensions;
}

export type EditImageAction = AgentArtifactActionCommand<
  "image.edit",
  EditImagePayload
>;

export interface ImageEditComposerProps {
  actionSchemaVersion: string;
  artifactId: string;
  baseRevision: bigint;
  disabled?: boolean;
  normalizedRegion?: NormalizedImageRegion;
  onSubmit: (action: EditImageAction) => void;
  representationId?: string;
  sourceDimensions: ImageDimensions;
}

export function ImageEditComposer({
  actionSchemaVersion,
  artifactId,
  baseRevision,
  disabled = false,
  normalizedRegion,
  onSubmit,
  representationId,
  sourceDimensions,
}: ImageEditComposerProps) {
  const [instruction, setInstruction] = useState("");
  const submittedInstruction = instruction.trim();

  return (
    <form
      className="space-y-3 rounded-md border border-border bg-card p-3"
      onSubmit={(event) => {
        event.preventDefault();
        if (!submittedInstruction || disabled) return;
        onSubmit({
          actionSchemaVersion,
          actionType: "image.edit",
          artifactId,
          baseRevision,
          commandId: crypto.randomUUID(),
          payload: {
            instruction: submittedInstruction,
            ...(normalizedRegion ? { normalizedRegion } : {}),
            sourceDimensions,
          },
          ...(representationId ? { representationId } : {}),
        });
      }}
    >
      <label className="block space-y-1.5 text-sm font-medium">
        编辑说明
        <textarea
          aria-label="编辑说明"
          className="min-h-24 w-full resize-y rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground outline-none placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring"
          disabled={disabled}
          onChange={(event) => setInstruction(event.target.value)}
          placeholder="描述希望对图片进行的修改"
          value={instruction}
        />
      </label>
      <div className="flex justify-end">
        <button
          className="inline-flex h-11 items-center gap-1.5 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground outline-none hover:opacity-90 focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
          disabled={disabled || !submittedInstruction}
          type="submit"
        >
          <Send className="size-4" />
          提交编辑
        </button>
      </div>
    </form>
  );
}
