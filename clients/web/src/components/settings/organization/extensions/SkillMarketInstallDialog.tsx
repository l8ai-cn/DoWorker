"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";
import type { SkillMarketItem } from "@/lib/api";
import { getLocalizedErrorMessage } from "@/lib/api/errors";
import { installSkillFromMarket } from "@/lib/api/facade/repoSkillExtension";
import { useCurrentOrg } from "@/stores/auth";
import { RepositorySelect } from "@/components/common/RepositorySelect";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import type { TranslationFn } from "../GeneralSettings";

interface SkillMarketInstallDialogProps {
  item: SkillMarketItem | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  t: TranslationFn;
}

export function SkillMarketInstallDialog({
  item,
  open,
  onOpenChange,
  t,
}: SkillMarketInstallDialogProps) {
  const currentOrg = useCurrentOrg();
  const orgSlug = currentOrg?.slug ?? "";
  const [repositoryId, setRepositoryId] = useState<number | null>(null);
  const [scope, setScope] = useState<"org" | "user">("user");
  const [installing, setInstalling] = useState(false);

  const handleOpenChange = (next: boolean) => {
    if (!next) {
      setRepositoryId(null);
      setScope("user");
    }
    onOpenChange(next);
  };

  const handleInstall = async () => {
    if (!item || !orgSlug || repositoryId == null) return;
    setInstalling(true);
    try {
      await installSkillFromMarket(orgSlug, repositoryId, {
        marketItemId: item.id,
        scope,
      });
      toast.success(t("extensions.installed"));
      handleOpenChange(false);
    } catch (error) {
      toast.error(getLocalizedErrorMessage(error, t, t("extensions.failedToInstall")));
    } finally {
      setInstalling(false);
    }
  };

  const skillName = item?.display_name || item?.slug || "";

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{t("extensions.skillMarket.installTitle")}</DialogTitle>
        </DialogHeader>
        <DialogBody className="space-y-4">
          <p className="text-sm text-muted-foreground">
            {t("extensions.skillMarket.installDescription", { name: skillName })}
          </p>
          <div>
            <label className="text-sm font-medium mb-1 block">
              {t("extensions.skillMarket.targetRepository")}
            </label>
            <RepositorySelect
              value={repositoryId}
              onChange={(id) => setRepositoryId(id)}
              placeholder={t("extensions.skillMarket.selectRepository")}
              disabled={installing}
            />
          </div>
          <div>
            <label className="text-sm font-medium mb-1 block">
              {t("extensions.skillMarket.installScope")}
            </label>
            <select
              className="w-full px-3 py-2 border border-border rounded-md bg-background"
              value={scope}
              onChange={(e) => setScope(e.target.value as "org" | "user")}
              disabled={installing}
            >
              <option value="user">{t("extensions.myInstalled")}</option>
              <option value="org">{t("extensions.orgInstalled")}</option>
            </select>
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="outline" onClick={() => handleOpenChange(false)} disabled={installing}>
            {t("extensions.cancel")}
          </Button>
          <Button onClick={handleInstall} disabled={installing || repositoryId == null}>
            {installing ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
                {t("extensions.installing")}
              </>
            ) : (
              t("extensions.install")
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
