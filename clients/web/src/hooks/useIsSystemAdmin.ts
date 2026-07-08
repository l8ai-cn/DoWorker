"use client";

import { useEffect, useState } from "react";
import { userApi } from "@/lib/api";

// The wasm-cached auth user omits is_system_admin (see clients/core auth_types),
// so admin-only chrome resolves it via GetMe. Cache the in-flight promise at
// module scope so multiple mounts (ActivityBar, guards) share one request.
let cached: Promise<boolean> | null = null;

export function resolveIsSystemAdmin(): Promise<boolean> {
  if (!cached) {
    cached = userApi
      .getMe()
      .then(({ user }) => user.is_system_admin)
      .catch(() => {
        cached = null;
        return false;
      });
  }
  return cached;
}

export function useIsSystemAdmin(): boolean {
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    let active = true;
    resolveIsSystemAdmin().then((v) => {
      if (active) setIsAdmin(v);
    });
    return () => {
      active = false;
    };
  }, []);

  return isAdmin;
}
