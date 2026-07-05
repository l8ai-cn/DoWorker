"use client";

import type React from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { ArrowUpRight, BookOpen, Code2, KeyRound, Terminal } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { getApiBaseUrl } from "@/lib/env";

export default function ApiAccessPage() {
  const params = useParams();
  const orgSlug = String(params.org ?? "dev-org");
  const apiBase = `${getApiBaseUrl()}/api/v1`;
  const createPodPath = `${apiBase}/organizations/${orgSlug}/pods`;

  return (
    <div className="flex h-full min-h-0 flex-col bg-background">
      <PageHeader
        title="API 接入"
        subtitle="通过 API 启动和管理 Do Worker Pod"
      />

      <div className="flex-1 overflow-y-auto bg-surface-muted/25">
        <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-6 py-6 lg:px-8">
          <p className="text-sm leading-6 text-muted-foreground">
            统一入口{" "}
            <code className="rounded bg-background px-1.5 py-0.5 font-mono text-xs">{apiBase}</code>
            。使用用户 Token 或 API Key 认证后，可在集群上创建环境、传入 AgentFile Layer、绑定仓库。
          </p>

          <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            <InfoCard icon={<Terminal className="h-4 w-4" />} title="Base URL" value={apiBase} />
            <InfoCard
              icon={<KeyRound className="h-4 w-4" />}
              title="认证方式"
              value="Authorization: Bearer <token>"
            />
            <InfoCard icon={<BookOpen className="h-4 w-4" />} title="完整文档" value="/docs/api/pods" />
          </section>

          <Card className="surface-card">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <Code2 className="h-4 w-4 text-primary" />
                创建 Pod 示例
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <CodeBlock
                code={`curl -X POST "${createPodPath}" \\
  -H "Authorization: Bearer $DO_WORKER_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "agent_slug": "codex-cli",
    "agentfile_layer": "PROMPT \\"分析当前仓库并给出修复方案\\"\\nMODE pty",
    "perpetual": false,
    "cols": 120,
    "rows": 32
  }'`}
              />
              <p className="text-sm text-muted-foreground">
                Pod 专属配置统一走 AgentFile Layer；仓库、分支、Skill、运行时 Bundle
                都应以 Layer 或平台字段传入，避免在客户端拼接临时命令。
              </p>
            </CardContent>
          </Card>

          <Card className="surface-card">
            <CardHeader>
              <CardTitle className="text-base">常用入口</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-2 sm:grid-cols-2">
              <DocLink href="/docs/api">API 总览</DocLink>
              <DocLink href="/docs/api/pods">Pod 生命周期接口</DocLink>
              <DocLink href="/docs/concepts/agentfile">AgentFile 语法</DocLink>
              <DocLink href="/docs/concepts/agentfile-layer">AgentFile Layer</DocLink>
              <DocLink href={`/${orgSlug}/settings?scope=organization&tab=api-keys`}>
                API Key 管理
              </DocLink>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function InfoCard({
  icon,
  title,
  value,
}: {
  icon: React.ReactNode;
  title: string;
  value: string;
}) {
  return (
    <Card variant="default" className="surface-card">
      <CardContent className="flex items-start gap-3 p-4">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
          {icon}
        </div>
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{title}</p>
          <p className="mt-1 break-all font-mono text-xs text-foreground">{value}</p>
        </div>
      </CardContent>
    </Card>
  );
}

function CodeBlock({ code }: { code: string }) {
  return (
    <pre className="overflow-x-auto rounded-lg bg-background p-4 font-mono text-xs leading-5 text-foreground ring-1 ring-border/60">
      <code>{code}</code>
    </pre>
  );
}

function DocLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      className="group flex items-center justify-between rounded-lg border border-border bg-background px-3 py-2.5 text-sm text-foreground motion-interactive hover:border-primary/30 hover:bg-surface-muted"
    >
      <span>{children}</span>
      <ArrowUpRight className="h-4 w-4 text-muted-foreground group-hover:text-primary" />
    </Link>
  );
}
