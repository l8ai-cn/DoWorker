"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { AppWindow, Store } from "lucide-react";

import { ApplicationCard } from "./ApplicationCard";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { CenteredSpinner } from "@/components/ui/spinner";
import { useExperts, useExpertStore } from "@/stores/expert";
import { useCurrentOrg } from "@/stores/auth";
import {
  expertIDFromRuntimeRef,
  fetchOrganizationApplications,
  type MarketplaceOrganizationApplication,
} from "@/lib/marketplace/application-api";

type LoadState =
  | { kind: "loading" }
  | { kind: "ready"; applications: MarketplaceOrganizationApplication[] }
  | { kind: "error"; message: string };

export function ApplicationsPage({
  orgSlug,
  installationID,
}: {
  orgSlug: string;
  installationID?: string;
}) {
  const currentOrg = useCurrentOrg();
  const experts = useExperts();
  const fetchExperts = useExpertStore((state) => state.fetchExperts);
  const expertError = useExpertStore((state) => state.error);
  const [state, setState] = useState<LoadState>({ kind: "loading" });

  useEffect(() => {
    if (!currentOrg || currentOrg.slug !== orgSlug) return;
    let active = true;
    void fetchExperts();
    void fetchOrganizationApplications(currentOrg.id)
      .then((applications) => {
        if (active) setState({ kind: "ready", applications });
      })
      .catch((cause) => {
        if (active) setState({ kind: "error", message: messageFrom(cause) });
      });
    return () => {
      active = false;
    };
  }, [currentOrg, fetchExperts, orgSlug]);

  if (!currentOrg || currentOrg.slug !== orgSlug || state.kind === "loading") {
    return <CenteredSpinner className="h-full" />;
  }
  const applications = state.kind === "ready" && installationID
    ? state.applications.filter((application) => application.installation_id === installationID)
    : state.kind === "ready" ? state.applications : [];

  return (
    <main className="mx-auto w-full max-w-6xl space-y-8 px-5 py-8 lg:px-8">
      <ApplicationsHeader />
      {state.kind === "error" ? <ApplicationsError message={state.message} /> : null}
      {state.kind === "ready" && applications.length === 0 ? <ApplicationsEmpty focused={Boolean(installationID)} /> : null}
      {state.kind === "ready" && applications.length > 0 ? (
        <>
          {expertError ? (
            <p role="status" className="rounded-lg border border-warning/30 bg-warning-bg px-4 py-3 text-sm text-foreground">
              已显示启用状态，但暂时无法确认专家任务入口：{expertError}
            </p>
          ) : null}
          <section aria-label="已启用应用列表" className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {applications.map((application) => (
              <ApplicationCard
                key={application.installation_id}
                application={application}
                orgSlug={orgSlug}
                expertSlug={expertSlugFor(application.runtime_ref, experts)}
              />
            ))}
          </section>
        </>
      ) : null}
    </main>
  );
}

function ApplicationsHeader() {
  return (
    <header className="flex flex-col justify-between gap-4 border-b border-border pb-6 sm:flex-row sm:items-end">
      <div>
        <p className="text-sm font-medium text-primary">组织应用中心</p>
        <h1 className="mt-1 text-3xl font-semibold tracking-tight text-foreground">已启用应用</h1>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          从这里开始第一个任务，并查看每个市场应用的真实启用状态。
        </p>
      </div>
      <Button asChild variant="outline" className="gap-2">
        <Link href="https://market.l8ai.cn">
          <Store className="h-4 w-4" />
          浏览应用市场
        </Link>
      </Button>
    </header>
  );
}

function ApplicationsEmpty({ focused = false }: { focused?: boolean }) {
  return (
    <EmptyState
      size="full"
      icon={<AppWindow className="h-12 w-12" />}
      title={focused ? "找不到这个已启用应用" : "还没有已启用的应用"}
      description={focused ? "该应用可能已移除，或你没有当前组织的访问权限。" : "请先在公开应用市场选择应用，完成启用后会显示在这里。"}
      actions={(
        <Button asChild>
          <Link href="https://market.l8ai.cn">前往应用市场</Link>
        </Button>
      )}
    />
  );
}

function ApplicationsError({ message }: { message: string }) {
  return (
    <EmptyState
      size="full"
      icon={<AppWindow className="h-12 w-12" />}
      title="应用中心暂时无法加载"
      description={message}
      actions={<Button onClick={() => window.location.reload()}>重新加载</Button>}
    />
  );
}

function expertSlugFor(
  runtimeRef: string,
  experts: Array<{ id: number; slug: string }>,
): string | undefined {
  const expertID = expertIDFromRuntimeRef(runtimeRef);
  return expertID ? experts.find((expert) => expert.id === expertID)?.slug : undefined;
}

function messageFrom(cause: unknown): string {
  return cause instanceof Error ? cause.message : "请检查网络或组织访问权限后重试。";
}
