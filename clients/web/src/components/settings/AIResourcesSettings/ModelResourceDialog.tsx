"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Dialog, DialogBody, DialogContent, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { ResourceInput } from "@/lib/api";
import type { ModelResource, ProviderConnection, ProviderDefinition } from "./types";

const capabilityOptions = {
  chat: ["text-generation"],
  image: ["image-generation"],
  audio: ["speech-to-text", "text-to-speech"],
  video: ["video-generation"],
  embedding: ["embedding"],
  multimodal: ["text-generation", "vision-input"],
} as const;

interface ModelResourceDialogProps {
  connection: ProviderConnection | null;
  resource?: ModelResource;
  provider?: ProviderDefinition;
  onOpenChange: (open: boolean) => void;
  onSubmit: (connectionId: number, input: ResourceInput) => Promise<boolean>;
  onUpdate: (resourceId: number, input: Omit<ResourceInput, "identifier">) => Promise<boolean>;
}

export function ModelResourceDialog({ connection, resource, provider, onOpenChange, onSubmit, onUpdate }: ModelResourceDialogProps) {
  const t = useTranslations();
  const editing = Boolean(resource);
  const [identifier, setIdentifier] = useState("");
  const [modelId, setModelId] = useState(resource?.modelId ?? "");
  const [displayName, setDisplayName] = useState(resource?.displayName ?? "");
  const [selectedModalities, setSelectedModalities] = useState<string[]>(resource?.modalities ?? provider?.modalities.slice(0, 1) ?? []);
  const [selectedCapabilities, setSelectedCapabilities] = useState<string[]>(resource?.capabilities ?? []);
  const [saving, setSaving] = useState(false);
  const availableCapabilities = [...new Set(selectedModalities.flatMap((modality) => capabilityOptions[modality as keyof typeof capabilityOptions] ?? []))];
  const everyModalityHasCapability = selectedModalities.every((modality) =>
    (capabilityOptions[modality as keyof typeof capabilityOptions] ?? []).some((capability) => selectedCapabilities.includes(capability)),
  );
  const canSubmit = Boolean(
    connection
      && provider
      && (editing || identifier.trim())
      && modelId.trim()
      && displayName.trim()
      && selectedModalities.length
      && everyModalityHasCapability,
  );

  const reset = () => {
    setIdentifier("");
    setModelId(resource?.modelId ?? "");
    setDisplayName(resource?.displayName ?? "");
    setSelectedModalities(resource?.modalities ?? provider?.modalities.slice(0, 1) ?? []);
    setSelectedCapabilities(resource?.capabilities ?? []);
    setSaving(false);
  };

  const close = () => {
    reset();
    onOpenChange(false);
  };

  const changeModalities = (modalities: string[]) => {
    const allowedCapabilities = new Set<string>(modalities.flatMap((modality) => capabilityOptions[modality as keyof typeof capabilityOptions] ?? []));
    setSelectedModalities(modalities);
    setSelectedCapabilities((current) => current.filter((capability) => allowedCapabilities.has(capability)));
  };

  const submit = async () => {
    if (!connection || !canSubmit) return;
    setSaving(true);
    try {
      const saved = editing && resource
        ? await onUpdate(resource.id, { modelId, displayName, modalities: selectedModalities, capabilities: selectedCapabilities })
        : await onSubmit(connection.id, { identifier, modelId, displayName, modalities: selectedModalities, capabilities: selectedCapabilities });
      if (saved) close();
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={Boolean(connection)} onOpenChange={(open) => open ? onOpenChange(true) : close()}>
      <div>
        <DialogContent title={t(`settings.aiResources.resource.${editing ? "editTitle" : "createTitle"}`)}>
          <DialogBody className="space-y-4">
            <Field label={t("settings.aiResources.resource.name")} value={displayName} onChange={setDisplayName} />
            <Field label={t("settings.aiResources.resource.modelId")} value={modelId} onChange={setModelId} />
            {!editing && <Field label={t("settings.aiResources.resource.identifier")} value={identifier} onChange={setIdentifier} />}
            <CheckboxGroup
              label={t("settings.aiResources.resource.modalities")}
              values={provider?.modalities ?? []}
              selected={selectedModalities}
              onChange={changeModalities}
              labelFor={(modality) => t(`settings.aiResources.modality.${modality}`)}
            />
            <CheckboxGroup
              label={t("settings.aiResources.resource.capabilities")}
              values={availableCapabilities}
              selected={selectedCapabilities}
              onChange={setSelectedCapabilities}
              labelFor={(capability) => t(`settings.aiResources.capability.${capabilityLabelKey(capability)}`)}
            />
          </DialogBody>
          <DialogFooter><Button variant="outline" onClick={close}>{t("common.cancel")}</Button><Button disabled={!canSubmit || saving} onClick={() => void submit()}>{t(`settings.aiResources.resource.${editing ? "save" : "create"}`)}</Button></DialogFooter>
        </DialogContent>
      </div>
    </Dialog>
  );
}

function Field({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) {
  const id = `model-resource-${label}`;
  return <div className="space-y-2"><Label htmlFor={id}>{label}</Label><Input id={id} value={value} onChange={(event) => onChange(event.target.value)} /></div>;
}

function CheckboxGroup({ label, values, selected, onChange, labelFor }: {
  label: string;
  values: readonly string[];
  selected: string[];
  onChange: (values: string[]) => void;
  labelFor: (value: string) => string;
}) {
  return <fieldset><legend className="text-sm font-medium">{label}</legend><div className="mt-2 flex flex-wrap gap-3">{values.map((value) => <label key={value} className="flex items-center gap-2 text-sm"><input type="checkbox" checked={selected.includes(value)} onChange={() => onChange(selected.includes(value) ? selected.filter((item) => item !== value) : [...selected, value])} />{labelFor(value)}</label>)}</div></fieldset>;
}

function capabilityLabelKey(capability: string) {
  return {
    "text-generation": "textGeneration",
    "vision-input": "visionInput",
    "image-generation": "imageGeneration",
    "speech-to-text": "speechToText",
    "text-to-speech": "textToSpeech",
    "video-generation": "videoGeneration",
    embedding: "embedding",
  }[capability] ?? capability;
}
