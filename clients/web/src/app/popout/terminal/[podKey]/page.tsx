"use client";

import { useEffect } from "react";
import { useParams } from "next/navigation";
import { useCurrentUser } from "@/stores/auth";
import { usePod } from "@/stores/pod";
import { RealtimeProvider } from "@/providers/RealtimeProvider";
import { AgentPanel } from "@/components/workspace/AgentPanel";
import { TerminalPane } from "@/components/workspace/TerminalPane";
import { getShortPodKey } from "@/lib/pod-display-name";
import { POD_MODE_ACP } from "@/lib/pod-modes";

export default function PopoutTerminalPage() {
  const { podKey } = useParams<{ podKey: string }>();
  const user = useCurrentUser();
  const pod = usePod(podKey);

  useEffect(() => {
    if (podKey) {
      document.title = `Terminal - ${getShortPodKey(podKey)}`;
    }
  }, [podKey]);

  if (!user || !podKey) return null;

  return (
    <RealtimeProvider>
      <div className="h-screen w-screen bg-terminal-bg">
        {pod?.interaction_mode === POD_MODE_ACP ? (
          <AgentPanel
            paneId={`popout-${podKey}`}
            podKey={podKey}
            isActive={true}
            showHeader={true}
            allowSplit={false}
          />
        ) : (
          <TerminalPane
            paneId={`popout-${podKey}`}
            podKey={podKey}
            isActive={true}
            showHeader={true}
            allowSplit={false}
          />
        )}
      </div>
    </RealtimeProvider>
  );
}
