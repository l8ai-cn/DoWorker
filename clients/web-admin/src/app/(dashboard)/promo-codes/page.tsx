"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Search, Plus, ChevronLeft, ChevronRight } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  listPromoCodes,
  activatePromoCode,
  deactivatePromoCode,
  deletePromoCode,
  PromoCode,
  PromoCodeType,
} from "@/lib/api/admin";
import type { PaginatedResponse } from "@/lib/api/base";
import { PromoCodesTable } from "./promo-codes-table";

const typeFilterLabels: Record<string, string> = {
  all: "全部类型",
  media: "媒体",
  partner: "合作伙伴",
  campaign: "活动",
  internal: "内部",
  referral: "推荐",
};

const statusFilterLabels: Record<string, string> = {
  all: "全部状态",
  active: "启用",
  inactive: "停用",
};

export default function PromoCodesPage() {
  const [search, setSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const [data, setData] = useState<PaginatedResponse<PromoCode> | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [refetchKey, setRefetchKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    listPromoCodes({
      search: search || undefined,
      type: typeFilter !== "all" ? (typeFilter as PromoCodeType) : undefined,
      is_active: statusFilter === "all" ? undefined : statusFilter === "active",
      page,
      page_size: pageSize,
    })
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
  }, [search, typeFilter, statusFilter, page, refetchKey]);

  const triggerRefetch = () => {
    setIsLoading(true);
    setRefetchKey((k) => k + 1);
  };

  const handleActivate = async (id: number) => {
    try {
      await activatePromoCode(id);
      toast.success("优惠码已启用");
      triggerRefetch();
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "启用优惠码失败");
    }
  };

  const handleDeactivate = async (id: number) => {
    try {
      await deactivatePromoCode(id);
      toast.success("优惠码已停用");
      triggerRefetch();
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "停用优惠码失败");
    }
  };

  const handleDelete = async (code: PromoCode) => {
    if (!confirm(`确定要删除 "${code.code}" 吗？此操作无法撤销。`)) return;
    try {
      await deletePromoCode(code.id);
      toast.success("优惠码已删除");
      triggerRefetch();
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "删除优惠码失败");
    }
  };

  const total = data?.total || 0;
  const totalPages = data?.total_pages || 1;

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">优惠码</h1>
          <p className="text-sm text-muted-foreground">
            管理订阅相关的优惠码。
          </p>
        </div>
        <Link href="/promo-codes/new">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            创建优惠码
          </Button>
        </Link>
      </div>

      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="按代码或名称搜索..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            className="pl-10"
          />
        </div>
        <Select value={typeFilter} onValueChange={(v) => { setTypeFilter(v); setPage(1); }}>
          <SelectTrigger className="w-40">
            <SelectValue placeholder="全部类型" displayValue={typeFilterLabels[typeFilter]} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部类型</SelectItem>
            <SelectItem value="media">媒体</SelectItem>
            <SelectItem value="partner">合作伙伴</SelectItem>
            <SelectItem value="campaign">活动</SelectItem>
            <SelectItem value="internal">内部</SelectItem>
            <SelectItem value="referral">推荐</SelectItem>
          </SelectContent>
        </Select>
        <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v); setPage(1); }}>
          <SelectTrigger className="w-40">
            <SelectValue placeholder="全部状态" displayValue={statusFilterLabels[statusFilter]} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="active">启用</SelectItem>
            <SelectItem value="inactive">停用</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <PromoCodesTable
        promoCodes={data?.data || []}
        isLoading={isLoading}
        onActivate={handleActivate}
        onDeactivate={handleDeactivate}
        onDelete={handleDelete}
      />

      {totalPages > 1 && (
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm text-muted-foreground">
            显示第 {(page - 1) * pageSize + 1} 到{" "}
            {Math.min(page * pageSize, total)} 条，共 {total} 个优惠码
          </p>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="icon" onClick={() => setPage(page - 1)} disabled={page <= 1}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="text-sm">第 {page} / {totalPages} 页</span>
            <Button variant="outline" size="icon" onClick={() => setPage(page + 1)} disabled={page >= totalPages}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
