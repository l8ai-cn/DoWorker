import { ShieldAlert } from "lucide-react";
import { useRef, useState } from "react";

import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import type {
  AgentPermissionRequest,
  AgentPermissionResolution,
  AgentSessionRuntime,
} from "./contracts";
import { PermissionQuestionForm } from "./PermissionQuestionForm";

export function ApprovalDock({
  disabled = false,
  onError,
  permissions,
  runtime,
  sessionId,
}: {
  disabled?: boolean;
  onError: (error: unknown) => void;
  permissions: AgentPermissionRequest[];
  runtime: AgentSessionRuntime;
  sessionId: string;
}) {
  const text = useAgentWorkspaceText();
  const resolvingRef = useRef<string | null>(null);
  const [resolvingId, setResolvingId] = useState<string | null>(null);
  if (permissions.length === 0) return null;
  const permission = permissions[0];
  const isResolving = disabled || resolvingId === permission.id;
  const resolve = (result: AgentPermissionResolution) => {
    if (resolvingRef.current === permission.id) return;
    resolvingRef.current = permission.id;
    setResolvingId(permission.id);
    void runtime
      .resolvePermission(
        sessionId,
        crypto.randomUUID(),
        permission.id,
        result,
      )
      .catch((error) => {
        if (resolvingRef.current === permission.id) {
          resolvingRef.current = null;
          setResolvingId(null);
        }
        onError(error);
      });
  };
  if (permission.kind === "question") {
    return (
      <PermissionQuestionForm
        key={permission.id}
        disabled={isResolving}
        onReject={() => resolve({ action: "decline" })}
        onSubmit={(answers) =>
          resolve({ action: "accept", content: { answers } })
        }
        permission={permission}
      />
    );
  }
  return (
    <section className="border-t border-amber-300 bg-amber-50 px-3 py-3 text-amber-950 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-50">
      <div className="flex items-start gap-2">
        <ShieldAlert className="mt-0.5 size-4 shrink-0" />
        <div className="min-w-0 flex-1">
          <div className="text-sm font-medium">{permission.title}</div>
          <div className="mt-0.5 text-xs opacity-80">{permission.description}</div>
        </div>
        <button
          className="h-11 rounded-md border border-current px-3 text-xs font-medium"
          disabled={isResolving}
          onClick={() => resolve({ action: "decline" })}
          type="button"
        >
          {text.reject}
        </button>
        <button
          className="h-11 rounded-md bg-amber-900 px-3 text-xs font-medium text-white dark:bg-amber-200 dark:text-amber-950"
          disabled={isResolving}
          onClick={() =>
            resolve({ action: "accept", content: { answers: {} } })
          }
          type="button"
        >
          {text.approve}
        </button>
      </div>
    </section>
  );
}
