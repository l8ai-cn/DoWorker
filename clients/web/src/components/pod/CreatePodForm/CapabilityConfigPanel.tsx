"use client";

import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

interface CapabilityConfigPanelProps {
  icon: LucideIcon;
  title: string;
  description: string;
  className?: string;
  testId?: string;
  children: React.ReactNode;
}

export function CapabilityConfigPanel({
  icon: Icon,
  title,
  description,
  className,
  testId,
  children,
}: CapabilityConfigPanelProps) {
  return (
    <section
      data-testid={testId}
      className={cn(
        "rounded-lg border border-border bg-background/80 p-4 shadow-xs",
        className,
      )}
    >
      <header className="mb-3 flex gap-3">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary/10">
          <Icon className="h-4 w-4 text-primary" />
        </div>
        <div className="min-w-0">
          <h4 className="text-sm font-semibold">{title}</h4>
          <p className="mt-0.5 text-xs leading-5 text-muted-foreground">{description}</p>
        </div>
      </header>
      {children}
    </section>
  );
}
