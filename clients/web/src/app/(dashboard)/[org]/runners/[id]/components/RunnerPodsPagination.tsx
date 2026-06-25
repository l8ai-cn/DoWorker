"use client";

import { Button } from "@/components/ui/button";

interface RunnerPodsPaginationProps {
  offset: number;
  limit: number;
  total: number;
  t: (key: string, params?: Record<string, string | number>) => string;
  onOffsetChange: (offset: number) => void;
}

export function RunnerPodsPagination({
  offset,
  limit,
  total,
  t,
  onOffsetChange,
}: RunnerPodsPaginationProps) {
  if (total <= limit) return null;

  return (
    <div className="flex items-center justify-between">
      <p className="text-sm text-muted-foreground">
        {t("runners.detail.showing", {
          from: offset + 1,
          to: Math.min(offset + limit, total),
          total,
        })}
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          disabled={offset === 0}
          onClick={() => onOffsetChange(Math.max(0, offset - limit))}
        >
          {t("common.previous")}
        </Button>
        <Button
          variant="outline"
          size="sm"
          disabled={offset + limit >= total}
          onClick={() => onOffsetChange(offset + limit)}
        >
          {t("common.next")}
        </Button>
      </div>
    </div>
  );
}
