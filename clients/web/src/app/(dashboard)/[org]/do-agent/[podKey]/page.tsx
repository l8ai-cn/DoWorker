"use client";

import { useParams } from "next/navigation";
import { useAcpRelay } from "@/hooks/useAcpRelay";
import { AcpActivityStream } from "@/components/workspace/acp/AcpActivityStream";
import { AcpPromptInput } from "@/components/workspace/acp/AcpPromptInput";
import { AcpPermissionDialog } from "@/components/workspace/acp/AcpPermissionDialog";
import { useAcpSessionField } from "@/stores/acpSession";
import { DoAgentTopBar } from "@/components/doagent/DoAgentTopBar";
import { DoAgentGoalBar, useDoAgentGoalSync } from "@/components/doagent/DoAgentGoalBar";

export default function DoAgentConsolePage() {
  const params = useParams();
  const podKey = typeof params.podKey === "string" ? params.podKey : "";
  const active = !!podKey;

  useAcpRelay(podKey, `doagent-${podKey}`, active);
  useDoAgentGoalSync(podKey, active);
  const pendingPermissions = useAcpSessionField(podKey, (s) => s.pendingPermissions);

  return (
    <div className="flex h-full w-full min-w-0 flex-col overflow-hidden">
      <DoAgentTopBar podKey={podKey} />
      <div className="min-h-0 flex-1 overflow-hidden">
        <AcpActivityStream podKey={podKey} />
      </div>
      <DoAgentGoalBar podKey={podKey} />
      <AcpPromptInput podKey={podKey} />
      {pendingPermissions.length > 0 && (
        <AcpPermissionDialog podKey={podKey} permissions={pendingPermissions} />
      )}
    </div>
  );
}
