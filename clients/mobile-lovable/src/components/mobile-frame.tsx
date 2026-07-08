import type { ReactNode } from "react";
import { BottomNav } from "./bottom-nav";

/**
 * MobileFrame — renders a phone-sized surface centered on desktop, and goes
 * full-viewport on real mobile devices. Keeps the design "a mobile app" no
 * matter where it's viewed.
 */
export function MobileFrame({ children, hideNav = false }: { children: ReactNode; hideNav?: boolean }) {
  const body = (
    <div className="flex min-h-full flex-col">
      <div className="flex-1">{children}</div>
      {!hideNav && <BottomNav />}
    </div>
  );

  return (
    <div className="min-h-screen w-full bg-background text-foreground">
      {/* Desktop: show phone chrome. Mobile: fill screen. */}
      <div className="hidden md:flex min-h-screen items-center justify-center p-8 grid-bg">
        <div className="relative">
          <div className="absolute -inset-4 rounded-[3rem] bg-gradient-to-br from-primary/20 via-transparent to-accent/20 blur-2xl" />
          <div className="relative h-[860px] w-[400px] overflow-hidden rounded-[2.5rem] border border-border bg-background shadow-2xl ring-1 ring-white/5">
            <div className="pointer-events-none absolute inset-x-0 top-0 z-50 flex justify-center pt-2">
              <div className="h-5 w-28 rounded-full bg-black/70" />
            </div>
            <div className="h-full w-full overflow-y-auto scrollbar-none md:pt-10">
              {body}
            </div>
          </div>
        </div>
      </div>
      <div className="md:hidden min-h-screen">{body}</div>
    </div>
  );
}

