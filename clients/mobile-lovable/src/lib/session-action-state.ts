import { createContext, useContext } from "react";
import type { SessionMessageAttachment } from "./session-message-upload";

export interface SessionActions {
  onSend?: (text: string, attachments: SessionMessageAttachment[]) => Promise<void>;
  onApprove?: (elicitationId: string, accept: boolean) => Promise<void>;
  onStop?: () => Promise<void>;
}

export const SessionActionContext = createContext<SessionActions>({});

export function useSessionActions(): SessionActions {
  return useContext(SessionActionContext);
}
