import { ChevronLeft, ChevronRight } from "lucide-react";

import { Button } from "@/components/ui/button";

interface ExpertReleasePaginationProps {
  total: number;
  limit: number;
  offset: number;
  isLoading: boolean;
  onPrevious: () => void;
  onNext: () => void;
}

export function ExpertReleasePagination({
  total,
  limit,
  offset,
  isLoading,
  onPrevious,
  onNext,
}: ExpertReleasePaginationProps) {
  if (total === 0) return null;

  const totalPages = Math.max(1, Math.ceil(total / limit));
  const currentPage = Math.floor(offset / limit) + 1;
  const firstItem = offset + 1;
  const lastItem = Math.min(offset + limit, total);

  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
      <p className="text-sm text-muted-foreground">
        显示第 {firstItem} 到 {lastItem} 条，共 {total} 条
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="icon"
          aria-label="上一页"
          title="上一页"
          disabled={isLoading || offset === 0}
          onClick={onPrevious}
        >
          <ChevronLeft />
        </Button>
        <span className="min-w-24 text-center text-sm">
          第 {currentPage} / {totalPages} 页
        </span>
        <Button
          variant="outline"
          size="icon"
          aria-label="下一页"
          title="下一页"
          disabled={isLoading || offset + limit >= total}
          onClick={onNext}
        >
          <ChevronRight />
        </Button>
      </div>
    </div>
  );
}
