import { createFileRoute, useRouter } from "@tanstack/react-router";
import { Eye, EyeOff, Loader2, Lock, Mail } from "lucide-react";
import { useState } from "react";
import { AgentCloudLogo } from "@/components/agent-cloud-logo";
import { MobileFrame } from "@/components/mobile-frame";
import { APP_NAME, APP_TAGLINE, pageTitle } from "@/lib/app-brand";
import { login } from "@/lib/auth-store";
import { cn } from "@/lib/utils";

const DEV_ACCOUNTS = import.meta.env.DEV
  ? [
      { label: "开发用户", email: "dev@agentcloud.local", password: "AdminAb123456" },
      { label: "管理员", email: "admin@agentcloud.local", password: "Ab123456" },
    ]
  : [];

export const Route = createFileRoute("/login")({
  validateSearch: (
    search: Record<string, unknown>,
  ): { workerPodKey?: string; workerTarget?: "preview" } => ({
    workerPodKey:
      typeof search.workerPodKey === "string" &&
      /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(search.workerPodKey)
        ? search.workerPodKey
        : undefined,
    workerTarget: search.workerTarget === "preview" ? "preview" : undefined,
  }),
  head: () => ({
    meta: [
      { title: pageTitle("登录") },
      { name: "description", content: `登录 ${APP_NAME}，连接 Codex / Claude Code 等 Agent。` },
    ],
  }),
  component: LoginPage,
});

function LoginPage() {
  const router = useRouter();
  const search = Route.useSearch();
  const [username, setUsername] = useState(DEV_ACCOUNTS[0]?.email ?? "");
  const [password, setPassword] = useState(DEV_ACCOUNTS[0]?.password ?? "");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await login(username.trim(), password);
      if (search.workerPodKey) {
        router.navigate({
          to: search.workerTarget === "preview" ? "/workers/$podKey/preview" : "/workers/$podKey",
          params: { podKey: search.workerPodKey },
          replace: true,
        });
      } else {
        router.navigate({ to: "/" });
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "登录失败");
    } finally {
      setBusy(false);
    }
  };

  return (
    <MobileFrame hideNav>
      <div className="relative flex min-h-screen flex-col overflow-hidden">
        <div className="pointer-events-none absolute -left-24 -top-24 h-64 w-64 rounded-full bg-primary/20 blur-3xl" />
        <div className="pointer-events-none absolute -bottom-16 -right-16 h-56 w-56 rounded-full bg-accent/25 blur-3xl" />
        <div className="pointer-events-none absolute inset-0 grid-bg opacity-40" />

        <div className="safe-top relative flex flex-1 flex-col justify-center px-6 py-10">
          <header className="mb-10 flex flex-col items-center text-center">
            <div className="relative mb-5">
              <div className="absolute -inset-3 rounded-[1.35rem] bg-primary/25 blur-xl" />
              <div className="relative flex h-16 w-16 items-center justify-center overflow-hidden rounded-2xl bg-primary/10 p-2.5 ring-1 ring-primary/35 shadow-lg">
                <AgentCloudLogo className="h-full w-full" />
              </div>
            </div>
            <h1 className="text-[1.65rem] font-semibold tracking-tight">{APP_NAME}</h1>
            <p className="mt-1.5 max-w-[16rem] text-[13px] leading-relaxed text-muted-foreground">
              {APP_TAGLINE}
              <span className="mt-1 block text-foreground/70">随时随地派发 Agent 任务</span>
            </p>
          </header>

          <form
            onSubmit={submit}
            className="rounded-2xl border border-border/70 bg-surface/80 p-5 shadow-xl backdrop-blur-xl ring-1 ring-white/5"
          >
            <p className="mb-4 text-center text-[13px] font-medium text-foreground/90">
              登录你的账号
            </p>

            <div className="space-y-3.5">
              <label className="block space-y-1.5">
                <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                  邮箱
                </span>
                <div className="relative">
                  <Mail className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <input
                    type="email"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    className="w-full rounded-xl border border-border/80 bg-background/60 py-3 pl-10 pr-3 text-sm outline-none transition focus:border-primary/50 focus:ring-2 focus:ring-primary/25"
                    autoComplete="username"
                    disabled={busy}
                  />
                </div>
              </label>

              <label className="block space-y-1.5">
                <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
                  密码
                </span>
                <div className="relative">
                  <Lock className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <input
                    type={showPassword ? "text" : "password"}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full rounded-xl border border-border/80 bg-background/60 py-3 pl-10 pr-11 text-sm outline-none transition focus:border-primary/50 focus:ring-2 focus:ring-primary/25"
                    autoComplete="current-password"
                    disabled={busy}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword((v) => !v)}
                    className="absolute right-2 top-1/2 flex h-8 w-8 -translate-y-1/2 items-center justify-center rounded-lg text-muted-foreground hover:bg-surface-2 hover:text-foreground"
                    aria-label={showPassword ? "隐藏密码" : "显示密码"}
                  >
                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </label>
            </div>

            {error && (
              <p className="mt-3 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-center text-xs text-destructive">
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={busy || !username.trim() || !password}
              className={cn(
                "mt-5 flex w-full items-center justify-center gap-2 rounded-xl bg-primary py-3 text-sm font-semibold text-primary-foreground transition active:scale-[0.98] disabled:opacity-50",
                !busy && "glow-primary",
              )}
            >
              {busy ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  登录中…
                </>
              ) : (
                "登录"
              )}
            </button>
          </form>

          {DEV_ACCOUNTS.length > 0 && (
            <div className="mt-6 rounded-xl border border-dashed border-border/60 bg-surface/40 px-3 py-3">
              <p className="text-center text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                本地开发账号
              </p>
              <div className="mt-2 flex flex-wrap justify-center gap-2">
                {DEV_ACCOUNTS.map((acct) => (
                  <button
                    key={acct.email}
                    type="button"
                    disabled={busy}
                    onClick={() => {
                      setUsername(acct.email);
                      setPassword(acct.password);
                      setError(null);
                    }}
                    className="rounded-full bg-surface px-3 py-1.5 text-[11px] font-medium text-foreground/80 ring-1 ring-border/50 transition hover:ring-primary/40 disabled:opacity-50"
                  >
                    {acct.label}
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </MobileFrame>
  );
}
