import { Eye } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { ExpertMarketRelease } from "@/lib/api/admin";
import { ExpertReleaseStatus } from "./expert-release-status";

interface ExpertReleaseListProps {
  releases: ExpertMarketRelease[];
  isLoading: boolean;
  error: string | null;
  onRetry: () => void;
  onView: (releaseId: number) => void;
}

export function ExpertReleaseList({
  releases,
  isLoading,
  error,
  onRetry,
  onView,
}: ExpertReleaseListProps) {
  if (isLoading) {
    return (
      <div className="rounded-md border border-border py-16 text-center text-sm text-muted-foreground">
        正在加载审核记录...
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-md border border-destructive/40 py-16 text-center">
        <p className="font-medium text-destructive">审核记录加载失败</p>
        <p className="mt-1 text-sm text-muted-foreground">{error}</p>
        <Button className="mt-4" variant="outline" onClick={onRetry}>
          重新加载
        </Button>
      </div>
    );
  }

  if (releases.length === 0) {
    return (
      <div className="rounded-md border border-border py-16 text-center text-sm text-muted-foreground">
        当前状态下暂无发布记录
      </div>
    );
  }

  return (
    <div className="rounded-md border border-border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>专家</TableHead>
            <TableHead>分类</TableHead>
            <TableHead>版本</TableHead>
            <TableHead>状态</TableHead>
            <TableHead>提交时间</TableHead>
            <TableHead className="w-24 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {releases.map((release) => (
            <TableRow key={release.id}>
              <TableCell>
                <p className="font-medium">{release.name}</p>
                <p className="max-w-md truncate text-xs text-muted-foreground">
                  {release.summary}
                </p>
              </TableCell>
              <TableCell>{release.category || "-"}</TableCell>
              <TableCell>v{release.version}</TableCell>
              <TableCell>
                <ExpertReleaseStatus status={release.status} />
              </TableCell>
              <TableCell>{formatDate(release.submitted_at)}</TableCell>
              <TableCell className="text-right">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onView(release.id)}
                >
                  <Eye />
                  查看详情
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function formatDate(value?: string): string {
  if (!value) return "-";
  return new Intl.DateTimeFormat("zh-CN", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
