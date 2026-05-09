import React, { useEffect, useState } from "react";
// Import via the `next-themes` alias (resolves to ./shims/next-themes →
// ./ThemeProvider) so this file and any web cross-imports of `next-themes`
// land on the same module identity. Direct `./ThemeProvider` imports
// can become a separate bundle module from the aliased path, causing
// the two `ThemeContext` instances and "useTheme must be used within
// ThemeProvider" failures we saw at runtime.
import { ThemeProvider } from "next-themes";
import { DesktopIntlProvider } from "./IntlProvider";
import { RealtimeProvider } from "./RealtimeProvider";
import { Toaster } from "sonner";
import { ensurePlatformReady } from "@agentsmesh/service-runtime";
import { useAuthStore } from "@/stores/auth";

async function bootstrapAuth() {
  await ensurePlatformReady();
  await useAuthStore.getState().bootstrap();
  useAuthStore.getState().setHasHydrated(true);
}

function PlatformGate({ children }: { children: React.ReactNode }) {
  const [ready, setReady] = useState(false);

  useEffect(() => {
    bootstrapAuth().then(() => setReady(true));
  }, []);

  if (!ready) return null;
  return <>{children}</>;
}

export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider defaultTheme="system" attribute="class">
      <DesktopIntlProvider>
        <PlatformGate>
          <RealtimeProvider>
            {children}
          </RealtimeProvider>
        </PlatformGate>
        <Toaster richColors position="top-right" />
      </DesktopIntlProvider>
    </ThemeProvider>
  );
}
