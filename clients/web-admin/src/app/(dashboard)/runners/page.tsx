"use client";

import { useState, useEffect } from "react";
import { Search } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  listRunners,
  disableRunner,
  enableRunner,
  deleteRunner,
  Runner,
} from "@/lib/api/admin";
import type { PaginatedResponse } from "@/lib/api/base";
import { RunnerRow } from "./_components/runner-row";

export default function RunnersPage() {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [data, setData] = useState<PaginatedResponse<Runner> | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [refetchKey, setRefetchKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    listRunners({ search, page, page_size: 20 })
      .then((result) => {
        if (cancelled) return;
        setData(result);
        setIsLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [search, page, refetchKey]);

  const triggerRefetch = () => {
    setIsLoading(true);
    setRefetchKey((k) => k + 1);
  };

  const handleDisable = async (runnerId: number) => {
    try {
      await disableRunner(runnerId);
      toast.success("Runner 已停用");
      triggerRefetch();
    } catch (err: unknown) {
      const message = (err as { error?: string })?.error || "停用 Runner 失败";
      toast.error(message);
    }
  };

  const handleEnable = async (runnerId: number) => {
    try {
      await enableRunner(runnerId);
      toast.success("Runner 已启用");
      triggerRefetch();
    } catch (err: unknown) {
      const message = (err as { error?: string })?.error || "启用 Runner 失败";
      toast.error(message);
    }
  };

  const handleDelete = async (runner: Runner) => {
    if (!confirm(`确定要删除 Runner "${runner.node_id}" 吗？此操作无法撤销。`)) {
      return;
    }
    try {
      await deleteRunner(runner.id);
      toast.success("Runner 已删除");
      triggerRefetch();
    } catch (err: unknown) {
      const message = (err as { error?: string })?.error || "删除 Runner 失败";
      toast.error(message);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <div className="relative flex-1 sm:max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="搜索 Runner..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            className="pl-9"
          />
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Runner ({data?.total || 0})</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-20 animate-pulse rounded-lg bg-muted" />
              ))}
            </div>
          ) : (
            <div className="space-y-2">
              {data?.data.map((runner) => (
                <RunnerRow
                  key={runner.id}
                  runner={runner}
                  onDisable={() => handleDisable(runner.id)}
                  onEnable={() => handleEnable(runner.id)}
                  onDelete={() => handleDelete(runner)}
                />
              ))}
              {data?.data.length === 0 && (
                <p className="py-8 text-center text-muted-foreground">
                  暂无 Runner
                </p>
              )}
            </div>
          )}

          {data && data.total_pages > 1 && (
            <div className="mt-4 flex items-center justify-between">
              <p className="text-sm text-muted-foreground">
                第 {data.page} / {data.total_pages} 页
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page === 1}
                  onClick={() => setPage(page - 1)}
                >
                  上一页
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= data.total_pages}
                  onClick={() => setPage(page + 1)}
                >
                  下一页
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
