"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  AlertCircle,
  Copy,
  ExternalLink,
  Loader2,
  MonitorSmartphone,
  RefreshCw,
} from "lucide-react";
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
  getMobileAccessDescriptor,
  type MobileAccessDescriptor,
} from "@/lib/api/facade/podConnect";
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
  const [descriptor, setDescriptor] =
    useState<MobileAccessDescriptor | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const displayPod = pod && typeof pod === "object"
    ? pod as DisplayPod
    : null;
  const podKey = displayPod
    ? String(displayPod.pod_key ?? "")
    : "";
  const loadDescriptor = useCallback(async () => {
    setLoading(true);
    setError(null);
    setDescriptor(null);
    try {
      setDescriptor(await getMobileAccessDescriptor(orgSlug, podKey));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
    } finally {
      setLoading(false);
    }
  }, [orgSlug, podKey]);

  useEffect(() => {
    if (!open || !orgSlug || !podKey) return;
    setTab("console");
    void loadDescriptor();
  }, [loadDescriptor, open, orgSlug, podKey]);

  const previewEnabled = descriptor?.preview_available ?? false;
  const selectedUrl = useMemo(() => {
    if (!descriptor) return "";
    return tab === "preview" && previewEnabled
      ? `${descriptor.canonical_url.replace(/\/$/, "")}/preview`
      : descriptor.canonical_url;
  }, [descriptor, previewEnabled, tab]);

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
          {loading && (
            <div className="flex min-h-64 items-center justify-center">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          )}
          {!loading && error && (
            <div className="flex min-h-64 items-center justify-center text-center">
              <div className="space-y-3">
                <AlertCircle className="mx-auto h-8 w-8 text-danger" />
                <p className="text-sm font-medium">
                  {t("mobile.access.loadFailed")}
                </p>
                <p className="max-w-sm text-xs text-muted-foreground">
                  {error}
                </p>
                <Button
                  type="button"
                  variant="outline"
                  className="h-11"
                  onClick={() => void loadDescriptor()}
                >
                  <RefreshCw className="h-4 w-4" />
                  {t("mobile.access.retry")}
                </Button>
              </div>
            </div>
          )}
          {!loading && descriptor && (
            <>
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
                <Button
                  type="button"
                  variant="outline"
                  disabled={!descriptor.console_available}
                  onClick={copySelectedUrl}
                  className="h-11 gap-2"
                >
                  <Copy className="h-4 w-4" />
                  {t("mobile.access.copy")}
                </Button>
                <a
                  href={selectedUrl}
                  target="_blank"
                  rel="noreferrer"
                  aria-disabled={!descriptor.console_available}
                  className="motion-interactive inline-flex h-11 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground shadow-sm hover:bg-primary-hover aria-disabled:pointer-events-none aria-disabled:opacity-50"
                >
                  <ExternalLink className="h-4 w-4" />
                  {t("mobile.access.open")}
                </a>
              </div>
              {!descriptor.console_available && (
                <p className="text-center text-xs text-warning">
                  {t("mobile.access.unavailable")}
                </p>
              )}
            </>
          )}
        </ResponsiveDialogBody>
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
