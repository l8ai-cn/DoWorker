"use client";

import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import type { CoordinatorExecution } from "@/lib/api/coordinatorApi";
import { EXECUTION_COLUMNS, statusBadgeVariant, statusColumn } from "./executionStatus";

interface Props {
  executions: CoordinatorExecution[];
}

export function ExecutionBoard({ executions }: Props) {
  const t = useTranslations("automation");

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
      {EXECUTION_COLUMNS.map((column) => {
        const items = executions.filter((e) => statusColumn(e.status) === column);
        return (
          <div key={column} className="flex flex-col gap-2">
            <div className="flex items-center justify-between px-1">
              <span className="text-sm font-medium">{t(`status.${column}`)}</span>
              <span className="text-xs text-muted-foreground">{items.length}</span>
            </div>
            {items.map((e) => (
              <Card key={e.id} className="p-3">
                <div className="flex items-center justify-between gap-2">
                  <span className="truncate text-sm font-medium">#{e.external_id || e.id}</span>
                  <Badge variant={statusBadgeVariant(e.status)}>{e.status}</Badge>
                </div>
                {e.summary && <p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{e.summary}</p>}
                {e.error && <p className="mt-1 line-clamp-2 text-xs text-destructive">{e.error}</p>}
              </Card>
            ))}
            {items.length === 0 && (
              <div className="rounded-md border border-dashed py-6 text-center text-xs text-muted-foreground">
                {t("board.empty")}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
