"use client";

import { useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { skillCatalogApi } from "@/lib/api";
import type { SkillImportAuthType } from "@/lib/api";
import { getLocalizedErrorMessage } from "@/lib/api/errors";
import { toast } from "sonner";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogBody, DialogFooter } from "@/components/ui/dialog";
import type { TranslationFn } from "../GeneralSettings";

const SUPPORTED_AGENTS = [
  { slug: "claude-code", label: "Claude Code" },
  { slug: "gemini-cli", label: "Gemini CLI" },
  { slug: "codex-cli", label: "Codex CLI" },
  { slug: "aider", label: "Aider" },
] as const;

interface ImportSkillDialogProps {
  t: TranslationFn;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onImported: () => void;
}

export function ImportSkillDialog({ t, open, onOpenChange, onImported }: ImportSkillDialogProps) {
  const [url, setUrl] = useState("");
  const [branch, setBranch] = useState("");
  const [subdir, setSubdir] = useState("");
  const [agents, setAgents] = useState<string[]>([]);
  const [authType, setAuthType] = useState<SkillImportAuthType>("none");
  const [authCredential, setAuthCredential] = useState("");
  const [importing, setImporting] = useState(false);

  const resetForm = useCallback(() => {
    setUrl("");
    setBranch("");
    setSubdir("");
    setAgents([]);
    setAuthType("none");
    setAuthCredential("");
  }, []);

  const handleImport = useCallback(async () => {
    if (!url.trim()) return;
    setImporting(true);
    try {
      const res = await skillCatalogApi.import({
        url: url.trim(),
        branch: branch.trim() || undefined,
        subdir: subdir.trim() || undefined,
        agent_filter: agents.length > 0 ? agents : undefined,
        auth_type: authType !== "none" ? authType : undefined,
        auth_credential: authCredential.trim() || undefined,
      });
      if (res.partial_errors) {
        toast.warning(t("extensions.skillCatalog.importedPartial", { count: res.imported }));
      } else {
        toast.success(t("extensions.skillCatalog.imported", { count: res.imported }));
      }
      onOpenChange(false);
      resetForm();
      onImported();
    } catch (error) {
      toast.error(getLocalizedErrorMessage(error, t, t("extensions.skillCatalog.failedToImport")));
    } finally {
      setImporting(false);
    }
  }, [url, branch, subdir, agents, authType, authCredential, t, onImported, onOpenChange, resetForm]);

  return (
    <Dialog open={open} onOpenChange={(o) => { onOpenChange(o); if (!o) resetForm(); }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("extensions.skillCatalog.import")}</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium mb-1 block">
                {t("extensions.repoUrl")} <span className="text-destructive">*</span>
              </label>
              <Input
                placeholder="https://github.com/owner/skills-repo"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
              />
              <p className="text-xs text-muted-foreground mt-1">
                {t("extensions.skillCatalog.importUrlHint")}
              </p>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm font-medium mb-1 block">{t("extensions.branch")}</label>
                <Input placeholder="main" value={branch} onChange={(e) => setBranch(e.target.value)} />
              </div>
              <div>
                <label className="text-sm font-medium mb-1 block">
                  {t("extensions.skillCatalog.subdir")}
                </label>
                <Input placeholder="skills/foo" value={subdir} onChange={(e) => setSubdir(e.target.value)} />
              </div>
            </div>

            <div>
              <label className="text-sm font-medium mb-1 block">
                {t("extensions.skillCatalog.compatibleAgents")}
              </label>
              <p className="text-xs text-muted-foreground mb-2">
                {t("extensions.skillCatalog.compatibleAgentsHint")}
              </p>
              <div className="flex flex-wrap gap-2">
                {SUPPORTED_AGENTS.map((agent) => {
                  const isSelected = agents.includes(agent.slug);
                  return (
                    <Button
                      key={agent.slug}
                      type="button"
                      variant={isSelected ? "default" : "outline"}
                      size="sm"
                      onClick={() =>
                        setAgents((prev) =>
                          isSelected ? prev.filter((s) => s !== agent.slug) : [...prev, agent.slug],
                        )
                      }
                    >
                      {agent.label}
                    </Button>
                  );
                })}
              </div>
            </div>

            <div className="border-t border-border pt-4">
              <label className="text-sm font-medium mb-1 block">
                {t("extensions.skillCatalog.authentication")}
              </label>
              <p className="text-xs text-muted-foreground mb-2">
                {t("extensions.skillCatalog.authenticationHint")}
              </p>
              <select
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={authType}
                onChange={(e) => {
                  setAuthType(e.target.value as SkillImportAuthType);
                  setAuthCredential("");
                }}
              >
                <option value="none">{t("extensions.skillCatalog.authNone")}</option>
                <option value="github_pat">{t("extensions.skillCatalog.authGitHubPAT")}</option>
                <option value="gitlab_pat">{t("extensions.skillCatalog.authGitLabPAT")}</option>
                <option value="ssh_key">{t("extensions.skillCatalog.authSSHKey")}</option>
              </select>

              {authType !== "none" && (
                <div className="mt-3">
                  <label className="text-sm font-medium mb-1 block">
                    {authType === "ssh_key"
                      ? t("extensions.skillCatalog.sshKeyLabel")
                      : t("extensions.skillCatalog.patLabel")}
                  </label>
                  {authType === "ssh_key" ? (
                    <textarea
                      className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono min-h-[100px] resize-y"
                      placeholder={t("extensions.skillCatalog.sshKeyPlaceholder")}
                      value={authCredential}
                      onChange={(e) => setAuthCredential(e.target.value)}
                      autoComplete="off"
                    />
                  ) : (
                    <Input
                      type="password"
                      placeholder={t("extensions.skillCatalog.patPlaceholder")}
                      value={authCredential}
                      onChange={(e) => setAuthCredential(e.target.value)}
                      autoComplete="off"
                    />
                  )}
                </div>
              )}
            </div>
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("common.cancel")}
          </Button>
          <Button disabled={importing || !url.trim()} onClick={handleImport}>
            {importing ? t("extensions.adding") : t("extensions.skillCatalog.import")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
