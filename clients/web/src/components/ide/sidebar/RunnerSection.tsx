"use client";

import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { Runner } from "@/stores/runner";
import {
  Server,
  ChevronDown,
  ChevronRight,
  Loader2,
} from "lucide-react";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

interface RunnerSectionProps {
  runners: Runner[];
  loading: boolean;
  expanded: boolean;
  onToggle: (expanded: boolean) => void;
  currentOrgSlug?: string;
  t: (key: string) => string;
}

export function RunnerSection({
  runners,
  loading,
  expanded,
  onToggle,
  currentOrgSlug,
  t,
}: RunnerSectionProps) {
  const router = useRouter();
  const onlineRunners = runners.filter(r => r.status === "online");

  return (
    <Collapsible open={expanded} onOpenChange={onToggle}>
      <CollapsibleTrigger asChild>
        <div className="nav-row pressable cursor-pointer hover:bg-surface-muted motion-interactive">
          <div className="flex items-center gap-2">
            <Server className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm font-medium">{t("workspace.runners.title")}</span>
            <span className="text-xs text-muted-foreground">
              ({onlineRunners.length} {t("workspace.runners.online")})
            </span>
          </div>
          {expanded ? (
            <ChevronDown className="w-4 h-4 text-muted-foreground" />
          ) : (
            <ChevronRight className="w-4 h-4 text-muted-foreground" />
          )}
        </div>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="py-1">
          {loading ? (
            <div className="flex items-center justify-center py-4">
              <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />
            </div>
          ) : runners.length === 0 ? (
            <div className="px-3 py-4 text-center">
              <p className="text-xs text-muted-foreground">{t("workspace.runners.noRunners")}</p>
            </div>
          ) : (
            <div className="py-1 max-h-32 overflow-y-auto">
              {runners.map((runner) => (
                <div
                  key={runner.id}
                  className="nav-row pressable text-sm cursor-pointer motion-interactive nav-row-idle hover:bg-surface-muted"
                  onClick={() => router.push(`/${currentOrgSlug}/infra?tab=runners&id=${runner.id}`)}
                >
                  <span
                    className={cn(
                      "w-2 h-2 rounded-full flex-shrink-0",
                      runner.status === "online" ? "bg-success" : "bg-muted-foreground"
                    )}
                  />
                  <span className="truncate flex-1">{runner.node_id}</span>
                  <span className="text-xs text-muted-foreground">
                    {runner.current_pods}/{runner.max_concurrent_pods}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}
