"use client";

import { SemanticChangeOperation } from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import type { SemanticChange } from "@proto/orchestration_resource/v1/orchestration_resource_pb";
import { Badge } from "@/components/ui/badge";

interface ResourceSemanticDiffProps {
  changes: SemanticChange[];
  emptyLabel: string;
}

export function ResourceSemanticDiff({
  changes,
  emptyLabel,
}: ResourceSemanticDiffProps) {
  if (changes.length === 0) {
    return <p className="text-sm text-muted-foreground">{emptyLabel}</p>;
  }
  return (
    <div className="divide-y divide-border rounded-md border border-border">
      {changes.map((change, index) => (
        <div
          key={`${change.path}-${index}`}
          className="grid min-w-0 grid-cols-[auto_minmax(0,1fr)] items-start gap-3 px-3 py-2.5 sm:grid-cols-[auto_minmax(0,1fr)_auto] sm:items-center"
        >
          <Badge variant={change.operation === SemanticChangeOperation.REMOVE
            ? "warning"
            : "outline"}
          >
            {SemanticChangeOperation[change.operation]}
          </Badge>
          <code
            className="min-w-0 break-all whitespace-normal text-xs"
            title={change.path}
          >
            {change.path}
          </code>
          <span className="col-span-2 justify-self-end font-mono text-[11px] text-muted-foreground sm:col-span-1">
            {changeDigest(change.before)} → {changeDigest(change.after)}
          </span>
        </div>
      ))}
    </div>
  );
}

function changeDigest(
  value: SemanticChange["before"],
): string {
  if (!value || value.value.case === undefined) return "∅";
  if (value.value.case === "digest") {
    return value.value.value.slice(0, 10);
  }
  return "redacted";
}
