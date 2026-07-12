import Link from "next/link";
import { CheckCircle2, Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import type { LightOrganization } from "@/lib/light-auth";

export function OrganizationStep({
  organizations,
  value,
  onChange,
  onContinue,
  fixedOrganization,
}: {
  organizations: LightOrganization[];
  value: string;
  onChange: (value: string) => void;
  onContinue: () => void;
  fixedOrganization?: LightOrganization;
}) {
  return (
    <section className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold text-foreground">
          {fixedOrganization ? "检查当前组织的启用条件" : "选择使用组织"}
        </h2>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          {fixedOrganization
            ? `应用会安装到「${fixedOrganization.name}」，创建后不能直接移动。`
            : "应用会安装到你选择的组织，创建后不能直接移动。"}
        </p>
      </div>
      {fixedOrganization ? (
        <div className="rounded-lg border border-primary/20 bg-primary/5 p-4 text-sm text-foreground">
          启用目标：{fixedOrganization.name}
        </div>
      ) : organizations.length === 0 ? (
        <ErrorState message="当前账户还没有可用组织，请先创建组织。" />
      ) : (
        <select
          value={value}
          onChange={(event) => onChange(event.target.value)}
          className="h-12 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
          aria-label="选择使用组织"
        >
          <option value="">请选择组织</option>
          {organizations.map((organization) => (
            <option key={organization.id} value={organization.id}>
              {organization.name}
            </option>
          ))}
        </select>
      )}
      <Button className="w-full" size="lg" disabled={!value} onClick={onContinue}>
        检查启用条件
      </Button>
    </section>
  );
}

export function AcquireShell({ children }: { children: React.ReactNode }) {
  return (
    <main className="min-h-screen bg-surface px-4 py-10 sm:py-16">
      <div className="mx-auto max-w-3xl space-y-8 rounded-xl border border-border bg-card p-6 shadow-sm sm:p-9">
        {children}
      </div>
    </main>
  );
}

export function LoadingState({
  label = "正在加载启用信息",
}: {
  label?: string;
}) {
  return (
    <div className="flex min-h-40 items-center justify-center gap-3 text-sm text-muted-foreground">
      <Loader2 className="h-5 w-5 animate-spin" />
      {label}
    </div>
  );
}

export function ErrorState({ message }: { message: string }) {
  return (
    <div className="rounded-lg border border-danger/30 bg-danger-bg p-5 text-sm text-foreground">
      {message}
    </div>
  );
}

export function InlineError({ message }: { message: string }) {
  return (
    <p role="alert" className="text-sm text-danger">
      {message}
    </p>
  );
}

export function SuccessState({
  organization,
}: {
  organization: LightOrganization;
}) {
  return (
    <section className="py-4 text-center">
      <CheckCircle2 className="mx-auto h-12 w-12 text-success" />
      <h2 className="mt-4 text-2xl font-semibold text-foreground">应用已启用</h2>
      <p className="mt-2 text-sm text-muted-foreground">
        已在「{organization.name}」创建可用实例。
      </p>
      <Button asChild className="mt-6">
        <Link href={`/${organization.slug}/experts`}>查看专家应用</Link>
      </Button>
    </section>
  );
}
