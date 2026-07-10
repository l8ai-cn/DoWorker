"use client";

import { useMemo, useState } from "react";
import { DatabaseZap, Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { PillTabs } from "@/components/ui/pill-tabs";
import { AIResourceDeletionDialog } from "./AIResourceDeletionDialog";
import { ConnectionCredentialsDialog } from "./ConnectionCredentialsDialog";
import { ModelResourceDialog } from "./ModelResourceDialog";
import { ProviderConnectionCard } from "./ProviderConnectionCard";
import { ProviderConnectionDialog } from "./ProviderConnectionDialog";
import { ResourceSummary } from "./ResourceSummary";
import type { AIResourceDeletionTarget, AIResourceScope, ModelResource, ProviderConnection } from "./types";
import { useAIResources } from "./useAIResources";

interface AIResourcesSettingsProps {
  scope: AIResourceScope;
  organizationSlug?: string;
  canManage: boolean;
}

const modalities = ["all", "chat", "image", "audio", "video", "embedding"];

export function AIResourcesSettings({ scope, organizationSlug, canManage }: AIResourcesSettingsProps) {
  const t = useTranslations();
  const [modality, setModality] = useState("all");
  const [connectionDialogOpen, setConnectionDialogOpen] = useState(false);
  const [connectionToEdit, setConnectionToEdit] = useState<ProviderConnection | null>(null);
  const [connectionToRotate, setConnectionToRotate] = useState<ProviderConnection | null>(null);
  const [resourceEditor, setResourceEditor] = useState<{ connection: ProviderConnection; resource?: ModelResource } | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<AIResourceDeletionTarget | null>(null);
  const resources = useAIResources(scope, organizationSlug);
  const visibleConnections = useMemo(() => resources.connections.filter((connection) => modality === "all" || connection.resources.some((resource) => resource.modalities.includes(modality))), [modality, resources.connections]);
  const closeConnectionDialog = () => {
    setConnectionDialogOpen(false);
    setConnectionToEdit(null);
  };

  if (resources.loading) return <AIResourcesLoading />;
  if (resources.error) return <AIResourcesError onRetry={resources.reload} />;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-foreground">{t("settings.aiResources.title")}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{t("settings.aiResources.description")}</p>
        </div>
        {canManage && <Button onClick={() => setConnectionDialogOpen(true)}><Plus className="mr-2 h-4 w-4" />{t("settings.aiResources.addConnection")}</Button>}
      </div>
      <section className="surface-card p-5">
        <h2 className="mb-4 text-base font-semibold text-foreground">{t("settings.aiResources.summary.title")}</h2>
        <ResourceSummary connections={resources.connections} effectiveResources={resources.effectiveResources} />
      </section>
      <PillTabs active={modality} onChange={setModality} tabs={modalities.map((item) => ({ id: item, label: t(`settings.aiResources.modality.${item}`) }))} />
      {resources.operationFailed && <AIResourcesOperationError />}
      {resources.connections.length === 0
        ? <AIResourcesEmpty canManage={canManage} onAdd={() => setConnectionDialogOpen(true)} />
        : visibleConnections.length === 0
          ? <AIResourcesFilteredEmpty />
          : <div className="space-y-3">{visibleConnections.map((connection) => <ProviderConnectionCard key={connection.id} connection={connection} modality={modality} activeModality={modality} canManage={canManage} onAddResource={(item) => setResourceEditor({ connection: item })} onEdit={setConnectionToEdit} onRotateCredentials={setConnectionToRotate} onEnabledChange={resources.changeConnectionEnabled} onValidate={resources.checkConnection} onDelete={(item) => setDeleteTarget({ kind: "connection", id: item.id, name: item.name })} onResourceEnabledChange={resources.changeResourceEnabled} onSetDefault={resources.makeDefault} onResourceEdit={(connection, resource) => setResourceEditor({ connection, resource })} onResourceDelete={(item) => setDeleteTarget({ kind: "resource", id: item.id, name: item.displayName })} />)}</div>}
      <ProviderConnectionDialog key={connectionToEdit?.id ?? "create"} open={connectionDialogOpen || Boolean(connectionToEdit)} catalog={resources.catalog} connection={connectionToEdit ?? undefined} onOpenChange={(open) => !open && closeConnectionDialog()} onSubmit={resources.createConnection} onUpdate={resources.updateProviderConnection} />
      <ConnectionCredentialsDialog key={`credentials-${connectionToRotate?.id ?? "none"}`} connection={connectionToRotate} provider={resources.catalog.find((provider) => provider.key === connectionToRotate?.providerKey)} onOpenChange={() => setConnectionToRotate(null)} onSubmit={resources.rotateCredentials} />
      <ModelResourceDialog key={`resource-${resourceEditor?.resource?.id ?? resourceEditor?.connection.id ?? "none"}`} connection={resourceEditor?.connection ?? null} resource={resourceEditor?.resource} provider={resources.catalog.find((provider) => provider.key === resourceEditor?.connection.providerKey)} onOpenChange={() => setResourceEditor(null)} onSubmit={resources.createModelResource} onUpdate={resources.updateModelResource} />
      <AIResourceDeletionDialog target={deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)} onConfirm={(target) => target.kind === "connection" ? resources.removeConnection(target.id) : resources.removeResource(target.id)} />
    </div>
  );
}

function AIResourcesLoading() {
  const t = useTranslations();
  return <div className="py-12 text-sm text-muted-foreground">{t("settings.aiResources.loading")}</div>;
}

function AIResourcesError({ onRetry }: { onRetry: () => Promise<unknown> }) {
  const t = useTranslations();
  return <div role="alert" className="rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">{t("settings.aiResources.loadError")}<Button className="ml-3" size="sm" variant="outline" onClick={() => void onRetry()}>{t("settings.aiResources.retry")}</Button></div>;
}

function AIResourcesOperationError() {
  const t = useTranslations();
  return <p role="alert" className="rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">{t("settings.aiResources.operationError")}</p>;
}

function AIResourcesFilteredEmpty() {
  const t = useTranslations();
  return <p className="py-6 text-sm text-muted-foreground">{t("settings.aiResources.emptyResources")}</p>;
}

function AIResourcesEmpty({ canManage, onAdd }: { canManage: boolean; onAdd: () => void }) {
  const t = useTranslations();
  return <EmptyState icon={<DatabaseZap />} title={t("settings.aiResources.empty.title")} description={t("settings.aiResources.empty.description")} actions={canManage ? <Button onClick={onAdd}>{t("settings.aiResources.addConnection")}</Button> : undefined} />;
}
