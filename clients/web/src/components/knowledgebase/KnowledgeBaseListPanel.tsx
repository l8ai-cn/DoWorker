"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { BookOpen, Plus, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ConfirmDialog, useConfirmDialog } from "@/components/ui/confirm-dialog";
import { EmptyState } from "@/components/ui/empty-state";
import { PageHeader } from "@/components/ui/page-header";
import { CenteredSpinner } from "@/components/ui/spinner";
import {
  deleteKnowledgeBase,
  listKnowledgeBases,
  type KnowledgeBase,
} from "@/lib/api/facade/knowledgeBaseApi";
import { CreateKnowledgeBaseDialog } from "./CreateKnowledgeBaseDialog";

const SOURCE_LABELS: Record<string, string> = {
  git: "Git",
  feishu: "飞书",
  dingtalk: "钉钉",
  google: "Google",
};

export function KnowledgeBaseListPanel() {
  const params = useParams();
  const orgSlug = String(params.org ?? "");
  const [kbs, setKbs] = useState<KnowledgeBase[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const { dialogProps, confirm } = useConfirmDialog();

  const reload = useCallback(() => {
    listKnowledgeBases(orgSlug)
      .then((items) => {
        setKbs(items);
        setError(null);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "加载知识库失败"))
      .finally(() => setLoading(false));
  }, [orgSlug]);

  useEffect(() => {
    reload();
  }, [reload]);

  const handleDelete = async (kb: KnowledgeBase) => {
    const ok = await confirm({
      title: `删除知识库 ${kb.name}？`,
      description: "将同时删除内部 Git 仓库及其全部历史，不可恢复。",
      confirmText: "删除",
      variant: "destructive",
    });
    if (!ok) return;
    try {
      await deleteKnowledgeBase(orgSlug, kb.slug);
      setKbs((prev) => prev.filter((k) => k.slug !== kb.slug));
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除知识库失败");
    }
  };

  return (
    <div className="flex h-full flex-col">
      <PageHeader
        title="知识库"
        subtitle="Git 为底座的 llm-wiki 知识库，可只读或读写挂载到 Agent Pod"
        actions={
          <Button size="sm" onClick={() => setShowCreate(true)}>
            <Plus className="mr-1 h-4 w-4" />
            新建知识库
          </Button>
        }
      />

      <div className="flex-1 overflow-auto p-6">
        {error && <p className="mb-4 text-sm text-destructive">{error}</p>}
        {loading ? (
          <CenteredSpinner />
        ) : kbs.length === 0 ? (
          <EmptyState
            icon={<BookOpen className="h-full w-full" />}
            title="还没有知识库"
            description="创建后系统会初始化 llms.txt 索引、AGENTS.md 维护规范以及 raw/ 与 wiki/ 目录。"
            actions={
              <Button onClick={() => setShowCreate(true)}>
                <Plus className="mr-1 h-4 w-4" />
                新建知识库
              </Button>
            }
          />
        ) : (
          <div className="grid max-w-6xl grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {kbs.map((kb) => (
              <Link
                key={kb.slug}
                href={`/${orgSlug}/knowledge-base/${kb.slug}`}
                className="surface-card group flex flex-col gap-2 rounded-lg border border-border p-4 transition-shadow hover:shadow-md"
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="flex min-w-0 items-center gap-2">
                    <BookOpen className="h-4 w-4 shrink-0 text-primary" />
                    <span className="truncate font-medium" title={kb.name}>
                      {kb.name}
                    </span>
                  </div>
                  <button
                    type="button"
                    className="invisible text-muted-foreground hover:text-destructive group-hover:visible"
                    onClick={(e) => {
                      e.preventDefault();
                      void handleDelete(kb);
                    }}
                    aria-label="删除知识库"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
                <p className="line-clamp-2 min-h-[2.5rem] text-sm text-muted-foreground">
                  {kb.description || "暂无描述"}
                </p>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Badge variant="secondary">{SOURCE_LABELS[kb.source_type] ?? kb.source_type}</Badge>
                  <span className="font-mono">{kb.slug}</span>
                  {kb.sync_status && kb.sync_status !== "idle" && (
                    <Badge variant={kb.sync_status === "error" ? "destructive" : "outline"}>
                      {kb.sync_status}
                    </Badge>
                  )}
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>

      <CreateKnowledgeBaseDialog
        orgSlug={orgSlug}
        open={showCreate}
        onOpenChange={setShowCreate}
        onCreated={reload}
      />
      <ConfirmDialog {...dialogProps} />
    </div>
  );
}
