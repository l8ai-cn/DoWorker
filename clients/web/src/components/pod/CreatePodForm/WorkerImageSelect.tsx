"use client";

import { cn } from "@/lib/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
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

  const selectedImage = images.find((image) => image.slug === selectedImageSlug);

  return (
    <div>
      <label
        htmlFor="worker-image-select"
        className="block text-sm font-medium mb-2"
      >
        {t("ide.createPod.selectImage")}
      </label>
      <Select
        value={selectedImageSlug || ""}
        onValueChange={(value) => onSelect(value || null)}
      >
        <SelectTrigger
          id="worker-image-select"
          aria-required="true"
          aria-invalid={!!error}
          aria-describedby={error ? "worker-image-error" : "worker-image-hint"}
          className={cn(error && "ring-destructive/60 focus:ring-destructive/40")}
        >
          <span className={cn(!selectedImageSlug && "text-muted-foreground")}>
            {selectedImage?.name ?? t("ide.createPod.selectImagePlaceholder")}
          </span>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="">{t("ide.createPod.selectImagePlaceholder")}</SelectItem>
          {images.map((image) => (
            <SelectItem key={image.slug} value={image.slug}>
              {image.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
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
