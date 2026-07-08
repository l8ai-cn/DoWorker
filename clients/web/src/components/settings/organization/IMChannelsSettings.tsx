"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ConfirmDialog, useConfirmDialog } from "@/components/ui/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Copy, Plus, Trash2 } from "lucide-react";
import {
  createIMConnection,
  deleteIMConnection,
  IM_CONFIG_EXAMPLES,
  listIMConnections,
  listIMProviders,
  pollWeixinQRLogin,
  startWeixinQRLogin,
  updateIMConnection,
  type IMConnection,
  type IMProviderMeta,
  type IMProviderType,
} from "@/lib/api/imChannelApi";
import type { TranslationFn } from "./GeneralSettings";

interface IMChannelsSettingsProps {
  t: TranslationFn;
}

const STATUS_VARIANT: Record<string, "default" | "secondary" | "destructive"> = {
  active: "default",
  disabled: "secondary",
  error: "destructive",
};

function isWeixinProvider(type: string) {
  return type === "weixin" || type === "wechat";
}

function isWeixinLoggedIn(conn: IMConnection) {
  const cfg = conn.config as Record<string, unknown> | undefined;
  return Boolean(cfg?.bot_token);
}

export function IMChannelsSettings({ t }: IMChannelsSettingsProps) {
  const [providers, setProviders] = useState<IMProviderMeta[]>([]);
  const [connections, setConnections] = useState<IMConnection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  const [formProvider, setFormProvider] = useState<IMProviderType>("feishu");
  const [formName, setFormName] = useState("");
  const [formChannelId, setFormChannelId] = useState("");
  const [formConfig, setFormConfig] = useState("");
  const [qrSessionId, setQrSessionId] = useState<string | null>(null);
  const [qrImageUrl, setQrImageUrl] = useState<string>("");
  const [qrStatus, setQrStatus] = useState<string>("");
  const [qrMessage, setQrMessage] = useState<string>("");
  const [qrLoading, setQrLoading] = useState(false);
  const { dialogProps, confirm } = useConfirmDialog();

  const providerLabel = useMemo(() => {
    const map = new Map(providers.map((p) => [p.type, p.display_name]));
    return (type: string) => map.get(type as IMProviderType) ?? type;
  }, [providers]);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [providerRows, connectionRows] = await Promise.all([
        listIMProviders(),
        listIMConnections(),
      ]);
      setProviders(providerRows);
      setConnections(connectionRows);
    } catch (err) {
      console.error("Failed to load IM channels:", err);
      setError(t("settings.imChannels.loadError"));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const openCreate = () => {
    setFormProvider("feishu");
    setFormName("");
    setFormChannelId("");
    setFormConfig(JSON.stringify(IM_CONFIG_EXAMPLES.feishu, null, 2));
    setShowCreate(true);
  };

  const onProviderChange = (value: IMProviderType) => {
    setFormProvider(value);
    const key = isWeixinProvider(value) ? "weixin" : value;
    setFormConfig(JSON.stringify(IM_CONFIG_EXAMPLES[key as IMProviderType] ?? {}, null, 2));
  };

  const handleCreate = async () => {
    setCreating(true);
    setError(null);
    try {
      const config = isWeixinProvider(formProvider)
        ? {}
        : (JSON.parse(formConfig) as Record<string, unknown>);
      await createIMConnection({
        provider: isWeixinProvider(formProvider) ? "weixin" : formProvider,
        name: formName.trim(),
        channel_id: formChannelId ? Number(formChannelId) : undefined,
        config,
        status: "disabled",
      });
      setShowCreate(false);
      await refresh();
    } catch (err) {
      console.error("Failed to create IM connection:", err);
      setError(t("settings.imChannels.createFailed"));
    } finally {
      setCreating(false);
    }
  };

  const toggleStatus = async (conn: IMConnection) => {
    const next = conn.status === "active" ? "disabled" : "active";
    try {
      await updateIMConnection(conn.id, { status: next });
      await refresh();
    } catch (err) {
      console.error("Failed to update IM connection:", err);
      setError(t("settings.imChannels.updateFailed"));
    }
  };

  const handleDelete = async (conn: IMConnection) => {
    const ok = await confirm({
      title: t("settings.imChannels.deleteDialog.title"),
      description: t("settings.imChannels.deleteDialog.description", { name: conn.name }),
      variant: "destructive",
      confirmText: t("settings.imChannels.deleteDialog.confirm"),
      cancelText: t("settings.imChannels.deleteDialog.cancel"),
    });
    if (!ok) return;
    try {
      await deleteIMConnection(conn.id);
      await refresh();
    } catch (err) {
      console.error("Failed to delete IM connection:", err);
      setError(t("settings.imChannels.deleteFailed"));
    }
  };

  const copyWebhook = async (url?: string) => {
    if (!url) return;
    await navigator.clipboard.writeText(url);
  };

  const startWeixinLogin = async (conn: IMConnection) => {
    setQrLoading(true);
    setQrMessage("");
    setQrStatus("");
    try {
      const resp = await startWeixinQRLogin(conn.id);
      setQrSessionId(resp.session_id);
      setQrImageUrl(resp.qrcode_url ?? "");
      setQrStatus(resp.status);
      setQrMessage(t("settings.imChannels.weixin.scanHint"));
    } catch (err) {
      console.error("Failed to start weixin QR login:", err);
      setError(t("settings.imChannels.weixin.loginFailed"));
    } finally {
      setQrLoading(false);
    }
  };

  useEffect(() => {
    if (!qrSessionId || qrStatus === "confirmed" || qrStatus === "failed" || qrStatus === "timed_out") {
      return;
    }
    const timer = window.setInterval(async () => {
      try {
        const resp = await pollWeixinQRLogin(qrSessionId);
        setQrStatus(resp.status);
        if (resp.qrcode_url) setQrImageUrl(resp.qrcode_url);
        if (resp.message) setQrMessage(resp.message);
        if (resp.status === "confirmed") {
          setQrMessage(t("settings.imChannels.weixin.loginSuccess"));
          await refresh();
          window.setTimeout(() => setQrSessionId(null), 1500);
        }
        if (resp.status === "failed" || resp.status === "timed_out") {
          setQrMessage(resp.message ?? t("settings.imChannels.weixin.loginFailed"));
        }
      } catch (err) {
        console.error("Weixin QR poll failed:", err);
      }
    }, 1200);
    return () => window.clearInterval(timer);
  }, [qrSessionId, qrStatus, refresh, t]);

  return (
    <div className="space-y-6">
      {error && (
        <div
          role="alert"
          className="bg-destructive/10 border border-destructive text-destructive px-4 py-3 rounded-lg flex items-center justify-between"
        >
          <span>{error}</span>
          <button onClick={() => setError(null)} className="text-sm underline">
            {t("settings.imChannels.dismiss")}
          </button>
        </div>
      )}

      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-lg font-semibold">{t("settings.imChannels.title")}</h2>
          <p className="text-sm text-muted-foreground mt-1">
            {t("settings.imChannels.description")}
          </p>
        </div>
        <Button onClick={openCreate}>
          <Plus className="w-4 h-4 mr-2" />
          {t("settings.imChannels.create")}
        </Button>
      </div>

      <div className="surface-card p-6 space-y-4">
        {loading ? (
          <p className="text-sm text-muted-foreground">{t("settings.imChannels.loading")}</p>
        ) : connections.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("settings.imChannels.empty")}</p>
        ) : (
          connections.map((conn) => (
            <div
              key={conn.id}
              className="border rounded-lg p-4 space-y-3"
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="font-medium">{conn.name}</h3>
                    <Badge variant={STATUS_VARIANT[conn.status] ?? "secondary"}>
                      {t(`settings.imChannels.status.${conn.status}`)}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">
                    {providerLabel(conn.provider)}
                    {conn.channel_id ? ` · ${t("settings.imChannels.boundChannel", { id: conn.channel_id })}` : ""}
                  </p>
                  {conn.last_error && (
                    <p className="text-sm text-destructive mt-2">{conn.last_error}</p>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  {isWeixinProvider(conn.provider) && !isWeixinLoggedIn(conn) && (
                    <Button variant="outline" size="sm" onClick={() => startWeixinLogin(conn)} disabled={qrLoading}>
                      {t("settings.imChannels.weixin.qrLogin")}
                    </Button>
                  )}
                  <Button variant="outline" size="sm" onClick={() => toggleStatus(conn)}>
                    {conn.status === "active"
                      ? t("settings.imChannels.disable")
                      : t("settings.imChannels.enable")}
                  </Button>
                  <Button variant="ghost" size="icon" onClick={() => handleDelete(conn)}>
                    <Trash2 className="w-4 h-4 text-destructive" />
                  </Button>
                </div>
              </div>

              {isWeixinProvider(conn.provider) ? (
                <p className="text-xs text-muted-foreground">
                  {isWeixinLoggedIn(conn)
                    ? t("settings.imChannels.weixin.loggedIn")
                    : t("settings.imChannels.weixin.loginRequired")}
                </p>
              ) : conn.webhook_url ? (
                <div className="space-y-1">
                  <Label className="text-xs text-muted-foreground">
                    {t("settings.imChannels.webhookUrl")}
                  </Label>
                  <div className="flex gap-2">
                    <Input readOnly value={conn.webhook_url} className="font-mono text-xs" />
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => copyWebhook(conn.webhook_url)}
                      title={t("settings.imChannels.copyWebhook")}
                    >
                      <Copy className="w-4 h-4" />
                    </Button>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {t("settings.imChannels.webhookHint")}
                  </p>
                </div>
              ) : null}
            </div>
          ))
        )}
      </div>

      <Dialog open={showCreate} onOpenChange={setShowCreate}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{t("settings.imChannels.createDialog.title")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>{t("settings.imChannels.createDialog.provider")}</Label>
              <Select value={formProvider} onValueChange={(v) => onProviderChange(v as IMProviderType)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {providers.map((p) => (
                    <SelectItem key={p.type} value={p.type}>
                      {p.display_name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>{t("settings.imChannels.createDialog.name")}</Label>
              <Input value={formName} onChange={(e) => setFormName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label>{t("settings.imChannels.createDialog.channelId")}</Label>
              <Input
                value={formChannelId}
                onChange={(e) => setFormChannelId(e.target.value)}
                placeholder={t("settings.imChannels.createDialog.channelIdPlaceholder")}
              />
            </div>
            {!isWeixinProvider(formProvider) && (
              <div className="space-y-2">
                <Label>{t("settings.imChannels.createDialog.config")}</Label>
                <Textarea
                  value={formConfig}
                  onChange={(e) => setFormConfig(e.target.value)}
                  rows={10}
                  className="font-mono text-xs"
                />
              </div>
            )}
            {isWeixinProvider(formProvider) && (
              <p className="text-sm text-muted-foreground">
                {t("settings.imChannels.weixin.createHint")}
              </p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowCreate(false)}>
              {t("settings.imChannels.createDialog.cancel")}
            </Button>
            <Button onClick={handleCreate} disabled={creating || !formName.trim()}>
              {t("settings.imChannels.createDialog.submit")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={Boolean(qrSessionId)} onOpenChange={(open) => !open && setQrSessionId(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t("settings.imChannels.weixin.qrTitle")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 text-center">
            {qrImageUrl ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={qrImageUrl}
                alt="WeChat QR"
                className="mx-auto w-56 h-56 border rounded-lg"
              />
            ) : (
              <p className="text-sm text-muted-foreground">{t("settings.imChannels.loading")}</p>
            )}
            <p className="text-sm text-muted-foreground">{qrMessage}</p>
            <p className="text-xs text-muted-foreground">{qrStatus}</p>
          </div>
        </DialogContent>
      </Dialog>

      <ConfirmDialog {...dialogProps} />
    </div>
  );
}
