"use client";

import { RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import type { ExpertMarketReleaseStatus } from "@/lib/api/admin";
import { cn } from "@/lib/utils";
import { ExpertReleaseDetail } from "./expert-release-detail";
import { ExpertReleaseList } from "./expert-release-list";
import { statusLabels } from "./expert-release-status";
import { useExpertMarketReview } from "./use-expert-market-review";

const statuses: ExpertMarketReleaseStatus[] = [
  "pending",
  "published",
  "rejected",
  "withdrawn",
];

export default function ExpertMarketPage() {
  const review = useExpertMarketReview();

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">专家市场审核</h1>
          <p className="text-sm text-muted-foreground">
            审核专家发布快照与运行依赖
          </p>
        </div>
        <Button
          variant="outline"
          onClick={review.refresh}
          disabled={review.isLoading}
        >
          <RefreshCw className={cn(review.isLoading && "animate-spin")} />
          刷新
        </Button>
      </header>

      <div className="flex w-fit max-w-full overflow-x-auto rounded-md border border-border p-1">
        {statuses.map((item) => (
          <Button
            key={item}
            size="sm"
            variant={review.status === item ? "secondary" : "ghost"}
            onClick={() => review.changeStatus(item)}
          >
            {statusLabels[item]}
          </Button>
        ))}
      </div>

      <ExpertReleaseList
        releases={review.releases}
        isLoading={review.isLoading}
        error={review.error}
        onRetry={review.loadReleases}
        onView={review.viewRelease}
      />

      {review.isDetailLoading && (
        <div className="rounded-md border border-border p-6 text-sm text-muted-foreground">
          正在加载发布详情...
        </div>
      )}
      {review.selected && (
        <section className="rounded-md border border-border bg-card p-4 md:p-6">
          <ExpertReleaseDetail
            release={review.selected}
            isActing={review.isActing}
            isRejecting={review.isRejecting}
            rejectionReason={review.rejectionReason}
            rejectionError={review.rejectionError}
            onApprove={review.approve}
            onRejectStart={review.startRejecting}
            onRejectCancel={review.cancelRejecting}
            onRejectConfirm={review.reject}
            onReasonChange={review.changeRejectionReason}
          />
        </section>
      )}
    </div>
  );
}
