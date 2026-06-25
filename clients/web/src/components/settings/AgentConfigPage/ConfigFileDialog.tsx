"use client";

import { useState, useEffect } from "react";
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
import type { ConfigFile } from "@/lib/api";
import type { ConfigFileBundleViewModel, ConfigFileFormData } from "./types";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editing: ConfigFileBundleViewModel | null;
  fileSpecs: ConfigFile[];
  onSubmit: (data: ConfigFileFormData, editing: ConfigFileBundleViewModel | null) => Promise<void>;
  t: (key: string) => string;
}

const DEFAULT_JSON = `{
  "providers": {}
}`;

export function ConfigFileDialog({
  open,
  onOpenChange,
  editing,
  fileSpecs,
  onSubmit,
  t,
}: Props) {
  const [formName, setFormName] = useState("");
  const [formDescription, setFormDescription] = useState("");
  const [jsonContent, setJsonContent] = useState(DEFAULT_JSON);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setFormName(editing?.name ?? "");
    setFormDescription(editing?.description ?? "");
    setJsonContent(editing?.json_content ?? DEFAULT_JSON);
    setError(null);
  }, [open, editing]);

  const jsonInvalid = (() => {
    try {
      JSON.parse(jsonContent);
      return false;
    } catch {
      return true;
    }
  })();

  const handleSubmit = async () => {
    if (!formName.trim() || jsonInvalid) return;
    try {
      setSubmitting(true);
      setError(null);
      await onSubmit(
        { name: formName, description: formDescription, jsonContent },
        editing
      );
      onOpenChange(false);
    } catch (err) {
      console.error("Failed to save config file bundle:", err);
      setError(t("settings.agentConfig.configFiles.saveFailed"));
    } finally {
      setSubmitting(false);
    }
  };

  const hint = fileSpecs[0]?.path_hint;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {editing
              ? t("settings.agentConfig.configFiles.editTitle")
              : t("settings.agentConfig.configFiles.addTitle")}
          </DialogTitle>
          <DialogDescription>
            {hint
              ? t("settings.agentConfig.configFiles.formDescriptionWithPath", { path: hint })
              : t("settings.agentConfig.configFiles.formDescription")}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 px-6 py-4">
          {error && <div className="text-sm text-destructive">{error}</div>}

          <div className="grid gap-2">
            <Label htmlFor="config-name">{t("settings.agentConfig.configFiles.name")}</Label>
            <Input
              id="config-name"
              value={formName}
              onChange={(e) => setFormName(e.target.value)}
              placeholder={t("settings.agentConfig.configFiles.namePlaceholder")}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="config-desc">
              {t("settings.agentConfig.configFiles.descriptionLabel")}
            </Label>
            <Textarea
              id="config-desc"
              value={formDescription}
              onChange={(e) => setFormDescription(e.target.value)}
              rows={2}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="config-json">{t("settings.agentConfig.configFiles.jsonTitle")}</Label>
            <Textarea
              id="config-json"
              value={jsonContent}
              onChange={(e) => setJsonContent(e.target.value)}
              rows={12}
              className="font-mono text-sm"
            />
            {jsonInvalid && (
              <p className="text-sm text-destructive">{t("settings.agentConfig.invalidJson")}</p>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("common.cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={submitting || !formName.trim() || jsonInvalid}>
            {submitting ? t("common.saving") : editing ? t("common.save") : t("common.create")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
