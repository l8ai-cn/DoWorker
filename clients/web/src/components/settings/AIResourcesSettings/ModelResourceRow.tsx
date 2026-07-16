import { useState } from "react";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { aiResourceValidationMessage } from "./aiResourceValidationMessage";
import type { ModelResource } from "./types";

interface ModelResourceRowProps {
  resource: ModelResource;
  activeModality: string;
  canManage: boolean;
  onEnabledChange: (resourceId: number, enabled: boolean) => Promise<boolean>;
  onSetDefault: (resourceId: number, modality: string) => Promise<boolean>;
  onEdit: (resource: ModelResource) => void;
  onDelete: (resource: ModelResource) => void;
}

export function ModelResourceRow({
  resource,
  activeModality,
  canManage,
  onEnabledChange,
  onSetDefault,
  onEdit,
  onDelete,
}: ModelResourceRowProps) {
  const t = useTranslations();
  const [selectedDefaultModality, setSelectedDefaultModality] = useState("");
  const defaultModality = activeModality === "all"
    ? resource.modalities.length === 1 ? resource.modalities[0] : selectedDefaultModality
    : resource.modalities.includes(activeModality) ? activeModality : "";
  const validationMessage = resource.validationError
    ? aiResourceValidationMessage(resource.validationError, t)
    : "";

  return (
    <div className="flex flex-wrap items-center gap-x-3 gap-y-2 border-t border-border/60 py-3 first:border-t-0">
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-foreground">{resource.displayName}</p>
        <p className="truncate text-xs text-muted-foreground">{resource.modelId}</p>
        {validationMessage && (
          <p role="alert" className="text-xs text-destructive">{validationMessage}</p>
        )}
      </div>
      <div className="flex flex-wrap items-center gap-1.5">
        {resource.modalities.map((modality) => <Badge key={modality} variant="outline">{modality}</Badge>)}
        {resource.defaultModalities.length > 0 && <Badge variant="secondary">{t("settings.aiResources.default", { modalities: resource.defaultModalities.join(", ") })}</Badge>}
        {resource.status === "invalid" && <Badge variant="destructive">{t("settings.aiResources.status.invalid")}</Badge>}
        {!resource.isEnabled && <Badge variant="warning">{t("settings.aiResources.status.disabled")}</Badge>}
      </div>
      {canManage && (
        <div className="flex items-center gap-2">
          {activeModality === "all" && resource.modalities.length > 1 && (
            <Select value={selectedDefaultModality} onValueChange={setSelectedDefaultModality}>
              <SelectTrigger aria-label={t("settings.aiResources.defaultModality", { name: resource.displayName })} className="h-8 w-28"><SelectValue /></SelectTrigger>
              <SelectContent>{resource.modalities.map((item) => <SelectItem key={item} value={item}>{t(`settings.aiResources.modality.${item}`)}</SelectItem>)}</SelectContent>
            </Select>
          )}
          <Button variant="outline" size="sm" disabled={!defaultModality} onClick={() => void onSetDefault(resource.id, defaultModality)}>
            {t("settings.aiResources.setDefault")}
          </Button>
          <Button variant="ghost" size="sm" aria-label={`${t("settings.aiResources.resource.edit")}: ${resource.displayName}`} onClick={() => onEdit(resource)}>
            {t("settings.aiResources.resource.edit")}
          </Button>
          <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" aria-label={`${t("settings.aiResources.resource.delete")}: ${resource.displayName}`} onClick={() => onDelete(resource)}>
            {t("settings.aiResources.resource.delete")}
          </Button>
          <Switch
            aria-label={`${t("settings.aiResources.resource.enabled")}: ${resource.displayName}`}
            checked={resource.isEnabled}
            onCheckedChange={(enabled) => void onEnabledChange(resource.id, enabled)}
          />
        </div>
      )}
    </div>
  );
}
