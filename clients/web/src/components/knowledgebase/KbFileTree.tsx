"use client";

import { useCallback, useEffect, useState } from "react";
import { ChevronDown, ChevronRight, FileText, Folder } from "lucide-react";

import { Spinner } from "@/components/ui/spinner";
import { listKbDir, type KbDirEntry } from "@/lib/api/facade/knowledgeBaseApi";
import { cn } from "@/lib/utils";

interface KbFileTreeProps {
  orgSlug: string;
  kbSlug: string;
  selectedPath: string | null;
  onSelectFile: (path: string) => void;
}

export function KbFileTree({ orgSlug, kbSlug, selectedPath, onSelectFile }: KbFileTreeProps) {
  const [entries, setEntries] = useState<KbDirEntry[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    listKbDir(orgSlug, kbSlug, "")
      .then((items) => {
        if (!cancelled) setEntries(items);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "加载目录失败");
      });
    return () => {
      cancelled = true;
    };
  }, [orgSlug, kbSlug]);

  if (error) return <p className="p-3 text-xs text-destructive">{error}</p>;
  if (entries === null) {
    return (
      <div className="flex items-center gap-2 p-3 text-xs text-muted-foreground">
        <Spinner size="sm" />
        加载文件树
      </div>
    );
  }

  return (
    <div className="py-1">
      {entries.map((entry) => (
        <TreeNode
          key={entry.path}
          orgSlug={orgSlug}
          kbSlug={kbSlug}
          entry={entry}
          depth={0}
          selectedPath={selectedPath}
          onSelectFile={onSelectFile}
        />
      ))}
    </div>
  );
}

function TreeNode({
  orgSlug,
  kbSlug,
  entry,
  depth,
  selectedPath,
  onSelectFile,
}: {
  orgSlug: string;
  kbSlug: string;
  entry: KbDirEntry;
  depth: number;
  selectedPath: string | null;
  onSelectFile: (path: string) => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const [children, setChildren] = useState<KbDirEntry[] | null>(null);
  const [loading, setLoading] = useState(false);

  const toggle = useCallback(() => {
    const next = !expanded;
    setExpanded(next);
    if (next && children === null && !loading) {
      setLoading(true);
      listKbDir(orgSlug, kbSlug, entry.path)
        .then(setChildren)
        .catch(() => setChildren([]))
        .finally(() => setLoading(false));
    }
  }, [expanded, children, loading, orgSlug, kbSlug, entry.path]);

  const indent = { paddingLeft: `${depth * 14 + 8}px` };

  if (entry.type === "dir") {
    return (
      <div>
        <button
          type="button"
          onClick={toggle}
          style={indent}
          className="flex w-full items-center gap-1.5 py-1 pr-2 text-left text-sm hover:bg-surface-muted"
        >
          {expanded ? (
            <ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          ) : (
            <ChevronRight className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          )}
          <Folder className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span className="truncate">{entry.name}</span>
        </button>
        {expanded && loading && (
          <div style={{ paddingLeft: `${(depth + 1) * 14 + 8}px` }} className="py-1">
            <Spinner size="sm" />
          </div>
        )}
        {expanded &&
          children?.map((child) => (
            <TreeNode
              key={child.path}
              orgSlug={orgSlug}
              kbSlug={kbSlug}
              entry={child}
              depth={depth + 1}
              selectedPath={selectedPath}
              onSelectFile={onSelectFile}
            />
          ))}
      </div>
    );
  }

  return (
    <button
      type="button"
      onClick={() => onSelectFile(entry.path)}
      style={indent}
      className={cn(
        "flex w-full items-center gap-1.5 py-1 pr-2 text-left text-sm hover:bg-surface-muted",
        selectedPath === entry.path && "bg-primary/10 text-primary",
      )}
    >
      <span className="w-3.5 shrink-0" />
      <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
      <span className="truncate">{entry.name}</span>
    </button>
  );
}
