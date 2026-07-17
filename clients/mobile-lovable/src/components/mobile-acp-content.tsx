import { TriangleAlert } from "lucide-react";
import { cn } from "@/lib/utils";

export function AcpMessageList({
  messages,
}: {
  messages: Array<{ text: string; role: string; complete?: boolean }>;
}) {
  return (
    <div className="min-h-0 flex-1 space-y-3 overflow-y-auto px-3 py-4">
      {messages.length === 0 ? (
        <p className="pt-8 text-center text-sm text-muted-foreground">等待 Worker 返回对话上下文</p>
      ) : (
        messages.map((message, index) => (
          <article
            key={`${message.role}-${index}-${message.text}`}
            className={cn(
              "max-w-[88%] rounded-lg px-3 py-2 text-sm leading-6",
              message.role === "user"
                ? "ml-auto bg-primary text-primary-foreground"
                : "bg-surface text-foreground",
            )}
          >
            {message.text}
          </article>
        ))
      )}
    </div>
  );
}

export function AcpPermissions(props: {
  disabled: boolean;
  permissions: Array<{ requestId: string; toolName: string; description: string }>;
  onRespond: (requestId: string, approved: boolean) => void;
}) {
  return props.permissions.map((permission) => (
    <section key={permission.requestId} className="border-t border-warning/30 bg-warning/10 px-3 py-3">
      <div className="flex gap-2">
        <TriangleAlert className="mt-0.5 h-4 w-4 shrink-0 text-warning" />
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium">{permission.description || permission.toolName}</p>
          <p className="mt-0.5 font-mono text-[11px] text-muted-foreground">{permission.toolName}</p>
          <div className="mt-2 flex gap-2">
            <button
              onClick={() => props.onRespond(permission.requestId, false)}
              disabled={props.disabled}
              className="min-h-9 rounded-md border border-border px-3 text-xs font-medium disabled:opacity-50"
            >
              拒绝
            </button>
            <button
              onClick={() => props.onRespond(permission.requestId, true)}
              disabled={props.disabled}
              aria-label={`允许 ${permission.toolName}`}
              className="min-h-9 rounded-md bg-primary px-3 text-xs font-semibold text-primary-foreground disabled:opacity-50"
            >
              允许
            </button>
          </div>
        </div>
      </div>
    </section>
  ));
}
