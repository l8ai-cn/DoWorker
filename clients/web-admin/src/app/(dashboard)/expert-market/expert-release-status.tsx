import { Badge } from "@/components/ui/badge";
import type { ExpertMarketReleaseStatus } from "@/lib/api/admin";

const statusLabels: Record<ExpertMarketReleaseStatus, string> = {
  pending: "待审核",
  published: "已发布",
  rejected: "已驳回",
  withdrawn: "已撤回",
};

const statusVariants = {
  pending: "warning",
  published: "success",
  rejected: "destructive",
  withdrawn: "secondary",
} as const;

export function ExpertReleaseStatus({
  status,
}: {
  status: ExpertMarketReleaseStatus;
}) {
  return (
    <Badge variant={statusVariants[status]}>
      {statusLabels[status]}
    </Badge>
  );
}

export { statusLabels };
