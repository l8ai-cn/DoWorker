"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Dialog, DialogBody, DialogContent, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { ProviderConnection, ProviderDefinition } from "./types";

interface ConnectionCredentialsDialogProps {
  connection: ProviderConnection | null;
  provider?: ProviderDefinition;
  onOpenChange: (open: boolean) => void;
  onSubmit: (connectionId: number, credentials: Record<string, string>) => Promise<boolean>;
}

export function ConnectionCredentialsDialog({ connection, provider, onOpenChange, onSubmit }: ConnectionCredentialsDialogProps) {
  const t = useTranslations();
  const [credentials, setCredentials] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);
  const canSubmit = Boolean(provider && provider.credentialFields.every((field) => !field.required || credentials[field.key]?.trim()));

  const close = () => {
    setCredentials({});
    setSaving(false);
    onOpenChange(false);
  };

  const submit = async () => {
    if (!connection || !canSubmit) return;
    setSaving(true);
    try {
      if (await onSubmit(connection.id, credentials)) close();
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={Boolean(connection)} onOpenChange={(open) => open ? onOpenChange(true) : close()}>
      <div>
        <DialogContent title={t("settings.aiResources.connection.rotateTitle")} description={t("settings.aiResources.connection.rotateDescription")}>
          <DialogBody className="space-y-4">
            {provider?.credentialFields.map((field) => {
              const id = `rotate-credential-${field.key}`;
              return <div key={field.key} className="space-y-2"><Label htmlFor={id}>{field.label}</Label><Input id={id} type={field.secret ? "password" : "text"} value={credentials[field.key] ?? ""} onChange={(event) => setCredentials((current) => ({ ...current, [field.key]: event.target.value }))} /></div>;
            })}
          </DialogBody>
          <DialogFooter>
            <Button variant="outline" onClick={close}>{t("common.cancel")}</Button>
            <Button disabled={saving || !canSubmit} onClick={() => void submit()}>{t("settings.aiResources.connection.rotate")}</Button>
          </DialogFooter>
        </DialogContent>
      </div>
    </Dialog>
  );
}
