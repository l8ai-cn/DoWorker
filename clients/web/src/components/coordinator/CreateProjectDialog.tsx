"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogBody, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { usePodCreationData } from "@/components/pod/hooks";
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
  const [agentSlug, setAgentSlug] = useState("");
  const [labels, setLabels] = useState("");
  const [interval, setInterval] = useState("300");
  const [submitting, setSubmitting] = useState(false);
  const { runners, availableAgents } = usePodCreationData(open);

  useEffect(() => {
    if (open) repositoryApi.list().then((r) => setRepos(r.items)).catch(() => setRepos([]));
  }, [open]);

  useEffect(() => {
    if (agentSlug && !availableAgents.some((agent) => agent.slug === agentSlug)) {
      setAgentSlug("");
    }
  }, [agentSlug, availableAgents]);

  const submit = async () => {
    if (!name.trim() || !repositoryId || !agentSlug) return;
    setSubmitting(true);
    try {
      await createProject({
        name: name.trim(),
        repository_id: Number(repositoryId),
        agent_slug: agentSlug,
        label_filter: labels.split(",").map((l) => l.trim()).filter(Boolean),
        scan_interval_seconds: Number(interval) || 300,
      });
      setName("");
      setRepositoryId("");
      setAgentSlug("");
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
            <Label>{t("create.agent")}</Label>
            <Select
              value={agentSlug}
              onValueChange={setAgentSlug}
              disabled={availableAgents.length === 0}
            >
              <SelectTrigger>
                <SelectValue placeholder={t("create.agentPlaceholder")} />
              </SelectTrigger>
              <SelectContent>
                {availableAgents.map((agent) => (
                  <SelectItem key={agent.slug} value={agent.slug}>
                    {agent.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {runners.length === 0 && (
              <p className="text-xs text-muted-foreground">{t("create.noOnlineRunners")}</p>
            )}
            {runners.length > 0 && availableAgents.length === 0 && (
              <p className="text-xs text-muted-foreground">{t("create.noCompatibleAgents")}</p>
            )}
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
          <Button onClick={submit} loading={submitting} disabled={!name.trim() || !repositoryId || !agentSlug}>
            {t("create.submit")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
