"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { listAuditLogs, AuditLog } from "@/lib/api/admin";
import type { PaginatedResponse } from "@/lib/api/base";
import { formatDate } from "@/lib/utils";

const actionColors: Record<string, "default" | "secondary" | "destructive" | "outline" | "success" | "warning"> = {
  "user.view": "secondary",
  "user.update": "default",
  "user.disable": "destructive",
  "user.enable": "success",
  "user.grant_admin": "warning",
  "user.revoke_admin": "warning",
  "organization.view": "secondary",
  "organization.update": "default",
  "organization.delete": "destructive",
  "runner.view": "secondary",
  "runner.disable": "destructive",
  "runner.enable": "success",
  "runner.delete": "destructive",
};

export default function AuditLogsPage() {
  const [page, setPage] = useState(1);
  const [targetType, setTargetType] = useState<string | undefined>();
  const [data, setData] = useState<PaginatedResponse<AuditLog> | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    listAuditLogs({ page, page_size: 50, target_type: targetType })
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
  }, [page, targetType]);

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-2 sm:gap-4">
        <div className="flex flex-wrap gap-2">
          <Button
            variant={!targetType ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setTargetType(undefined);
              setPage(1);
            }}
          >
            全部
          </Button>
          <Button
            variant={targetType === "user" ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setTargetType("user");
              setPage(1);
            }}
          >
            用户
          </Button>
          <Button
            variant={targetType === "organization" ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setTargetType("organization");
              setPage(1);
            }}
          >
            组织
          </Button>
          <Button
            variant={targetType === "runner" ? "default" : "outline"}
            size="sm"
            onClick={() => {
              setTargetType("runner");
              setPage(1);
            }}
          >
            Runner
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>审计日志 ({data?.total || 0})</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 10 }).map((_, i) => (
                <div key={i} className="h-12 animate-pulse rounded-lg bg-muted" />
              ))}
            </div>
          ) : (
            <div className="space-y-2">
              {data?.data.map((log) => (
                <AuditLogRow key={log.id} log={log} />
              ))}
              {data?.data.length === 0 && (
                <p className="py-8 text-center text-muted-foreground">
                  暂无审计日志
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

function AuditLogRow({ log }: { log: AuditLog }) {
  const actionColor = actionColors[log.action] || "secondary";

  return (
    <div className="flex flex-col gap-3 rounded-lg border border-border p-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-center gap-4">
        <div className="flex flex-col gap-1">
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant={actionColor}>{log.action}</Badge>
            <span className="text-sm text-muted-foreground">
              {log.target_type} #{log.target_id}
            </span>
          </div>
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            {log.admin_user && (
              <span>由 {log.admin_user.name || log.admin_user.username} 操作</span>
            )}
            {log.ip_address && <span>来自 {log.ip_address}</span>}
          </div>
        </div>
      </div>
      <div className="hidden text-right text-xs text-muted-foreground sm:block">
        {formatDate(log.created_at)}
      </div>
    </div>
  );
}
