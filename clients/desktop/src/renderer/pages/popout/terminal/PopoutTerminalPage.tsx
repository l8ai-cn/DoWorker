import { useEffect } from "react";
import { useParams } from "next/navigation";
import { useCurrentUser } from "@/stores/auth";
import { RealtimeProvider } from "@/providers/RealtimeProvider";
import { TerminalPane } from "@/components/workspace/TerminalPane";
import { getShortPodKey } from "@/lib/pod-display-name";

// Auth gating happens via <RequireAuth> in router.tsx so this component
// is only mounted when a user exists. New BrowserWindows opened via
// window.open share the renderer's bootstrapAuth() flow (PlatformGate).
export function PopoutTerminalPage() {
  const { podKey } = useParams<{ podKey: string }>();
  const user = useCurrentUser();

  useEffect(() => {
    if (podKey) {
      document.title = `Terminal - ${getShortPodKey(podKey)}`;
    }
  }, [podKey]);

  if (!user || !podKey) return null;

  return (
    <RealtimeProvider>
      <div className="h-screen w-screen bg-terminal-bg">
        <TerminalPane
          paneId={`popout-${podKey}`}
          podKey={podKey}
          isActive={true}
          showHeader={true}
          allowSplit={false}
        />
      </div>
    </RealtimeProvider>
  );
}
