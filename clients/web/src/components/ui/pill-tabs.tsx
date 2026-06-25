"use client";

import { cn } from "@/lib/utils";

export interface PillTabItem {
  id: string;
  label: string;
  testId?: string;
}

interface PillTabsProps {
  active: string;
  onChange: (id: string) => void;
  tabs: PillTabItem[];
  className?: string;
  stretch?: boolean;
}

export function PillTabs({ active, onChange, tabs, className, stretch }: PillTabsProps) {
  return (
    <div
      className={cn(
        "inline-flex rounded-lg bg-surface-muted p-1",
        stretch && "flex w-full",
        className,
      )}
      role="tablist"
    >
      {tabs.map((tab) => {
        const isActive = active === tab.id;
        return (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={isActive}
            data-testid={tab.testId ?? `pill-tab-${tab.id}`}
            onClick={() => onChange(tab.id)}
            className={cn(
              "motion-interactive rounded-md px-4 py-2 text-sm font-medium",
              stretch && "flex-1",
              isActive
                ? "bg-surface-raised text-primary shadow-[var(--shadow-soft)]"
                : "text-muted-foreground hover:text-foreground",
            )}
          >
            {tab.label}
          </button>
        );
      })}
    </div>
  );
}

interface PillTabsRowProps {
  children: React.ReactNode;
  className?: string;
}

export function PillTabsRow({ children, className }: PillTabsRowProps) {
  return <div className={cn("mb-6 flex items-center gap-2", className)}>{children}</div>;
}
