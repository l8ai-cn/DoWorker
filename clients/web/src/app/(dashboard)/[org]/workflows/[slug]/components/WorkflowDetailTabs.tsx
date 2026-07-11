"use client";

import { PillTabs, type PillTabItem } from "@/components/ui/pill-tabs";

interface WorkflowDetailTabsProps {
  active: string;
  onChange: (id: string) => void;
  tabs: PillTabItem[];
  rightSlot?: React.ReactNode;
}

export function WorkflowDetailTabs({ active, onChange, tabs, rightSlot }: WorkflowDetailTabsProps) {
  return (
    <div className="mb-4 flex items-center gap-2">
      <PillTabs
        active={active}
        onChange={onChange}
        tabs={tabs.map((tab) => ({ ...tab, testId: tab.testId ?? `workflow-tab-${tab.id}` }))}
        stretch
      />
      {rightSlot}
    </div>
  );
}
