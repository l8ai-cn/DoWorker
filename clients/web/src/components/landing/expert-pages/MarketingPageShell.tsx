import type { ReactNode } from "react";

import { FinalCTA, Footer, Navbar } from "@/components/landing";

export function MarketingPageShell({ children }: { children: ReactNode }) {
  return (
    <div className="azure-theme expert-home min-h-screen bg-[var(--expert-bg)]">
      <Navbar />
      <main>
        {children}
        <FinalCTA />
      </main>
      <Footer />
    </div>
  );
}
