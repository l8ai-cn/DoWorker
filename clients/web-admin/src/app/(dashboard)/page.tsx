"use client";

import { useState, useEffect } from "react";
import {
  Users,
  Building2,
  Server,
  Activity,
  UserPlus,
  TrendingUp,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getDashboardStats, DashboardStats } from "@/lib/api/admin";

function StatCard({
  title,
  value,
  subValue,
  icon: Icon,
  trend,
}: {
  title: string;
  value: number;
  subValue?: string;
  icon: React.ComponentType<{ className?: string }>;
  trend?: { value: number; label: string };
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value.toLocaleString()}</div>
        {subValue && (
          <p className="text-xs text-muted-foreground">{subValue}</p>
        )}
        {trend && (
          <div className="mt-2 flex items-center gap-1 text-xs">
            <TrendingUp className="h-3 w-3 text-success" />
            <span className="text-success">+{trend.value}</span>
            <span className="text-muted-foreground">{trend.label}</span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);
  const [refetchKey, setRefetchKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    getDashboardStats()
      .then((result) => {
        if (cancelled) return;
        setStats(result);
        setError(null);
        setIsLoading(false);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err);
        setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [refetchKey]);

  useEffect(() => {
    const interval = setInterval(() => setRefetchKey((k) => k + 1), 30000);
    return () => clearInterval(interval);
  }, []);

  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 8 }).map((_, i) => (
          <Card key={i} className="animate-pulse">
            <CardHeader className="pb-2">
              <div className="h-4 w-24 rounded bg-muted" />
            </CardHeader>
            <CardContent>
              <div className="h-8 w-16 rounded bg-muted" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (error || !stats) {
    return (
      <div className="flex h-64 items-center justify-center">
        <p className="text-muted-foreground">仪表盘数据加载失败</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="用户总数"
          value={stats.total_users}
          subValue={`${stats.active_users} 个活跃`}
          icon={Users}
          trend={{ value: stats.new_users_today, label: "今日新增" }}
        />
        <StatCard
          title="组织"
          value={stats.total_organizations}
          icon={Building2}
        />
        <StatCard
          title="Runner"
          value={stats.total_runners}
          subValue={`${stats.online_runners} 个在线`}
          icon={Server}
        />
        <StatCard
          title="活跃 Pod"
          value={stats.active_pods}
          subValue={`共 ${stats.total_pods} 个`}
          icon={Activity}
        />
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <UserPlus className="h-5 w-5" />
              新增用户
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <p className="text-2xl font-bold">{stats.new_users_today}</p>
                <p className="text-xs text-muted-foreground">今日</p>
              </div>
              <div>
                <p className="text-2xl font-bold">{stats.new_users_this_week}</p>
                <p className="text-xs text-muted-foreground">本周</p>
              </div>
              <div>
                <p className="text-2xl font-bold">{stats.new_users_this_month}</p>
                <p className="text-xs text-muted-foreground">本月</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>订阅</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-2xl font-bold">{stats.active_subscriptions}</p>
                <p className="text-xs text-muted-foreground">活跃</p>
              </div>
              <div className="text-right">
                <p className="text-2xl font-bold">{stats.total_subscriptions}</p>
                <p className="text-xs text-muted-foreground">总计</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>系统健康</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <div className="h-3 w-3 rounded-full bg-success" />
              <span className="text-sm">所有系统运行正常</span>
            </div>
            <p className="mt-2 text-xs text-muted-foreground">
              {stats.total_runners} 个 Runner 中有 {stats.online_runners} 个在线
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
