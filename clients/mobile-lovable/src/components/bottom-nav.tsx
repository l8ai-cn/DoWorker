import { Link, useRouterState } from "@tanstack/react-router";
import { Home, Users, ShieldCheck, User, Plus, Zap, FolderPlus, X } from "lucide-react";
import { useEffect, useState } from "react";
import { useSessionsList } from "@/hooks/useSessionsList";
import { readAuthToken } from "@/lib/auth-store";

type NavItem = { to: "/" | "/experts" | "/approvals" | "/me"; label: string; icon: typeof Home };
const items: NavItem[] = [
  { to: "/", label: "首页", icon: Home },
  { to: "/experts", label: "专家库", icon: Users },
  { to: "/approvals", label: "审批", icon: ShieldCheck },
  { to: "/me", label: "我的", icon: User },
];

export function BottomNav() {
  const path = useRouterState({ select: (s) => s.location.pathname });
  return <BottomNavContent key={path} path={path} />;
}

function BottomNavContent({ path }: { path: string }) {
  const liveList = useSessionsList();
  const authed = Boolean(readAuthToken());
  const pending = authed ? liveList.pendingCount : 0;
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && setOpen(false);
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open]);

  return (
    <nav className="safe-bottom sticky bottom-0 z-40 mt-auto border-t border-border/60 bg-background/90 px-2 pt-1.5 pb-1 backdrop-blur-xl">
      {open && (
        <>
          <button
            aria-label="关闭"
            onClick={() => setOpen(false)}
            className="fixed inset-0 z-40 bg-background/40 backdrop-blur-[2px] stream-in"
          />
          <div className="absolute bottom-[calc(100%+8px)] left-1/2 z-50 w-60 -translate-x-1/2 overflow-hidden rounded-2xl border border-border/60 bg-card shadow-2xl stream-in">
            <div className="flex items-center justify-between px-3 py-2 text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
              新建
              <button
                onClick={() => setOpen(false)}
                className="rounded-full p-0.5 hover:bg-surface"
              >
                <X className="h-3 w-3" />
              </button>
            </div>
            <Link
              to="/new"
              onClick={() => setOpen(false)}
              className="flex items-start gap-3 border-t border-border/50 px-3 py-2.5 hover:bg-surface"
            >
              <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/15 text-primary">
                <Zap className="h-4 w-4" />
              </span>
              <div className="min-w-0 flex-1">
                <p className="text-[13px] font-semibold">新任务</p>
                <p className="text-[10.5px] text-muted-foreground">选择工具与专家，下发一次执行</p>
              </div>
            </Link>
            <Link
              to="/projects/new"
              onClick={() => setOpen(false)}
              className="flex items-start gap-3 border-t border-border/50 px-3 py-2.5 hover:bg-surface"
            >
              <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-accent/15 text-accent">
                <FolderPlus className="h-4 w-4" />
              </span>
              <div className="min-w-0 flex-1">
                <p className="text-[13px] font-semibold">新项目</p>
                <p className="text-[10.5px] text-muted-foreground">接入代码仓库或工作目录</p>
              </div>
            </Link>
          </div>
        </>
      )}
      <ul className="relative z-50 flex items-end justify-between">
        {items.slice(0, 2).map((it) => (
          <NavCell
            key={it.to}
            item={it}
            active={it.to === "/" ? path === "/" : path.startsWith(it.to)}
            pending={0}
          />
        ))}
        <li className="flex-1">
          <button
            onClick={() => setOpen((v) => !v)}
            aria-label="新建"
            aria-expanded={open}
            className={
              "mx-auto -mt-5 flex h-12 w-12 items-center justify-center rounded-2xl bg-primary text-primary-foreground shadow-lg ring-4 ring-background transition active:scale-95 " +
              (open ? "rotate-45" : "")
            }
          >
            <Plus className="h-5 w-5" />
          </button>
        </li>
        {items.slice(2).map((it) => (
          <NavCell
            key={it.to}
            item={it}
            active={path.startsWith(it.to)}
            pending={it.to === "/approvals" ? pending : 0}
          />
        ))}
      </ul>
    </nav>
  );
}

function NavCell({ item, active, pending }: { item: NavItem; active: boolean; pending: number }) {
  const { to, label, icon: Icon } = item;
  return (
    <li className="flex-1">
      <Link
        to={to}
        className={
          "relative mx-auto flex flex-col items-center gap-0.5 rounded-lg px-2 py-1 text-[10px] transition-colors " +
          (active ? "text-primary" : "text-muted-foreground hover:text-foreground")
        }
      >
        <Icon className="h-4 w-4" />
        <span className="leading-none">{label}</span>
        {pending > 0 && (
          <span className="absolute right-2 top-0 flex h-3.5 min-w-3.5 items-center justify-center rounded-full bg-warning px-1 text-[8px] font-bold text-primary-foreground">
            {pending}
          </span>
        )}
      </Link>
    </li>
  );
}
