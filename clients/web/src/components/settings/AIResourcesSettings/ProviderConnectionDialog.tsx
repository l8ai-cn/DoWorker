"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Dialog, DialogBody, DialogContent, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { ConnectionInput } from "@/lib/api";
import { getAIResourceCredentialLabel } from "./aiResourceCredentialLabel";
import type { ProviderConnection, ProviderDefinition } from "./types";

interface ProviderConnectionDialogProps {
  open: boolean;
  catalog: ProviderDefinition[];
  connection?: ProviderConnection;
  onOpenChange: (open: boolean) => void;
  onSubmit: (input: ConnectionInput) => Promise<boolean>;
  onUpdate: (connectionId: number, input: { name: string; baseUrl: string }) => Promise<boolean>;
}

export function ProviderConnectionDialog({ open, catalog, connection, onOpenChange, onSubmit, onUpdate }: ProviderConnectionDialogProps) {
  const t = useTranslations();
  const editing = Boolean(connection);
  const [providerKey, setProviderKey] = useState(connection?.providerKey ?? "");
  const [name, setName] = useState(connection?.name ?? "");
  const [identifier, setIdentifier] = useState("");
  const [baseUrl, setBaseUrl] = useState(connection?.baseUrl ?? "");
  const [credentials, setCredentials] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);
  const provider = useMemo(() => catalog.find((item) => item.key === providerKey), [catalog, providerKey]);
  const canSubmit = Boolean(
    name.trim()
      && (editing || provider && identifier.trim() && provider.credentialFields.every((field) => !field.required || credentials[field.key]?.trim())),
  );

  const reset = () => {
    setProviderKey(connection?.providerKey ?? "");
    setName(connection?.name ?? "");
    setIdentifier("");
    setBaseUrl(connection?.baseUrl ?? "");
    setCredentials({});
    setSaving(false);
  };

  const close = () => {
    reset();
    onOpenChange(false);
  };

  const selectProvider = (key: string) => {
    const next = catalog.find((item) => item.key === key);
    setProviderKey(key);
    setBaseUrl(next?.defaultBaseUrl ?? "");
    setCredentials({});
  };

  const submit = async () => {
    if (!canSubmit) return;
    setSaving(true);
    try {
      const saved = editing && connection
        ? await onUpdate(connection.id, { name, baseUrl })
        : provider && await onSubmit({ identifier, providerKey, name, baseUrl, credentials });
      if (saved) close();
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(nextOpen) => nextOpen ? onOpenChange(true) : close()}>
      <div>
        <DialogContent title={t(`settings.aiResources.connection.${editing ? "editTitle" : "createTitle"}`)} description={editing ? undefined : t("settings.aiResources.connection.createDescription")}>
          <DialogBody className="space-y-4">
            <div className="space-y-2">
              <Label>{t("settings.aiResources.connection.provider")}</Label>
              {editing ? <p className="text-sm text-muted-foreground">{provider?.displayName ?? connection?.providerKey}</p> : <Select value={providerKey} onValueChange={selectProvider}>
                <SelectTrigger aria-label={t("settings.aiResources.connection.provider")}><SelectValue placeholder={t("settings.aiResources.connection.provider")} /></SelectTrigger>
                <SelectContent>{catalog.map((item) => <SelectItem key={item.key} value={item.key}>{item.displayName}</SelectItem>)}</SelectContent>
              </Select>}
            </div>
            <Field label={t("settings.aiResources.connection.name")} value={name} onChange={setName} />
            {!editing && <Field label={t("settings.aiResources.connection.identifier")} value={identifier} onChange={setIdentifier} />}
            <Field label={t("settings.aiResources.connection.baseUrl")} value={baseUrl} onChange={setBaseUrl} />
            {!editing && provider?.credentialFields.map((field) => (
              <Field key={field.key} label={getAIResourceCredentialLabel(field, t)} value={credentials[field.key] ?? ""} type={field.secret ? "password" : "text"} onChange={(value) => setCredentials((current) => ({ ...current, [field.key]: value }))} />
            ))}
          </DialogBody>
          <DialogFooter>
            <Button variant="outline" onClick={close}>{t("common.cancel")}</Button>
            <Button disabled={saving || !canSubmit} onClick={() => void submit()}>{t(`settings.aiResources.connection.${editing ? "save" : "create"}`)}</Button>
          </DialogFooter>
        </DialogContent>
      </div>
    </Dialog>
  );
}

function Field({ label, value, onChange, type = "text" }: { label: string; value: string; onChange: (value: string) => void; type?: string }) {
  const id = `ai-resource-${label}`;
  return <div className="space-y-2"><Label htmlFor={id}>{label}</Label><Input id={id} type={type} value={value} onChange={(event) => onChange(event.target.value)} /></div>;
}
