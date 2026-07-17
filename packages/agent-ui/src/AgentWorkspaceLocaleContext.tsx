import { createContext, useContext, useMemo } from "react";

import {
  agentWorkspaceText,
  type AgentWorkspaceLocale,
} from "./agentWorkspaceText";

const AgentWorkspaceTextContext = createContext(agentWorkspaceText("en-US"));

export function AgentWorkspaceLocaleProvider({
  children,
  locale,
}: {
  children: React.ReactNode;
  locale: AgentWorkspaceLocale;
}) {
  const text = useMemo(() => agentWorkspaceText(locale), [locale]);
  return (
    <AgentWorkspaceTextContext.Provider value={text}>
      {children}
    </AgentWorkspaceTextContext.Provider>
  );
}

export function useAgentWorkspaceText() {
  return useContext(AgentWorkspaceTextContext);
}
