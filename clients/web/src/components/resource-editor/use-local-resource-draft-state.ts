"use client";

import { useCallback, useEffect, useReducer } from "react";
import { resourceDraftReducer } from "./resource-draft-reducer";
import type {
  ResourceDraftAction,
  ResourceDraftState,
} from "./resource-draft-state-types";

interface LocalDraftState {
  identity: string;
  state: ResourceDraftState;
}

interface LocalDraftDispatch {
  identity: string;
  initialState: ResourceDraftState;
  action?: ResourceDraftAction;
}

export function useLocalResourceDraftState(
  identity: string,
  initialState: ResourceDraftState,
) {
  const [local, dispatchLocal] = useReducer(localDraftReducer, {
    identity,
    state: initialState,
  });
  const state = local.identity === identity ? local.state : initialState;

  useEffect(() => {
    if (local.identity === identity) return;
    dispatchLocal({ identity, initialState });
  }, [identity, initialState, local.identity]);

  const dispatch = useCallback((action: ResourceDraftAction) => {
    dispatchLocal({ identity, initialState, action });
  }, [identity, initialState]);

  return [state, dispatch] as const;
}

function localDraftReducer(
  local: LocalDraftState,
  dispatch: LocalDraftDispatch,
): LocalDraftState {
  const state = local.identity === dispatch.identity
    ? local.state
    : dispatch.initialState;
  return {
    identity: dispatch.identity,
    state: dispatch.action
      ? resourceDraftReducer(state, dispatch.action)
      : state,
  };
}
