"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Users,
  Building2,
  Server,
  ScrollText,
  LogOut,
  Tag,
  Radio,
  MessageSquare,
  KeyRound,
  Store,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth";
import { DoWorkerMark } from "@/components/brand/DoWorkerMark";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetTitle,
} from "@/components/ui/sheet";

const navItems = [
  {
    title: "仪表盘",
    href: "/",
    icon: LayoutDashboard,
  },
  {
    title: "用户",
    href: "/users",
    icon: Users,
  },
  {
    title: "组织",
    href: "/organizations",
    icon: Building2,
  },
  {
    title: "单点登录",
    href: "/sso",
    icon: KeyRound,
  },
  {
    title: "Runner",
    href: "/runners",
    icon: Server,
  },
  {
    title: "中继",
    href: "/relays",
    icon: Radio,
  },
  {
    title: "优惠码",
    href: "/promo-codes",
    icon: Tag,
  },
  {
    title: "支持工单",
    href: "/support-tickets",
    icon: MessageSquare,
  },
  {
    title: "专家市场审核",
    href: "/expert-market",
    icon: Store,
  },
  {
    title: "审计日志",
    href: "/audit-logs",
    icon: ScrollText,
  },
];

export function SidebarContent({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname();
  const { user, logout } = useAuthStore();

  return (
    <>
      <div className="flex h-16 items-center gap-2 border-b border-border px-6">
        <span className="flex h-8 w-8 items-center justify-center overflow-hidden rounded-lg shadow-sm">
          <DoWorkerMark className="h-full w-full" />
        </span>
        <div className="min-w-0">
          <p className="truncate text-base font-semibold leading-tight">Do Worker</p>
          <p className="truncate text-xs text-muted-foreground">管理控制台</p>
        </div>
      </div>

      <nav className="flex-1 space-y-1 overflow-y-auto p-4">
        {navItems.map((item) => {
          const isActive = pathname === item.href ||
            (item.href !== "/" && pathname.startsWith(item.href));
          return (
            <Link
              key={item.href}
              href={item.href}
              onClick={onNavigate}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                isActive
                  ? "bg-primary/10 text-primary"
                  : "text-muted-foreground hover:bg-accent hover:text-foreground"
              )}
            >
              <item.icon className="h-5 w-5" />
              {item.title}
            </Link>
          );
        })}
      </nav>

      <div className="border-t border-border p-4">
        {user && (
          <div className="mb-3 flex items-center gap-3">
            {user.avatar_url ? (
              <img
                src={user.avatar_url}
                alt={user.username}
                className="h-8 w-8 rounded-full"
              />
            ) : (
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/20 text-sm font-medium text-primary">
                {user.username[0].toUpperCase()}
              </div>
            )}
            <div className="flex-1 truncate">
              <p className="text-sm font-medium truncate">{user.name || user.username}</p>
              <p className="text-xs text-muted-foreground truncate">{user.email}</p>
            </div>
          </div>
        )}
        <Button
          variant="ghost"
          className="w-full justify-start text-muted-foreground"
          onClick={logout}
        >
          <LogOut className="mr-2 h-4 w-4" />
          退出登录
        </Button>
      </div>
    </>
  );
}

export function Sidebar() {
  return (
    <aside className="hidden md:flex h-screen w-64 flex-col border-r border-border bg-card">
      <SidebarContent />
    </aside>
  );
}

export function MobileSidebar({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="left" className="p-0">
        <SheetTitle className="sr-only">导航</SheetTitle>
        <SidebarContent onNavigate={() => onOpenChange(false)} />
      </SheetContent>
    </Sheet>
  );
}
