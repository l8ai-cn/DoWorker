"use client";

import { PillTabs, type PillTabItem } from "@/components/ui/pill-tabs";

interface LoopDetailTabsProps {
  active: string;
  onChange: (id: string) => void;
  tabs: PillTabItem[];
  rightSlot?: React.ReactNode;
}

export function LoopDetailTabs({ active, onChange, tabs, rightSlot }: LoopDetailTabsProps) {
  return (
    <div className="mb-4 flex items-center gap-2">
      <PillTabs
        active={active}
        onChange={onChange}
        tabs={tabs.map((tab) => ({ ...tab, testId: tab.testId ?? `loop-tab-${tab.id}` }))}
        stretch
      />
      {rightSlot}
    </div>
  );
}
