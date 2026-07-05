"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  createKnowledgeBase,
  type KnowledgeBase,
} from "@/lib/api/facade/knowledgeBaseApi";

interface CreateKnowledgeBaseDialogProps {
  orgSlug: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated: (kb: KnowledgeBase) => void;
}

export function CreateKnowledgeBaseDialog({
  orgSlug,
  open,
  onOpenChange,
  onCreated,
}: CreateKnowledgeBaseDialogProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    if (!name.trim()) return;
    setSubmitting(true);
    setError(null);
    try {
      const kb = await createKnowledgeBase(orgSlug, {
        name: name.trim(),
        description: description.trim() || undefined,
      });
      setName("");
      setDescription("");
      onOpenChange(false);
      onCreated(kb);
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建知识库失败");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        title="新建知识库"
        description="将创建一个带 llms.txt / AGENTS.md / raw / wiki 标准布局的 Git 仓库。"
      >
        <div className="space-y-4">
          <div>
            <label htmlFor="kb-name" className="mb-1 block text-sm font-medium">
              名称
            </label>
            <Input
              id="kb-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="例如：团队文档"
              maxLength={100}
              autoFocus
            />
          </div>
          <div>
            <label htmlFor="kb-description" className="mb-1 block text-sm font-medium">
              描述（可选）
            </label>
            <Textarea
              id="kb-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="这个知识库覆盖什么内容？"
              rows={3}
            />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={submitting}>
            取消
          </Button>
          <Button onClick={handleSubmit} loading={submitting} disabled={!name.trim()}>
            创建
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
