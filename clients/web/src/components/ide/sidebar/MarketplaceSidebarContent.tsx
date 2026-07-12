"use client";

import { usePathname, useRouter } from "next/navigation";
import { BookOpen, Blocks, Store } from "lucide-react";

import { cn } from "@/lib/utils";
import { useCurrentOrg } from "@/stores/auth";

const links = [
  { icon: Store, label: "应用市场", path: "marketplace" },
  { icon: Blocks, label: "组织技能库", path: "skills" },
];

export function MarketplaceSidebarContent({ className }: { className?: string }) {
  const pathname = usePathname();
  const router = useRouter();
  const currentOrg = useCurrentOrg();
  const orgSlug = currentOrg?.slug;

  return (
    <div className={cn("flex h-full flex-col", className)}>
      <div className="border-b border-border p-3">
        <h2 className="text-sm font-semibold">市场</h2>
        <p className="mt-1 text-xs leading-5 text-muted-foreground">
          为当前组织启用开箱即用的专家应用与能力组件。
        </p>
      </div>
      <div className="flex-1 space-y-0.5 p-2">
        {links.map(({ icon: Icon, label, path }) => {
          const href = orgSlug ? `/${orgSlug}/${path}` : "";
          const active = pathname.startsWith(href);
          return (
            <button
              key={path}
              type="button"
              disabled={!href}
              onClick={() => router.push(href)}
              className={cn(
                "flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm transition-colors",
                active
                  ? "bg-primary/10 font-medium text-primary"
                  : "text-muted-foreground hover:bg-muted hover:text-foreground",
              )}
            >
              <Icon className="h-4 w-4 shrink-0" />
              <span className="truncate">{label}</span>
            </button>
          );
        })}
      </div>
      <div className="bg-surface-muted/30 px-3 py-3 text-xs leading-5 text-muted-foreground">
        <BookOpen className="mr-1.5 inline h-3.5 w-3.5" />
        启用前会展示权限、运行要求与市场额度。
      </div>
    </div>
  );
}
