"use client";

import { usePathname } from "next/navigation";
import { Bell, Menu } from "lucide-react";
import { Button } from "@/components/ui/button";

const pageTitles: Record<string, string> = {
  "/": "仪表盘",
  "/users": "用户",
  "/organizations": "组织",
  "/sso": "单点登录",
  "/runners": "Runner",
  "/relays": "中继",
  "/promo-codes": "优惠码",
  "/support-tickets": "支持工单",
  "/audit-logs": "审计日志",
};

export function Header({ onMenuClick }: { onMenuClick?: () => void }) {
  const pathname = usePathname();

  let title = pageTitles[pathname];
  if (!title) {
    if (pathname.startsWith("/users/")) title = "用户详情";
    else if (pathname.startsWith("/organizations/")) title = "组织详情";
    else if (pathname.startsWith("/runners/")) title = "Runner 详情";
    else if (pathname.startsWith("/relays/")) title = "中继详情";
    else if (pathname.startsWith("/promo-codes/new")) title = "创建优惠码";
    else if (pathname.startsWith("/promo-codes/")) title = "优惠码详情";
    else if (pathname.startsWith("/support-tickets/")) title = "工单详情";
    else title = "管理控制台";
  }

  return (
    <header className="flex h-16 items-center justify-between border-b border-border bg-card px-4 md:px-6">
      <div className="flex items-center gap-2">
        {onMenuClick && (
          <Button
            variant="ghost"
            size="icon"
            className="md:hidden"
            onClick={onMenuClick}
          >
            <Menu className="h-5 w-5" />
            <span className="sr-only">打开菜单</span>
          </Button>
        )}
        <h1 className="text-xl font-semibold">{title}</h1>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="icon">
          <Bell className="h-5 w-5" />
        </Button>
      </div>
    </header>
  );
}
