"use client";

import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { CredentialFormFields } from "../CredentialFormFields";
import { getCredentialFormSpecFromFields } from "../envBundleCredentialForms";
import type { CredentialField } from "@/lib/api";
import type { CredentialProfileViewModel } from "../_shared/credentialViewModel";
import type { CredentialBundleFormData } from "./types";

interface CredentialBundleDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  agentSlug: string;
  credentialFields: CredentialField[];
  editing: CredentialProfileViewModel | null;
  onSubmit: (
    data: CredentialBundleFormData,
    editing: CredentialProfileViewModel | null,
  ) => Promise<void>;
  t: (key: string) => string;
}

function buildData(
  values: Record<string, string>,
  removedKeys: string[],
): Record<string, string> {
  const data: Record<string, string> = {};
  for (const [key, value] of Object.entries(values)) {
    if (value.trim()) data[key] = value;
  }
  for (const key of removedKeys) data[key] = "";
  return data;
}

export function CredentialBundleDialog({
  open,
  onOpenChange,
  agentSlug,
  credentialFields,
  editing,
  onSubmit,
  t,
}: CredentialBundleDialogProps) {
  const spec = useMemo(
    () => getCredentialFormSpecFromFields(agentSlug, credentialFields),
    [agentSlug, credentialFields],
  );
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [values, setValues] = useState<Record<string, string>>({});
  const [selectedOneOf, setSelectedOneOf] = useState<Record<string, string>>({});
  const [removedKeys, setRemovedKeys] = useState<string[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setName(editing?.name ?? "");
    setDescription(editing?.description ?? "");
    setValues(editing?.configured_values ?? {});
    setSelectedOneOf({});
    setRemovedKeys([]);
    setError(null);
  }, [editing, open]);

  const handleSubmit = async () => {
    if (!name.trim()) return;
    try {
      setSubmitting(true);
      setError(null);
      await onSubmit(
        { name, description, data: buildData(values, removedKeys) },
        editing,
      );
      onOpenChange(false);
    } catch {
      setError(t("settings.agentConfig.credentialBundles.saveFailed"));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {editing
              ? t("settings.agentConfig.credentialBundles.editTitle")
              : t("settings.agentConfig.credentialBundles.addTitle")}
          </DialogTitle>
          <DialogDescription>
            {t("settings.agentConfig.credentialBundles.description")}
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 px-6 py-4">
          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="grid gap-2">
            <Label htmlFor="credential-bundle-name">
              {t("settings.agentConfig.credentialBundles.name")}
            </Label>
            <Input
              id="credential-bundle-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder={t("settings.agentConfig.credentialBundles.namePlaceholder")}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="credential-bundle-description">
              {t("settings.agentConfig.credentialBundles.descriptionLabel")}
            </Label>
            <Textarea
              id="credential-bundle-description"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              rows={2}
            />
          </div>
          <CredentialFormFields
            spec={spec}
            values={values}
            onValueChange={(key, value) => setValues((previous) => ({ ...previous, [key]: value }))}
            selectedOneOf={selectedOneOf}
            onOneOfChange={(group, key) => setSelectedOneOf((previous) => ({ ...previous, [group]: key }))}
            configuredKeys={editing?.configured_fields}
            removedKeys={removedKeys}
            onRemoveKey={(key) => setRemovedKeys((previous) => [...previous, key])}
            onRestoreKey={(key) => setRemovedKeys((previous) => previous.filter((item) => item !== key))}
            isEditing={editing !== null}
            t={t}
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("common.cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={submitting || !name.trim()}>
            {submitting ? t("common.saving") : editing ? t("common.save") : t("common.create")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
