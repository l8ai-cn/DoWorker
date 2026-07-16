import type { ReactNode } from "react";
import { SessionActionContext, type SessionActions } from "./session-action-state";

export function SessionActionProvider({
  value,
  children,
}: {
  value: SessionActions;
  children: ReactNode;
}) {
  return <SessionActionContext.Provider value={value}>{children}</SessionActionContext.Provider>;
}
