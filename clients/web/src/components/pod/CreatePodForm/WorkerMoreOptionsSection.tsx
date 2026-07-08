"use client";

import { useState } from "react";
import { ChevronRight } from "lucide-react";
import {
  Collapsible,
  CollapsibleTrigger,
  CollapsibleContent,
} from "@/components/ui/collapsible";

interface WorkerMoreOptionsSectionProps {
  children: React.ReactNode;
  t: (key: string) => string;
}

export function WorkerMoreOptionsSection({ children, t }: WorkerMoreOptionsSectionProps) {
  const [open, setOpen] = useState(false);

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger className="flex w-full items-center gap-2 py-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground">
        <ChevronRight
          className={`h-4 w-4 transition-transform duration-200 ${open ? "rotate-90" : ""}`}
        />
        {t("ide.createPod.moreOptions")}
      </CollapsibleTrigger>
      <CollapsibleContent className="space-y-4 border-t border-border pt-4">
        {children}
      </CollapsibleContent>
    </Collapsible>
  );
}
