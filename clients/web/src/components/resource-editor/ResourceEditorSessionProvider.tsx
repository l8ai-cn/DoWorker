"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
} from "react";
import { resourceDraftReducer } from "./resource-draft-reducer";
import type {
  ResourceDraftAction,
  ResourceDraftState,
} from "./resource-draft-state-types";

interface ResourceEditorSessionContextValue {
  sessions: Record<string, ResourceDraftState>;
  dispatch: (action: ResourceEditorSessionAction) => void;
}

type ResourceEditorSessionAction =
  | { type: "initialize"; key: string; initialState: ResourceDraftState }
  | {
    type: "dispatch";
    key: string;
    initialState: ResourceDraftState;
    action: ResourceDraftAction;
  };

interface ResourceEditorSession {
  state: ResourceDraftState;
  dispatch: (action: ResourceDraftAction) => void;
}

const ResourceEditorSessionContext =
  createContext<ResourceEditorSessionContextValue | null>(null);

export function ResourceEditorSessionProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [sessions, dispatch] = useReducer(resourceEditorSessionReducer, {});
  const value = useMemo(
    () => ({ sessions, dispatch }),
    [sessions],
  );

  return (
    <ResourceEditorSessionContext.Provider value={value}>
      {children}
    </ResourceEditorSessionContext.Provider>
  );
}

export function useResourceEditorSession(
  key: string | undefined,
  initialState: ResourceDraftState,
): ResourceEditorSession | null {
  const context = useContext(ResourceEditorSessionContext);
  if (key && !context) {
    throw new Error("Resource editor sessions require ResourceEditorSessionProvider.");
  }
  const dispatchToSession = context?.dispatch;

  useEffect(() => {
    if (!key || !dispatchToSession) return;
    dispatchToSession({ type: "initialize", key, initialState });
  }, [dispatchToSession, initialState, key]);

  const dispatch = useCallback((action: ResourceDraftAction) => {
    if (!key || !dispatchToSession) return;
    dispatchToSession({ type: "dispatch", key, initialState, action });
  }, [dispatchToSession, initialState, key]);

  if (!key || !context) return null;
  return {
    state: context.sessions[key] ?? initialState,
    dispatch,
  };
}

function resourceEditorSessionReducer(
  sessions: Record<string, ResourceDraftState>,
  action: ResourceEditorSessionAction,
): Record<string, ResourceDraftState> {
  if (action.type === "initialize") {
    if (sessions[action.key]) return sessions;
    return { ...sessions, [action.key]: action.initialState };
  }
  const state = sessions[action.key] ?? action.initialState;
  return {
    ...sessions,
    [action.key]: resourceDraftReducer(state, action.action),
  };
}
