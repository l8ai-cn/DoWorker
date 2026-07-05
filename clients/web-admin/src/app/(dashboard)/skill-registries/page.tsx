"use client";

import { useState, useEffect } from "react";
import { Plus, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  listSkillRegistries,
  createSkillRegistry,
  syncSkillRegistry,
  deleteSkillRegistry,
  SkillRegistry,
} from "@/lib/api/admin";
import { SkillRegistriesTable } from "./skill-registries-table";

export default function SkillRegistriesPage() {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [formUrl, setFormUrl] = useState("");
  const [formBranch, setFormBranch] = useState("");
  const [formSourceType, setFormSourceType] = useState("");
  const [syncingIds, setSyncingIds] = useState<Set<number>>(new Set());
  const [isCreating, setIsCreating] = useState(false);

  const [data, setData] = useState<{ items: SkillRegistry[]; total: number } | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [refetchKey, setRefetchKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    listSkillRegistries()
      .then((result) => {
        if (cancelled) return;
        setData(result);
        setIsLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [refetchKey]);

  const triggerRefetch = () => setRefetchKey((k) => k + 1);

  const resetForm = () => {
    setFormUrl("");
    setFormBranch("");
    setFormSourceType("");
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formUrl.trim()) {
      toast.error("请输入仓库 URL");
      return;
    }
    setIsCreating(true);
    try {
      await createSkillRegistry({
        repository_url: formUrl.trim(),
        branch: formBranch.trim() || undefined,
        source_type: formSourceType.trim() || undefined,
      });
      toast.success("技能源已添加");
      setDialogOpen(false);
      resetForm();
      triggerRefetch();
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "创建技能源失败");
    } finally {
      setIsCreating(false);
    }
  };

  const handleSync = async (id: number) => {
    setSyncingIds((prev) => new Set(prev).add(id));
    try {
      await syncSkillRegistry(id);
      toast.success("同步已触发");
      triggerRefetch();
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "同步技能源失败");
    } finally {
      setSyncingIds((prev) => { const next = new Set(prev); next.delete(id); return next; });
    }
  };

  const handleDelete = async (registry: SkillRegistry) => {
    if (!confirm(`确定要删除 "${registry.repository_url}" 吗？此操作无法撤销。`)) return;
    try {
      await deleteSkillRegistry(registry.id);
      toast.success("技能源已删除");
      triggerRefetch();
    } catch (err: unknown) {
      toast.error((err as { error?: string })?.error || "删除技能源失败");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">技能源</h1>
          <p className="text-sm text-muted-foreground">
            管理平台级技能仓库，同步后的技能可供所有组织使用。
          </p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={(open) => { setDialogOpen(open); if (!open) resetForm(); }}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              添加技能源
            </Button>
          </DialogTrigger>
          <DialogContent>
            <form onSubmit={handleCreate}>
              <DialogHeader>
                <DialogTitle>添加技能源</DialogTitle>
                <DialogDescription>
                  添加新的技能仓库，创建后会自动同步技能。
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                  <Label htmlFor="repository_url">仓库 URL</Label>
                  <Input
                    id="repository_url"
                    placeholder="https://github.com/org/repo"
                    value={formUrl}
                    onChange={(e) => setFormUrl(e.target.value)}
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="branch">分支</Label>
                  <Input id="branch" placeholder="main" value={formBranch} onChange={(e) => setFormBranch(e.target.value)} />
                  <p className="text-xs text-muted-foreground">留空则使用默认分支。</p>
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="source_type">来源类型</Label>
                  <Input id="source_type" placeholder="auto-detect" value={formSourceType} onChange={(e) => setFormSourceType(e.target.value)} />
                  <p className="text-xs text-muted-foreground">留空则自动识别。</p>
                </div>
              </div>
              <DialogFooter>
                <Button type="submit" disabled={isCreating}>
                  {isCreating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  添加技能源
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <SkillRegistriesTable
        registries={data?.items || []}
        isLoading={isLoading}
        syncingIds={syncingIds}
        onSync={handleSync}
        onDelete={handleDelete}
      />
    </div>
  );
}
