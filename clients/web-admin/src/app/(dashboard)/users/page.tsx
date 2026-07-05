"use client";

import { useState, useEffect } from "react";
import { Search } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  listUsers,
  disableUser,
  enableUser,
  grantAdmin,
  revokeAdmin,
  verifyUserEmail,
  unverifyUserEmail,
  User,
} from "@/lib/api/admin";
import type { PaginatedResponse } from "@/lib/api/base";
import { UserRow } from "./user-row";

export default function UsersPage() {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [data, setData] = useState<PaginatedResponse<User> | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [refetchKey, setRefetchKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    listUsers({ search, page, page_size: 20 })
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

  const makeHandler = (action: (id: number) => Promise<unknown>, successMsg: string, errorMsg: string) => {
    return async (userId: number) => {
      try {
        await action(userId);
        toast.success(successMsg);
        setIsLoading(true);
        setRefetchKey((k) => k + 1);
      } catch (err: unknown) {
        toast.error((err as { error?: string })?.error || errorMsg);
      }
    };
  };

  const handleDisable = makeHandler(disableUser, "用户已停用", "停用用户失败");
  const handleEnable = makeHandler(enableUser, "用户已启用", "启用用户失败");
  const handleGrantAdmin = makeHandler(grantAdmin, "管理员权限已授予", "授予管理员权限失败");
  const handleRevokeAdmin = makeHandler(revokeAdmin, "管理员权限已撤销", "撤销管理员权限失败");
  const handleVerifyEmail = makeHandler(verifyUserEmail, "邮箱已验证", "验证邮箱失败");
  const handleUnverifyEmail = makeHandler(unverifyUserEmail, "邮箱验证已取消", "取消邮箱验证失败");

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <div className="relative flex-1 sm:max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="搜索用户..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            className="pl-9"
          />
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>用户 ({data?.total || 0})</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-16 animate-pulse rounded-lg bg-muted" />
              ))}
            </div>
          ) : (
            <div className="space-y-2">
              {data?.data.map((user) => (
                <UserRow
                  key={user.id}
                  user={user}
                  onDisable={() => handleDisable(user.id)}
                  onEnable={() => handleEnable(user.id)}
                  onGrantAdmin={() => handleGrantAdmin(user.id)}
                  onRevokeAdmin={() => handleRevokeAdmin(user.id)}
                  onVerifyEmail={() => handleVerifyEmail(user.id)}
                  onUnverifyEmail={() => handleUnverifyEmail(user.id)}
                />
              ))}
              {data?.data.length === 0 && (
                <p className="py-8 text-center text-muted-foreground">暂无用户</p>
              )}
            </div>
          )}

          {data && data.total_pages > 1 && (
            <div className="mt-4 flex items-center justify-between">
              <p className="text-sm text-muted-foreground">
                第 {data.page} / {data.total_pages} 页
              </p>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" disabled={page === 1} onClick={() => setPage(page - 1)}>
                  上一页
                </Button>
                <Button variant="outline" size="sm" disabled={page >= data.total_pages} onClick={() => setPage(page + 1)}>
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
