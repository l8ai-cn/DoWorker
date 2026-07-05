"use client";

import type { AgentData } from "@/lib/api";

interface WorkerImageSelectProps {
  images: AgentData[];
  selectedImageSlug: string | null;
  onSelect: (imageSlug: string | null) => void;
  hasOnlineClusters?: boolean;
  error?: string;
  t: (key: string) => string;
}

export function WorkerImageSelect({
  images,
  selectedImageSlug,
  onSelect,
  hasOnlineClusters = true,
  error,
  t,
}: WorkerImageSelectProps) {
  if (!hasOnlineClusters) {
    return (
      <div>
        <label className="block text-sm font-medium mb-2">
          {t("ide.createPod.selectImage")}
        </label>
        <p className="text-sm text-muted-foreground py-2">
          {t("ide.createPod.noOnlineRunnersHint")}
        </p>
      </div>
    );
  }

  if (images.length === 0) {
    return (
      <div>
        <label className="block text-sm font-medium mb-2">
          {t("ide.createPod.selectImage")}
        </label>
        <p className="text-sm text-muted-foreground py-2">
          {t("ide.createPod.noImagesForCluster")}
        </p>
      </div>
    );
  }

  return (
    <div>
      <label
        htmlFor="worker-image-select"
        className="block text-sm font-medium mb-2"
      >
        {t("ide.createPod.selectImage")}
      </label>
      <select
        id="worker-image-select"
        className={`w-full px-3 py-2 border rounded-md bg-background ${
          error ? "border-destructive" : "border-border"
        }`}
        value={selectedImageSlug || ""}
        onChange={(e) => onSelect(e.target.value || null)}
        aria-required="true"
        aria-invalid={!!error}
        aria-describedby={error ? "worker-image-error" : "worker-image-hint"}
      >
        <option value="">{t("ide.createPod.selectImagePlaceholder")}</option>
        {images.map((image) => (
          <option key={image.slug} value={image.slug}>
            {image.name}
          </option>
        ))}
      </select>
      {error ? (
        <p id="worker-image-error" className="text-xs text-destructive mt-1">
          {error}
        </p>
      ) : (
        <p id="worker-image-hint" className="text-xs text-muted-foreground mt-1">
          {t("ide.createPod.imageHint")}
        </p>
      )}
    </div>
  );
}
