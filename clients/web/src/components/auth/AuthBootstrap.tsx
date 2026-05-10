"use client";

import { useEffect, useState } from "react";
import { useAuthStore } from "@/stores/auth";
import { WasmProvider } from "@/providers/WasmProvider";
import { CenteredSpinner } from "@/components/ui/spinner";

// Single hydrate entry for any wasm-bound route group.
//
// Wraps WasmProvider and, once wasm is ready, runs the Rust auth
// bootstrap protocol (storage → token validation → optional refresh →
// identity re-fetch) before flipping `_hasHydrated`. Downstream
// components that gate on `_hasHydrated && user` therefore see a
// consistent state — no false-anonymous flash, no premature `/login`
// redirect for users whose session lives in localStorage of a window
// just opened (e.g. the popout terminal).
//
// Bootstrap is a no-op for genuinely anonymous users — the result is
// `Anonymous` and `_hasHydrated` still flips so /login can render.
export function AuthBootstrap({ children }: { children: React.ReactNode }) {
  return (
    <WasmProvider>
      <BootstrapGate>{children}</BootstrapGate>
    </WasmProvider>
  );
}

function BootstrapGate({ children }: { children: React.ReactNode }) {
  const [done, setDone] = useState(false);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        await useAuthStore.getState().bootstrap();
      } finally {
        if (cancelled) return;
        useAuthStore.getState().setHasHydrated(true);
        setDone(true);
      }
    })();
    return () => { cancelled = true; };
  }, []);

  if (!done) return <CenteredSpinner />;
  return <>{children}</>;
}
