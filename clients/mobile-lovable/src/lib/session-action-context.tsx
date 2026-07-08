import { createContext, useContext, type ReactNode } from "react";

export interface SessionActions {
  onSend?: (text: string) => Promise<void>;
  onApprove?: (elicitationId: string, accept: boolean) => Promise<void>;
  onStop?: () => Promise<void>;
}

const SessionActionContext = createContext<SessionActions>({});

export function SessionActionProvider({
  value,
  children,
}: {
  value: SessionActions;
  children: ReactNode;
}) {
  return <SessionActionContext.Provider value={value}>{children}</SessionActionContext.Provider>;
}

export function useSessionActions(): SessionActions {
  return useContext(SessionActionContext);
}
