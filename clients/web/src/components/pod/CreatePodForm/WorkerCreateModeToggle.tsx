"use client";

import { FileText, LayoutList } from "lucide-react";
import { cn } from "@/lib/utils";

interface WorkerCreateModeToggleProps {
  sourceMode: boolean;
  onChange: (sourceMode: boolean) => void;
  t: (key: string) => string;
}

export function WorkerCreateModeToggle({
  sourceMode,
  onChange,
  t,
}: WorkerCreateModeToggleProps) {
  return (
    <div
      role="tablist"
      aria-label={t("ide.createPod.modeToggleLabel")}
      className="inline-flex items-center gap-1 rounded-lg border border-border bg-muted/40 p-1"
    >
      <ModeButton
        active={!sourceMode}
        icon={<LayoutList className="h-4 w-4" />}
        label={t("ide.createPod.modeForm")}
        onClick={() => onChange(false)}
      />
      <ModeButton
        active={sourceMode}
        icon={<FileText className="h-4 w-4" />}
        label={t("ide.createPod.modeSource")}
        onClick={() => onChange(true)}
      />
    </div>
  );
}

function ModeButton({
  active,
  icon,
  label,
  onClick,
}: {
  active: boolean;
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      className={cn(
        "flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
        active
          ? "bg-background text-foreground shadow-xs"
          : "text-muted-foreground hover:text-foreground",
      )}
    >
      {icon}
      {label}
    </button>
  );
}
