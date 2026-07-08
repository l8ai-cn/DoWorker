"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Check, Copy, Sparkles } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/ui/page-header";
import { CenteredSpinner } from "@/components/ui/spinner";
import {
  getKnowledgeBase,
  type KnowledgeBase,
} from "@/lib/api/facade/knowledgeBaseApi";
import { usePodCreationStore } from "@/stores/podCreation";
import { KbFileTree } from "./KbFileTree";
import { KbFileViewer } from "./KbFileViewer";
import { KnowledgeBaseSourceSettings } from "./KnowledgeBaseSourceSettings";
import { SOURCE_LABELS, SYNC_STATUS_LABELS, syncStatusVariant } from "./sourceConfig";

export function KnowledgeBaseDetailPanel() {
  const params = useParams();
  const router = useRouter();
  const orgSlug = String(params.org ?? "");
  const kbSlug = String(params.slug ?? "");

  const [kb, setKb] = useState<KnowledgeBase | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [selectedPath, setSelectedPath] = useState<string | null>("llms.txt");
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let cancelled = false;
    getKnowledgeBase(orgSlug, kbSlug)
      .then((item) => {
        if (!cancelled) setKb(item);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "加载知识库失败");
      });
    return () => {
      cancelled = true;
    };
  }, [orgSlug, kbSlug]);

  const copyCloneUrl = async () => {
    if (!kb) return;
    await navigator.clipboard.writeText(kb.http_clone_url);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  if (error) {
    return (
      <div className="p-6">
        <p className="text-sm text-destructive">{error}</p>
      </div>
    );
  }
  if (!kb) return <CenteredSpinner />;

  return (
    <div className="flex h-full flex-col">
      <PageHeader
        title={
          <span className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => router.push(`/${orgSlug}/knowledge-base`)}
              className="text-muted-foreground hover:text-foreground"
              aria-label="返回知识库列表"
            >
              <ArrowLeft className="h-4 w-4" />
            </button>
            {kb.name}
          </span>
        }
        subtitle={kb.description || undefined}
        actions={
          <>
            <Button variant="outline" size="sm" onClick={copyCloneUrl}>
              {copied ? <Check className="mr-1 h-3.5 w-3.5" /> : <Copy className="mr-1 h-3.5 w-3.5" />}
              克隆地址
            </Button>
            <Button
              size="sm"
              onClick={() => {
                // Seed the create-pod form (it initializes from last choices)
                // with an rw mount before the workspace opens the modal.
                usePodCreationStore.getState().setLastChoices({
                  lastKnowledgeMounts: [{ slug: kb.slug, mode: "rw" }],
                });
                router.push(`/${orgSlug}/workspace?ingest_kb=${kb.slug}`);
              }}
              title="创建一个 rw 挂载本知识库的 Pod，把 raw/ 新资料编译进 wiki/"
            >
              <Sparkles className="mr-1 h-3.5 w-3.5" />
              Ingest
            </Button>
          </>
        }
      />

      <div className="flex items-center gap-2 border-b border-border px-6 py-2 text-xs text-muted-foreground">
        <span className="font-mono">{kb.slug}</span>
        <Badge variant="outline">{kb.default_branch}</Badge>
        <Badge variant="secondary">{SOURCE_LABELS[kb.source_type] ?? kb.source_type}</Badge>
        {kb.sync_status && kb.sync_status !== "idle" && (
          <Badge variant={syncStatusVariant(kb.sync_status)}>
            {SYNC_STATUS_LABELS[kb.sync_status] ?? kb.sync_status}
          </Badge>
        )}
        {kb.last_synced_at && <span>上次同步：{kb.last_synced_at}</span>}
      </div>

      <KnowledgeBaseSourceSettings orgSlug={orgSlug} kb={kb} onUpdated={setKb} />

      <div className="flex min-h-0 flex-1">
        <aside className="w-64 shrink-0 overflow-auto border-r border-border">
          <QuickNav selectedPath={selectedPath} onSelect={setSelectedPath} />
          <KbFileTree
            orgSlug={orgSlug}
            kbSlug={kbSlug}
            selectedPath={selectedPath}
            onSelectFile={setSelectedPath}
          />
        </aside>
        <main className="min-w-0 flex-1">
          <KbFileViewer orgSlug={orgSlug} kbSlug={kbSlug} path={selectedPath} />
        </main>
      </div>
    </div>
  );
}

const QUICK_NAV = [
  { path: "llms.txt", label: "llms.txt 索引" },
  { path: "AGENTS.md", label: "AGENTS.md 规范" },
  { path: "wiki/index.md", label: "wiki 总览" },
  { path: "wiki/log.md", label: "变更日志" },
];

function QuickNav({
  selectedPath,
  onSelect,
}: {
  selectedPath: string | null;
  onSelect: (path: string) => void;
}) {
  return (
    <div className="border-b border-border px-2 py-2">
      <p className="px-2 pb-1 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
        快速导航
      </p>
      {QUICK_NAV.map((item) => (
        <button
          key={item.path}
          type="button"
          onClick={() => onSelect(item.path)}
          className={`block w-full rounded px-2 py-1 text-left text-sm hover:bg-surface-muted ${
            selectedPath === item.path ? "bg-primary/10 text-primary" : ""
          }`}
        >
          {item.label}
        </button>
      ))}
    </div>
  );
}
