"use client";

import { useState, useEffect } from "react";
import { Search, ChevronLeft, ChevronRight } from "lucide-react";
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
  listSupportTickets,
  getSupportTicketStats,
  SupportTicket,
  SupportTicketStats,
} from "@/lib/api/admin";
import type { PaginatedResponse } from "@/lib/api/base";
import { TicketStatsCards } from "./ticket-stats-cards";
import { TicketsTable } from "./tickets-table";

const statusFilterLabels: Record<string, string> = {
  all: "全部状态",
  open: "待处理",
  in_progress: "处理中",
  resolved: "已解决",
  closed: "已关闭",
};

const categoryFilterLabels: Record<string, string> = {
  all: "全部分类",
  bug: "缺陷",
  feature_request: "功能请求",
  usage_question: "使用问题",
  account: "账号",
  other: "其他",
};

export default function SupportTicketsPage() {
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [categoryFilter, setCategoryFilter] = useState<string>("all");
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const [stats, setStats] = useState<SupportTicketStats | null>(null);
  const [data, setData] = useState<PaginatedResponse<SupportTicket> | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [statsKey, setStatsKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    listSupportTickets({
      search: search || undefined,
      status: statusFilter !== "all" ? statusFilter : undefined,
      category: categoryFilter !== "all" ? categoryFilter : undefined,
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
  }, [search, statusFilter, categoryFilter, page]);

  useEffect(() => {
    let cancelled = false;
    getSupportTicketStats()
      .then((result) => {
        if (!cancelled) setStats(result);
      })
      .catch(() => {
        // Stats are non-critical
      });
    return () => {
      cancelled = true;
    };
  }, [statsKey]);

  useEffect(() => {
    const interval = setInterval(() => setStatsKey((k) => k + 1), 30000);
    return () => clearInterval(interval);
  }, []);

  const total = data?.total || 0;
  const totalPages = data?.total_pages || 1;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">支持工单</h1>
        <p className="text-sm text-muted-foreground">
          管理并响应用户支持请求
        </p>
      </div>

      {stats && <TicketStatsCards stats={stats} />}

      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="搜索工单..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            className="pl-10"
          />
        </div>
        <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v); setPage(1); }}>
          <SelectTrigger className="w-40">
            <SelectValue
              placeholder="全部状态"
              displayValue={statusFilterLabels[statusFilter]}
            />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="open">待处理</SelectItem>
            <SelectItem value="in_progress">处理中</SelectItem>
            <SelectItem value="resolved">已解决</SelectItem>
            <SelectItem value="closed">已关闭</SelectItem>
          </SelectContent>
        </Select>
        <Select value={categoryFilter} onValueChange={(v) => { setCategoryFilter(v); setPage(1); }}>
          <SelectTrigger className="w-44">
            <SelectValue
              placeholder="全部分类"
              displayValue={categoryFilterLabels[categoryFilter]}
            />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部分类</SelectItem>
            <SelectItem value="bug">缺陷</SelectItem>
            <SelectItem value="feature_request">功能请求</SelectItem>
            <SelectItem value="usage_question">使用问题</SelectItem>
            <SelectItem value="account">账号</SelectItem>
            <SelectItem value="other">其他</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <TicketsTable tickets={data?.data || []} isLoading={isLoading} />

      {totalPages > 1 && (
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm text-muted-foreground">
            显示第 {(page - 1) * pageSize + 1} 到 {Math.min(page * pageSize, total)} 条，共 {total} 个工单
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
