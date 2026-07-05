"use client";

import { useState, useEffect } from "react";
import { Search, Plus, ChevronLeft, ChevronRight } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { listSSOConfigs, enableSSOConfig, disableSSOConfig, deleteSSOConfig, createSSOConfig, updateSSOConfig, testSSOConfig, SSOConfig, SSOProtocol, CreateSSOConfigRequest } from "@/lib/api/sso";
import { SSOFormDialog } from "./sso-form-dialog";
import { SSOTable } from "./sso-table";
import { SSODeleteDialog } from "./sso-delete-dialog";

export default function SSOPage() {
  const [search, setSearch] = useState("");
  const [protocolFilter, setProtocolFilter] = useState<string>("all");
  const [configs, setConfigs] = useState<SSOConfig[]>([]);
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const [refetchKey, setRefetchKey] = useState(0);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingConfig, setEditingConfig] = useState<SSOConfig | null>(null);
  const [deletingConfig, setDeletingConfig] = useState<SSOConfig | null>(null);

  useEffect(() => {
    let cancelled = false;
    listSSOConfigs({
      search: search || undefined,
      protocol: protocolFilter !== "all" ? (protocolFilter as SSOProtocol) : undefined,
      page,
      page_size: pageSize,
    })
      .then((result) => {
        if (cancelled) return;
        setConfigs(result.data || []);
        setTotal(result.total || 0);
        setIsLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [search, protocolFilter, page, refetchKey]);

  const triggerRefetch = () => {
    setIsLoading(true);
    setRefetchKey((k) => k + 1);
  };

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const handleFormSubmit = async (data: CreateSSOConfigRequest) => {
    try {
      if (editingConfig) {
        await updateSSOConfig(editingConfig.id, data);
        toast.success("SSO 配置已更新");
      } else {
        await createSSOConfig(data);
        toast.success("SSO 配置已创建");
      }
      triggerRefetch();
    } catch (err: unknown) {
      const message = (err as { error?: string })?.error || "保存 SSO 配置失败";
      toast.error(message);
      throw err;
    }
  };

  const handleEnable = async (id: number) => {
    try {
      await enableSSOConfig(id);
      toast.success("SSO 配置已启用");
      triggerRefetch();
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "启用 SSO 配置失败");
    }
  };

  const handleDisable = async (id: number) => {
    try {
      await disableSSOConfig(id);
      toast.success("SSO 配置已停用");
      triggerRefetch();
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "停用 SSO 配置失败");
    }
  };

  const handleDeleteConfirm = async () => {
    if (!deletingConfig) return;
    try {
      await deleteSSOConfig(deletingConfig.id);
      toast.success("SSO 配置已删除");
      setDeletingConfig(null);
      triggerRefetch();
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "删除 SSO 配置失败");
    }
  };

  const handleTest = async (config: SSOConfig) => {
    try {
      const result = await testSSOConfig(config.id);
      if (result.success) {
        toast.success(result.message || "连接测试通过");
      } else {
        toast.error(result.error || result.message || "连接测试失败");
      }
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "连接测试失败");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">单点登录</h1>
          <p className="text-sm text-muted-foreground">
            管理域名的单点登录配置
          </p>
        </div>
        <Button onClick={() => { setEditingConfig(null); setDialogOpen(true); }}>
          <Plus className="mr-2 h-4 w-4" />
          创建 SSO 配置
        </Button>
      </div>

      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="按域名或名称搜索..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1); }}
            className="pl-10"
          />
        </div>
        <Select value={protocolFilter} onValueChange={(value) => { setProtocolFilter(value); setPage(1); }}>
          <SelectTrigger className="w-40">
            <SelectValue placeholder="全部协议" displayValue={protocolFilter === "all" ? "全部协议" : protocolFilter.toUpperCase()} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部协议</SelectItem>
            <SelectItem value="oidc">OIDC</SelectItem>
            <SelectItem value="saml">SAML</SelectItem>
            <SelectItem value="ldap">LDAP</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <SSOTable
        configs={configs}
        isLoading={isLoading}
        onEdit={(config) => { setEditingConfig(config); setDialogOpen(true); }}
        onTest={handleTest}
        onEnable={handleEnable}
        onDisable={handleDisable}
        onDelete={setDeletingConfig}
      />

      {totalPages > 1 && (
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm text-muted-foreground">
            显示第 {(page - 1) * pageSize + 1} 到 {Math.min(page * pageSize, total)} 条，共 {total} 个配置
          </p>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="icon" onClick={() => setPage(page - 1)} disabled={page <= 1}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="text-sm">第 {page} / {totalPages} 页</span>
            <Button variant="outline" size="icon" onClick={() => setPage(page + 1)} disabled={page >= totalPages}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}

      <SSOFormDialog open={dialogOpen} onOpenChange={setDialogOpen} config={editingConfig} onSubmit={handleFormSubmit} />
      <SSODeleteDialog config={deletingConfig} onOpenChange={() => setDeletingConfig(null)} onConfirm={handleDeleteConfirm} />
    </div>
  );
}
