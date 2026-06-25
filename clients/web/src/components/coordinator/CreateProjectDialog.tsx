"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogBody, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { repositoryApi, type RepositoryData } from "@/lib/api/facade/repository";
import { useCoordinatorStore } from "@/stores/coordinator";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateProjectDialog({ open, onOpenChange }: Props) {
  const t = useTranslations("automation");
  const createProject = useCoordinatorStore((s) => s.createProject);
  const [repos, setRepos] = useState<RepositoryData[]>([]);
  const [name, setName] = useState("");
  const [repositoryId, setRepositoryId] = useState("");
  const [labels, setLabels] = useState("");
  const [interval, setInterval] = useState("300");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (open) repositoryApi.list().then((r) => setRepos(r.items)).catch(() => setRepos([]));
  }, [open]);

  const submit = async () => {
    if (!name.trim() || !repositoryId) return;
    setSubmitting(true);
    try {
      await createProject({
        name: name.trim(),
        repository_id: Number(repositoryId),
        label_filter: labels.split(",").map((l) => l.trim()).filter(Boolean),
        scan_interval_seconds: Number(interval) || 300,
      });
      setName("");
      setRepositoryId("");
      setLabels("");
      onOpenChange(false);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent title={t("create.title")} description={t("create.description")}>
        <DialogBody className="space-y-4">
          <div className="space-y-1.5">
            <Label>{t("create.name")}</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder={t("create.namePlaceholder")} />
          </div>
          <div className="space-y-1.5">
            <Label>{t("create.repository")}</Label>
            <Select value={repositoryId} onValueChange={setRepositoryId}>
              <SelectTrigger>
                <SelectValue placeholder={t("create.repositoryPlaceholder")} />
              </SelectTrigger>
              <SelectContent>
                {repos.map((r) => (
                  <SelectItem key={r.id} value={String(r.id)}>
                    {r.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label>{t("create.labels")}</Label>
            <Input value={labels} onChange={(e) => setLabels(e.target.value)} placeholder={t("create.labelsPlaceholder")} />
          </div>
          <div className="space-y-1.5">
            <Label>{t("create.interval")}</Label>
            <Input type="number" value={interval} onChange={(e) => setInterval(e.target.value)} />
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("create.cancel")}
          </Button>
          <Button onClick={submit} loading={submitting} disabled={!name.trim() || !repositoryId}>
            {t("create.submit")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
