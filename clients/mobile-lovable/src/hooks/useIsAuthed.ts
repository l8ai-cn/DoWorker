import { useSyncExternalStore } from "react";
import { readAuthToken, subscribeAuthChanges } from "@/lib/auth-store";

export function useIsAuthed(): boolean {
  return useSyncExternalStore(
    subscribeAuthChanges,
    () => Boolean(readAuthToken()),
    () => false,
  );
}
