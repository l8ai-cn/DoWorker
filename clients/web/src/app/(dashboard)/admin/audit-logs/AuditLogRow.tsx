import { Badge, type BadgeProps } from "@/components/ui/badge";
import type { AuditLog } from "@/lib/api/admin/types";

const actionColors: Record<string, BadgeProps["variant"]> = {
  "user.view": "secondary",
  "user.update": "default",
  "user.disable": "destructive",
  "user.enable": "success",
  "user.grant_admin": "warning",
  "user.revoke_admin": "warning",
  "organization.view": "secondary",
  "organization.update": "default",
  "organization.delete": "destructive",
  "runner.view": "secondary",
  "runner.disable": "destructive",
  "runner.enable": "success",
  "runner.delete": "destructive",
};

export function AuditLogRow({ log }: { log: AuditLog }) {
  return (
    <div className="flex flex-col gap-3 rounded-lg border border-border p-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex flex-col gap-1">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant={actionColors[log.action] || "secondary"}>{log.action}</Badge>
          <span className="text-sm text-muted-foreground">
            {log.target_type} #{log.target_id}
          </span>
        </div>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          {log.admin_user && (
            <span>by {log.admin_user.name || log.admin_user.username}</span>
          )}
          {log.ip_address && <span>from {log.ip_address}</span>}
        </div>
      </div>
      <div className="hidden text-right text-xs text-muted-foreground sm:block">
        {new Date(log.created_at).toLocaleString()}
      </div>
    </div>
  );
}
