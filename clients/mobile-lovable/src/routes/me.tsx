import { Link, createFileRoute, useRouter } from "@tanstack/react-router";
import {
  ArrowLeft,
  Bell,
  ChevronRight,
  Github,
  KeyRound,
  LifeBuoy,
  LogOut,
  Moon,
  Palette,
  ShieldCheck,
  Sparkles,
  Sun,
} from "lucide-react";
import { useEffect, useState } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { useSessionsList } from "@/hooks/useSessionsList";
import { APP_NAME, pageTitle } from "@/lib/app-brand";
import { logout, restoreAuthIdentity } from "@/lib/auth-store";
import { resetNotificationsForLogout } from "@/lib/notifications";

export const Route = createFileRoute("/me")({
  head: () => ({
    meta: [
      { title: pageTitle("个人中心") },
      { name: "description", content: "账号信息、偏好设置与使用统计。" },
    ],
  }),
  component: MePage,
});

function MePage() {
  const router = useRouter();
  const liveList = useSessionsList();
  const [identity, setIdentity] = useState({
    authenticated: false,
    email: null as string | null,
    orgSlug: null as string | null,
  });
  useEffect(() => {
    void restoreAuthIdentity().then(setIdentity);
  }, []);
  const authed = identity.authenticated;
  const email = identity.email ?? "未登录";
  const org = identity.orgSlug;

  const liveActive = liveList.items.filter(
    (s) => s.status === "running" || s.pendingApprovals > 0,
  ).length;
  const liveDone = liveList.items.filter(
    (s) => s.status === "idle" && s.pendingApprovals === 0,
  ).length;

  const done = authed ? liveDone : 0;
  const active = authed ? liveActive : 0;
  const totalSessions = authed ? liveList.items.length : 0;

  const displayName = email.includes("@") ? email.split("@")[0] : email;
  const initial = displayName.charAt(0).toUpperCase() || "A";

  // theme toggle (client-only)
  const [dark, setDark] = useState(
    () => typeof document !== "undefined" && document.documentElement.classList.contains("dark"),
  );
  const toggleDark = () => {
    const next = !dark;
    setDark(next);
    document.documentElement.classList.toggle("dark", next);
  };

  const [notify, setNotify] = useState(true);

  const handleLogout = async () => {
    await logout();
    resetNotificationsForLogout();
    router.navigate({ to: "/login" });
  };

  return (
    <MobileFrame>
      <div className="flex min-h-screen flex-col">
        <header className="safe-top sticky top-0 z-30 flex items-center gap-2 border-b border-border/60 bg-background/85 px-4 pb-3 pt-3 backdrop-blur-xl">
          <Link
            to="/"
            className="flex h-8 w-8 items-center justify-center rounded-full hover:bg-surface"
          >
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div className="flex-1">
            <h1 className="text-[14px] font-semibold">个人中心</h1>
            <p className="text-[10.5px] text-muted-foreground">账号与偏好</p>
          </div>
        </header>

        <div className="flex-1 space-y-5 px-5 pb-24 pt-4">
          {!authed && (
            <div className="rounded-2xl border border-border/60 bg-surface/40 px-4 py-3 text-center text-[12px] text-muted-foreground">
              <Link to="/login" className="font-medium text-primary">
                登录 {APP_NAME}
              </Link>{" "}
              以同步真实数据
            </div>
          )}

          {/* Profile card */}
          <section className="relative overflow-hidden rounded-3xl border border-border/60 bg-gradient-to-br from-primary/10 via-card to-card p-5">
            <div className="absolute -right-8 -top-8 h-32 w-32 rounded-full bg-primary/20 blur-3xl" />
            <div className="relative flex items-center gap-4">
              <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary text-lg font-bold text-primary-foreground shadow-lg">
                {initial}
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-[15px] font-semibold capitalize">{displayName}</p>
                <p className="truncate text-[11.5px] text-muted-foreground">{email}</p>
                {org && (
                  <span className="mt-1 inline-flex items-center gap-1 rounded-full bg-primary/15 px-2 py-0.5 text-[10px] font-medium text-primary">
                    <Sparkles className="h-2.5 w-2.5" /> {org}
                  </span>
                )}
              </div>
            </div>

            <div className="relative mt-4 grid grid-cols-3 gap-2 rounded-2xl border border-border/50 bg-background/40 p-2">
              <Stat label="进行中" value={active} />
              <Stat label="已完成" value={done} />
              <Stat label="总会话" value={totalSessions} />
            </div>
          </section>

          {/* Preferences */}
          <Group title="偏好设置">
            <ToggleRow
              icon={dark ? Moon : Sun}
              label="深色模式"
              hint={dark ? "已开启" : "跟随系统"}
              checked={dark}
              onChange={toggleDark}
            />
            <ToggleRow
              icon={Bell}
              label="通知提醒"
              hint="审批 / 任务完成"
              checked={notify}
              onChange={() => setNotify((v) => !v)}
            />
            <LinkRow icon={Palette} label="外观与主题" hint="配色 · 字号" />
          </Group>

          {/* Account */}
          <Group title="账号">
            <LinkRow icon={KeyRound} label="API Key" hint="用于本地客户端接入" />
            <LinkRow icon={ShieldCheck} label="安全与权限" hint="会话审批策略" />
            <LinkRow icon={Github} label="绑定 GitHub" hint="未绑定" />
          </Group>

          {/* Help */}
          <Group title="更多">
            <LinkRow icon={LifeBuoy} label="帮助与反馈" />
            <button
              type="button"
              onClick={() => void handleLogout()}
              disabled={!authed}
              className="flex w-full items-center gap-3 px-4 py-3 text-left text-[13px] text-destructive hover:bg-destructive/5 disabled:opacity-40"
            >
              <LogOut className="h-4 w-4" />
              <span className="flex-1 font-medium">退出登录</span>
            </button>
          </Group>

          <p className="pt-2 text-center text-[10.5px] text-muted-foreground/70">
            {APP_NAME} · v0.1.0
          </p>
        </div>
      </div>
    </MobileFrame>
  );
}

function Stat({
  label,
  value,
  suffix,
}: {
  label: string;
  value: number | string;
  suffix?: string;
}) {
  return (
    <div className="flex flex-col items-center gap-0.5 py-1">
      <p className="text-[15px] font-bold tabular-nums">
        {value}
        {suffix && (
          <span className="ml-0.5 text-[10px] font-normal text-muted-foreground">{suffix}</span>
        )}
      </p>
      <p className="text-[10px] text-muted-foreground">{label}</p>
    </div>
  );
}

function Group({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section>
      <h2 className="mb-2 px-1 text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">
        {title}
      </h2>
      <div className="divide-y divide-border/60 overflow-hidden rounded-2xl border border-border/60 bg-card">
        {children}
      </div>
    </section>
  );
}

function LinkRow({ icon: Icon, label, hint }: { icon: typeof Bell; label: string; hint?: string }) {
  return (
    <button
      type="button"
      className="flex w-full items-center gap-3 px-4 py-3 text-left hover:bg-surface/60"
    >
      <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-surface text-muted-foreground">
        <Icon className="h-4 w-4" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="text-[13px] font-medium">{label}</p>
        {hint && <p className="truncate text-[10.5px] text-muted-foreground">{hint}</p>}
      </div>
      <ChevronRight className="h-4 w-4 text-muted-foreground/60" />
    </button>
  );
}

function ToggleRow({
  icon: Icon,
  label,
  hint,
  checked,
  onChange,
}: {
  icon: typeof Bell;
  label: string;
  hint?: string;
  checked: boolean;
  onChange: () => void;
}) {
  return (
    <div className="flex w-full items-center gap-3 px-4 py-3">
      <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-surface text-muted-foreground">
        <Icon className="h-4 w-4" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="text-[13px] font-medium">{label}</p>
        {hint && <p className="truncate text-[10.5px] text-muted-foreground">{hint}</p>}
      </div>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={onChange}
        className={
          "relative h-5 w-9 shrink-0 rounded-full transition-colors " +
          (checked ? "bg-primary" : "bg-surface-2")
        }
      >
        <span
          className={
            "absolute top-0.5 h-4 w-4 rounded-full bg-background shadow transition-transform " +
            (checked ? "translate-x-[18px]" : "translate-x-0.5")
          }
        />
      </button>
    </div>
  );
}
