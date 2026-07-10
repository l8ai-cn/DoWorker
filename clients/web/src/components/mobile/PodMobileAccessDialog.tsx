"use client";

import { useMemo, useState } from "react";
import { Copy, ExternalLink, MonitorSmartphone } from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import {
  ResponsiveDialog,
  ResponsiveDialogBody,
  ResponsiveDialogContent,
  ResponsiveDialogDescription,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
} from "@/components/ui/responsive-dialog";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  buildPodMobileConsoleUrl,
  buildPodMobilePreviewUrl,
  podHasPreviewAccess,
} from "@/lib/pod-mobile-access";
import { getPodDisplayName } from "@/lib/pod-display-name";

interface PodMobileAccessDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgSlug: string;
  pod: unknown | null;
}

interface DisplayPod {
  pod_key: string;
  alias?: string | null;
  title?: string | null;
}

export function PodMobileAccessDialog({
  open,
  onOpenChange,
  orgSlug,
  pod,
}: PodMobileAccessDialogProps) {
  const t = useTranslations();
  const [tab, setTab] = useState("console");
  const displayPod = pod && typeof pod === "object"
    ? pod as DisplayPod
    : null;
  const podKey = displayPod
    ? String(displayPod.pod_key ?? "")
    : "";
  const previewEnabled = podHasPreviewAccess(pod);
  const urls = useMemo(() => ({
    console: buildPodMobileConsoleUrl(orgSlug, podKey),
    preview: buildPodMobilePreviewUrl(orgSlug, podKey),
  }), [orgSlug, podKey]);
  const selectedUrl = tab === "preview" && previewEnabled ? urls.preview : urls.console;

  if (!displayPod || !podKey) return null;

  const copySelectedUrl = () => {
    if (!navigator.clipboard) {
      toast.error(t("mobile.access.copyFailed"));
      return;
    }
    navigator.clipboard.writeText(selectedUrl).then(
      () => toast.success(t("mobile.access.copied")),
      () => toast.error(t("mobile.access.copyFailed")),
    );
  };

  return (
    <ResponsiveDialog open={open} onOpenChange={onOpenChange}>
      <ResponsiveDialogContent className="max-w-[440px]">
        <ResponsiveDialogHeader onClose={() => onOpenChange(false)} className="items-start">
          <div className="min-w-0">
            <ResponsiveDialogTitle className="flex items-center gap-2">
              <MonitorSmartphone className="h-5 w-5 text-primary" />
              {t("mobile.access.title")}
            </ResponsiveDialogTitle>
            <ResponsiveDialogDescription className="truncate">
              {getPodDisplayName(displayPod)}
            </ResponsiveDialogDescription>
          </div>
        </ResponsiveDialogHeader>
        <ResponsiveDialogBody className="space-y-4">
          {previewEnabled && (
            <Tabs value={tab} onValueChange={setTab}>
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="console">{t("mobile.access.console")}</TabsTrigger>
                <TabsTrigger value="preview">{t("mobile.access.preview")}</TabsTrigger>
              </TabsList>
            </Tabs>
          )}
          <div className="flex justify-center rounded-md border border-border/60 bg-surface-muted p-4">
            <QRCodeSVG
              value={selectedUrl}
              size={192}
              level="M"
              className="h-48 w-48 rounded bg-white p-2"
            />
          </div>
          <div className="rounded-md border border-border/60 bg-background px-3 py-2">
            <p className="break-all font-mono text-xs text-muted-foreground">{selectedUrl}</p>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <Button type="button" variant="outline" onClick={copySelectedUrl} className="gap-2">
              <Copy className="h-4 w-4" />
              {t("mobile.access.copy")}
            </Button>
            <a
              href={selectedUrl}
              target="_blank"
              rel="noreferrer"
              className="motion-interactive inline-flex h-9 items-center justify-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm hover:bg-primary-hover"
            >
              <ExternalLink className="h-4 w-4" />
              {t("mobile.access.open")}
            </a>
          </div>
        </ResponsiveDialogBody>
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
