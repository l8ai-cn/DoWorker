"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
} from "@/components/ui/dialog";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import {
  createKnowledgeBase,
  syncKnowledgeBase,
  type KnowledgeBase,
} from "@/lib/api/facade/knowledgeBaseApi";
import { SourceConfigFields } from "./SourceConfigFields";
import {
  KB_SOURCE_OPTIONS,
  buildSourceConfigJson,
  emptySourceConfig,
  isExternalSource,
  type KBSourceType,
  type SourceConfigForm,
  validateSourceConfig,
} from "./sourceConfig";

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
  const [sourceType, setSourceType] = useState<KBSourceType>("git");
  const [sourceConfig, setSourceConfig] = useState<SourceConfigForm>({});
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const resetForm = () => {
    setName("");
    setDescription("");
    setSourceType("git");
    setSourceConfig({});
    setError(null);
  };

  const handleSourceTypeChange = (next: KBSourceType) => {
    setSourceType(next);
    if (isExternalSource(next)) {
      setSourceConfig(emptySourceConfig(next));
    } else {
      setSourceConfig({});
    }
  };

  const handleSubmit = async () => {
    if (!name.trim()) return;
    if (isExternalSource(sourceType)) {
      const validationError = validateSourceConfig(sourceType, sourceConfig);
      if (validationError) {
        setError(validationError);
        return;
      }
    }

    setSubmitting(true);
    setError(null);
    try {
      const kb = await createKnowledgeBase(orgSlug, {
        name: name.trim(),
        description: description.trim() || undefined,
        sourceType: sourceType === "git" ? undefined : sourceType,
        sourceConfigJson: isExternalSource(sourceType)
          ? buildSourceConfigJson(sourceType, sourceConfig)
          : undefined,
      });

      if (isExternalSource(sourceType)) {
        try {
          await syncKnowledgeBase(orgSlug, kb.slug);
        } catch {
          // Create succeeded; first sync can be retried from detail page.
        }
      }

      resetForm();
      onOpenChange(false);
      onCreated(kb);
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建知识库失败");
    } finally {
      setSubmitting(false);
    }
  };

  const selectedSource = KB_SOURCE_OPTIONS.find((o) => o.value === sourceType);

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) resetForm();
        onOpenChange(next);
      }}
    >
      <DialogContent
        title="新建知识库"
        description="创建 Git 知识库，或绑定飞书/钉钉/Google 文档作为外部同步源。"
      >
        <DialogBody className="space-y-4">
          {error && (
            <p className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive" role="alert">
              {error}
            </p>
          )}
          <FormField label="名称" htmlFor="kb-name">
            <Input
              id="kb-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="例如：团队文档"
              maxLength={100}
              autoFocus
            />
          </FormField>
          <FormField label="描述（可选）" htmlFor="kb-description">
            <Textarea
              id="kb-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="这个知识库覆盖什么内容？"
              rows={3}
            />
          </FormField>
          <FormField label="数据源" htmlFor="kb-source-type">
            <Select value={sourceType} onValueChange={(v) => handleSourceTypeChange(v as KBSourceType)}>
              <SelectTrigger id="kb-source-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {KB_SOURCE_OPTIONS.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {selectedSource && (
              <p className="mt-1 text-xs text-muted-foreground">{selectedSource.description}</p>
            )}
          </FormField>
          {isExternalSource(sourceType) && (
            <SourceConfigFields
              sourceType={sourceType}
              value={sourceConfig}
              onChange={setSourceConfig}
            />
          )}
        </DialogBody>
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
