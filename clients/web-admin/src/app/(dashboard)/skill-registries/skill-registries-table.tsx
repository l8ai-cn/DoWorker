import {
  RefreshCw,
  Trash2,
  GitBranch,
  Loader2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { SkillRegistry } from "@/lib/api/admin";
import { formatDate, formatRelativeTime } from "@/lib/utils";

interface SkillRegistriesTableProps {
  registries: SkillRegistry[];
  isLoading: boolean;
  syncingIds: Set<number>;
  onSync: (id: number) => void;
  onDelete: (registry: SkillRegistry) => void;
}

function SyncStatusBadge({ status }: { status: string }) {
  switch (status) {
    case "success":
      return <Badge variant="success">成功</Badge>;
    case "syncing":
      return (
        <Badge variant="warning" className="gap-1">
          <Loader2 className="h-3 w-3 animate-spin" />
          同步中
        </Badge>
      );
    case "failed":
      return <Badge variant="destructive">失败</Badge>;
    case "pending":
      return <Badge variant="secondary">待同步</Badge>;
    default:
      return <Badge variant="secondary">{status}</Badge>;
  }
}

export function SkillRegistriesTable({
  registries,
  isLoading,
  syncingIds,
  onSync,
  onDelete,
}: SkillRegistriesTableProps) {
  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>仓库 URL</TableHead>
            <TableHead>分支</TableHead>
            <TableHead>类型</TableHead>
            <TableHead>状态</TableHead>
            <TableHead>技能数量</TableHead>
            <TableHead>上次同步</TableHead>
            <TableHead className="w-28">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading ? (
            Array.from({ length: 3 }).map((_, i) => (
              <TableRow key={i}>
                <TableCell colSpan={7}>
                  <div className="h-12 animate-pulse rounded bg-muted" />
                </TableCell>
              </TableRow>
            ))
          ) : registries.length === 0 ? (
            <TableRow>
              <TableCell colSpan={7} className="py-8 text-center text-muted-foreground">
                暂无技能源配置
              </TableCell>
            </TableRow>
          ) : (
            registries.map((registry) => (
              <RegistryRow
                key={registry.id}
                registry={registry}
                isSyncing={syncingIds.has(registry.id) || registry.sync_status === "syncing"}
                onSync={() => onSync(registry.id)}
                onDelete={() => onDelete(registry)}
              />
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}

function RegistryRow({
  registry,
  isSyncing,
  onSync,
  onDelete,
}: {
  registry: SkillRegistry;
  isSyncing: boolean;
  onSync: () => void;
  onDelete: () => void;
}) {
  return (
    <TableRow>
      <TableCell>
        <div className="flex items-center gap-2 font-medium">
          <GitBranch className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span className="truncate max-w-xs" title={registry.repository_url}>
            {registry.repository_url}
          </span>
        </div>
      </TableCell>
      <TableCell>
        <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
          {registry.branch || "main"}
        </code>
      </TableCell>
      <TableCell>
        <Badge variant="outline">{registry.source_type}</Badge>
      </TableCell>
      <TableCell>
        <div className="flex flex-col gap-1">
          <SyncStatusBadge status={registry.sync_status} />
          {registry.sync_status === "failed" && registry.sync_error && (
            <span
              className="text-xs text-destructive truncate max-w-[200px]"
              title={registry.sync_error}
            >
              {registry.sync_error}
            </span>
          )}
        </div>
      </TableCell>
      <TableCell>
        <span className="font-medium">{registry.skill_count}</span>
      </TableCell>
      <TableCell>
        {registry.last_synced_at ? (
          <span className="text-sm text-muted-foreground" title={formatDate(registry.last_synced_at)}>
            {formatRelativeTime(registry.last_synced_at)}
          </span>
        ) : (
          <span className="text-sm text-muted-foreground">从未同步</span>
        )}
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="icon" onClick={onSync} disabled={isSyncing} title="立即同步">
            <RefreshCw className={`h-4 w-4 ${isSyncing ? "animate-spin" : ""}`} />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            onClick={onDelete}
            title="删除技能源"
            className="text-destructive hover:text-destructive"
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </TableCell>
    </TableRow>
  );
}
