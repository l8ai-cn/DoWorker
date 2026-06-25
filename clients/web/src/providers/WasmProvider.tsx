"use client";

import { useEffect, useState } from "react";
import { initWasmCore } from "@/lib/wasm-core";

// Loads the wasm core. Does NOT touch auth state — that's AuthBootstrap's
// job (see components/auth/AuthBootstrap.tsx).
export function WasmProvider({ children }: { children: React.ReactNode }) {
  const [ready, setReady] = useState(false);

  useEffect(() => {
    initWasmCore().then(() => setReady(true));
  }, []);

  if (!ready) return null;
  return <>{children}</>;
}
