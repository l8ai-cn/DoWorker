import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { ModelResourceRow } from "./ModelResourceRow";
import type { ModelResource, ProviderConnection } from "./types";

interface ProviderConnectionCardProps {
  connection: ProviderConnection;
  modality: string;
  activeModality: string;
  canManage: boolean;
  onAddResource: (connection: ProviderConnection) => void;
  onEdit: (connection: ProviderConnection) => void;
  onRotateCredentials: (connection: ProviderConnection) => void;
  onEnabledChange: (connectionId: number, enabled: boolean) => Promise<boolean>;
  onValidate: (connectionId: number) => Promise<boolean>;
  onDelete: (connection: ProviderConnection) => void;
  onResourceEnabledChange: (resourceId: number, enabled: boolean) => Promise<boolean>;
  onSetDefault: (resourceId: number, modality: string) => Promise<boolean>;
  onResourceEdit: (connection: ProviderConnection, resource: ModelResource) => void;
  onResourceDelete: (resource: ModelResource) => void;
}

export function ProviderConnectionCard({
  connection,
  modality,
  activeModality,
  canManage,
  onAddResource,
  onEdit,
  onRotateCredentials,
  onEnabledChange,
  onValidate,
  onDelete,
  onResourceEnabledChange,
  onSetDefault,
  onResourceEdit,
  onResourceDelete,
}: ProviderConnectionCardProps) {
  const t = useTranslations();
  const manageable = canManage && connection.canManage;
  const resources = filterResources(connection.resources, modality);

  return (
    <section className="rounded-xl border border-border/70 bg-card" aria-label={connection.name}>
      <div className="flex flex-wrap items-start gap-3 border-b border-border/60 px-4 py-4 sm:px-6">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="truncate text-base font-semibold text-foreground">{connection.name}</h2>
            {connection.status === "invalid" && <Badge variant="destructive">{t("settings.aiResources.status.invalid")}</Badge>}
            {!connection.isEnabled && <Badge variant="warning">{t("settings.aiResources.status.disabled")}</Badge>}
          </div>
          <p className="mt-1 truncate text-sm text-muted-foreground">{connection.providerKey} · {connection.baseUrl}</p>
        </div>
        {manageable && (
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => void onValidate(connection.id)}>
              {t("settings.aiResources.validate")}
            </Button>
            <Button variant="ghost" size="sm" aria-label={`${t("settings.aiResources.connection.edit")}: ${connection.name}`} onClick={() => onEdit(connection)}>
              {t("settings.aiResources.connection.edit")}
            </Button>
            <Button variant="ghost" size="sm" aria-label={`${t("settings.aiResources.connection.rotate")}: ${connection.name}`} onClick={() => onRotateCredentials(connection)}>
              {t("settings.aiResources.connection.rotate")}
            </Button>
            <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" aria-label={`${t("settings.aiResources.connection.delete")}: ${connection.name}`} onClick={() => onDelete(connection)}>
              {t("settings.aiResources.connection.delete")}
            </Button>
            <Switch
              aria-label={`${t("settings.aiResources.connection.enabled")}: ${connection.name}`}
              checked={connection.isEnabled}
              onCheckedChange={(enabled) => void onEnabledChange(connection.id, enabled)}
            />
          </div>
        )}
      </div>
      <div className="px-4 py-2 sm:px-6">
        {resources.map((resource) => (
          <ModelResourceRow
            key={resource.id}
            resource={resource}
            activeModality={activeModality}
            canManage={manageable}
            onEnabledChange={onResourceEnabledChange}
            onSetDefault={onSetDefault}
            onEdit={(item) => onResourceEdit(connection, item)}
            onDelete={onResourceDelete}
          />
        ))}
        {resources.length === 0 && <p className="py-3 text-sm text-muted-foreground">{t("settings.aiResources.emptyResources")}</p>}
        {manageable && (
          <Button variant="ghost" size="sm" className="mt-1" onClick={() => onAddResource(connection)}>
            {t("settings.aiResources.addResource")}
          </Button>
        )}
      </div>
    </section>
  );
}

function filterResources(resources: ModelResource[], modality: string) {
  return modality === "all" ? resources : resources.filter((resource) => resource.modalities.includes(modality));
}
