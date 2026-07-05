"use client";

import { useEffect, useState } from "react";
import { FileText } from "lucide-react";

import { EmptyState } from "@/components/ui/empty-state";
import { Markdown } from "@/components/ui/markdown";
import { CenteredSpinner } from "@/components/ui/spinner";
import { getKbFile } from "@/lib/api/facade/knowledgeBaseApi";

interface KbFileViewerProps {
  orgSlug: string;
  kbSlug: string;
  path: string | null;
}

// Loaded results carry the path they belong to so loading/error states are
// derived by comparison instead of being reset synchronously in the effect.
export function KbFileViewer({ orgSlug, kbSlug, path }: KbFileViewerProps) {
  const [file, setFile] = useState<{ path: string; content: string } | null>(null);
  const [error, setError] = useState<{ path: string; message: string } | null>(null);

  useEffect(() => {
    if (!path) return;
    let cancelled = false;
    getKbFile(orgSlug, kbSlug, path)
      .then((f) => {
        if (!cancelled) setFile({ path, content: f.content });
      })
      .catch((err) => {
        if (!cancelled) {
          setError({ path, message: err instanceof Error ? err.message : "读取文件失败" });
        }
      });
    return () => {
      cancelled = true;
    };
  }, [orgSlug, kbSlug, path]);

  if (!path) {
    return (
      <EmptyState
        size="compact"
        icon={<FileText className="h-full w-full" />}
        title="选择左侧文件查看内容"
        description="建议从 llms.txt 索引开始导航。"
      />
    );
  }
  if (error?.path === path) {
    return <p className="p-4 text-sm text-destructive">{error.message}</p>;
  }
  if (file?.path !== path) return <CenteredSpinner />;

  const isMarkdown = path.endsWith(".md") || path.endsWith(".markdown");
  return (
    <div className="h-full overflow-auto p-4">
      <div className="mb-3 border-b border-border pb-2 font-mono text-xs text-muted-foreground">
        {path}
      </div>
      {isMarkdown ? (
        <Markdown content={file.content} />
      ) : (
        <pre className="whitespace-pre-wrap break-words font-mono text-sm">{file.content}</pre>
      )}
    </div>
  );
}
