import { useEffect, useState } from "react";
import type { Host } from "@/hooks/useHosts";
import { getHostIdentity, isElectronShell, onHostStatusChanged, type HostIdentity } from "@/lib/nativeBridge";

export function useNewChatDesktopHost(allHosts: Host[]) {
  const [desktopHost, setDesktopHost] = useState<HostIdentity | null>(null);

  useEffect(() => {
    if (!isElectronShell()) return;
    let cancelled = false;
    const refresh = () => void getHostIdentity().then((state) => !cancelled && setDesktopHost(state));
    refresh();
    const unsubscribe = onHostStatusChanged(refresh);
    return () => {
      cancelled = true;
      unsubscribe();
    };
  }, []);

  const thisMachineHostId = desktopHost?.hostId ?? null;
  const thisMachineInList = thisMachineHostId != null && allHosts.some((host) => host.host_id === thisMachineHostId);
  const canConnectThisMachine = Boolean(desktopHost?.cliInstalled);
  return {
    setDesktopHost,
    thisMachineHostId,
    canConnectThisMachine,
    showConnectThisMachine: canConnectThisMachine && !thisMachineInList,
  };
}
