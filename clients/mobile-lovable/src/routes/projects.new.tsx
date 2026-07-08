import { Link, createFileRoute, useRouter } from "@tanstack/react-router";
import { ArrowLeft, FolderPlus } from "lucide-react";
import { useState } from "react";
import { MobileFrame } from "@/components/mobile-frame";
import { pageTitle } from "@/lib/app-brand";
import { readAuthToken } from "@/lib/auth-store";
import { projectIdFromName } from "@/lib/project-label";
import { saveLocalProject } from "@/lib/projects-local";

export const Route = createFileRoute("/projects/new")({
  head: () => ({ meta: [{ title: pageTitle("新建项目") }] }),
  component: NewProject,
});

const COLORS = ["primary", "accent", "info"] as const;

function NewProject() {
  const router = useRouter();
  const authed = Boolean(readAuthToken());
  const [name, setName] = useState("");
  const [repo, setRepo] = useState("");
  const [host, setHost] = useState("");
  const [color, setColor] = useState<(typeof COLORS)[number]>("primary");

  const submit = () => {
    if (!name.trim()) return;
    if (!authed) {
      router.navigate({ to: "/login" });
      return;
    }
    saveLocalProject({
      name: name.trim(),
      repo: repo.trim() || undefined,
      host: host.trim() || undefined,
      color,
    });
    router.navigate({ to: "/projects/$projectId", params: { projectId: projectIdFromName(name.trim()) } });
  };

  return (
    <MobileFrame>
      <div className="flex min-h-screen flex-col">
        <header className="safe-top sticky top-0 z-30 flex items-center gap-2 border-b border-border/60 bg-background/85 px-4 pb-3 pt-3 backdrop-blur-xl">
          <Link to="/" className="flex h-8 w-8 items-center justify-center rounded-full hover:bg-surface">
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <h1 className="flex-1 text-[14px] font-semibold">新建项目</h1>
        </header>

        <div className="flex-1 space-y-4 px-5 py-5">
          <p className="text-[12px] text-muted-foreground">
            项目名会作为会话标签（omni_project）同步到 Do Worker 服务端。首个任务创建后出现在项目列表。
          </p>
          <Field label="项目名称" required>
            <input
              autoFocus
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="例如：API Gateway"
              className="w-full bg-transparent text-[14px] outline-none placeholder:text-muted-foreground"
            />
          </Field>
          <Field label="仓库（可选，仅本地展示）">
            <input
              value={repo}
              onChange={(e) => setRepo(e.target.value)}
              placeholder="acme/api-gateway"
              className="w-full bg-transparent font-mono text-[13px] outline-none placeholder:text-muted-foreground"
            />
          </Field>
          <Field label="Runner 主机（可选，仅本地展示）">
            <input
              value={host}
              onChange={(e) => setHost(e.target.value)}
              placeholder="dev-runner-codex"
              className="w-full bg-transparent font-mono text-[13px] outline-none placeholder:text-muted-foreground"
            />
          </Field>
        </div>

        <div className="safe-bottom sticky bottom-0 border-t border-border/60 bg-background/95 px-5 pt-3 backdrop-blur-xl">
          <button
            onClick={submit}
            disabled={!name.trim()}
            className="flex w-full items-center justify-center gap-2 rounded-full bg-primary py-3.5 text-[14px] font-semibold text-primary-foreground glow-primary transition active:scale-[0.98] disabled:opacity-40"
          >
            <FolderPlus className="h-4 w-4" />
            创建项目
          </button>
        </div>
      </div>
    </MobileFrame>
  );
}

function Field({ label, required, children }: { label: string; required?: boolean; children: React.ReactNode }) {
  return (
    <label className="block rounded-2xl bg-card p-3 ring-1 ring-border/50 focus-within:ring-primary/50">
      <p className="mb-1 text-[10.5px] uppercase tracking-wider text-muted-foreground">
        {label}
        {required && <span className="text-warning"> *</span>}
      </p>
      {children}
    </label>
  );
}
