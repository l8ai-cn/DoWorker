"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { Logo } from "@/components/common";
import { Button } from "@/components/ui/button";
import { useLightSession } from "@/hooks/useLightSession";
import { fetchFirstOrgSlug } from "@/lib/light-auth";
import { updateLightSessionOrgSlug } from "@/lib/light-session";

export function MarketplaceHeader() {
  const router = useRouter();
  const { session, hydrated } = useLightSession();

  async function openConsole() {
    if (!session?.isAuthenticated) {
      router.push("/login?redirect=%2Fmarketplace");
      return;
    }
    const orgSlug = session.currentOrgSlug || (await fetchFirstOrgSlug());
    if (!orgSlug) {
      router.push("/onboarding/create-org");
      return;
    }
    updateLightSessionOrgSlug(orgSlug);
    router.push(`/${orgSlug}/workspace`);
  }

  return (
    <header className="sticky top-0 z-40 border-b border-border bg-background/95 backdrop-blur">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <div className="flex items-center gap-8">
          <Link href="/" className="flex items-center gap-2.5" aria-label="Do Worker 首页">
            <span className="h-8 w-8 overflow-hidden rounded-md">
              <Logo />
            </span>
            <span className="text-base font-semibold text-foreground">Do Worker</span>
          </Link>
          <nav className="hidden items-center gap-6 text-sm md:flex">
            <Link href="/marketplace" className="font-medium text-foreground">
              专家应用
            </Link>
            <Link href="/docs" className="text-muted-foreground hover:text-foreground">
              使用文档
            </Link>
          </nav>
        </div>
        {hydrated ? (
          <Button size="sm" variant="outline" onClick={openConsole}>
            {session?.isAuthenticated ? "控制台" : "登录"}
          </Button>
        ) : null}
      </div>
    </header>
  );
}
