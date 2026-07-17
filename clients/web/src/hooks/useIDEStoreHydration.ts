"use client";

import { useEffect } from "react";
import { useIDEStore } from "@/stores/ide";

export function useIDEStoreHydration(): boolean {
  const hydrated = useIDEStore((state) => state._hasHydrated);

  useEffect(() => {
    if (hydrated) return;
    void Promise.resolve(useIDEStore.persist.rehydrate()).then(
      () => {
        if (!useIDEStore.getState()._hasHydrated) {
          useIDEStore.getState().setHasHydrated(true);
        }
      },
      (error) => {
        console.error("[IDEStore] persisted state hydration failed", error);
      },
    );
  }, [hydrated]);

  return hydrated;
}
