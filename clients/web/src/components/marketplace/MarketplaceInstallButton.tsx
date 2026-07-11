"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { ArrowRight, Loader2 } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { useLightSession } from "@/hooks/useLightSession";
import { fetchFirstOrgSlug, lightFetch } from "@/lib/light-auth";
import { updateLightSessionOrgSlug } from "@/lib/light-session";

interface InstallResponse {
  expert: { slug: string };
  already_installed: boolean;
}

export function MarketplaceInstallButton({ applicationSlug }: { applicationSlug: string }) {
  const router = useRouter();
  const { session, hydrated } = useLightSession();

  if (!hydrated) {
    return (
      <Button className="w-full gap-2" disabled>
        <Loader2 className="h-4 w-4 animate-spin" />
        检查账户
      </Button>
    );
  }

  if (!session?.isAuthenticated) {
    return (
      <Button
        className="w-full gap-2"
        onClick={() => router.push("/login?redirect=%2Fmarketplace")}
      >
        登录后启用
        <ArrowRight className="h-4 w-4" />
      </Button>
    );
  }

  return (
    <InstallAction
      applicationSlug={applicationSlug}
      orgSlug={session.currentOrgSlug}
      onInstalled={(orgSlug, expertSlug, alreadyInstalled) => {
        toast.success(alreadyInstalled ? "专家应用已在组织中" : "专家应用已启用");
        router.push(`/${orgSlug}/experts/${expertSlug}`);
      }}
      onNeedsOrganization={() => router.push("/onboarding/create-org")}
    />
  );
}

function InstallAction({
  applicationSlug,
  orgSlug,
  onInstalled,
  onNeedsOrganization,
}: {
  applicationSlug: string;
  orgSlug: string | null;
  onInstalled: (orgSlug: string, expertSlug: string, alreadyInstalled: boolean) => void;
  onNeedsOrganization: () => void;
}) {
  const [installing, setInstalling] = useState(false);

  async function install() {
    setInstalling(true);
    try {
      const targetOrgSlug = orgSlug || (await fetchFirstOrgSlug());
      if (!targetOrgSlug) {
        setInstalling(false);
        onNeedsOrganization();
        return;
      }
      updateLightSessionOrgSlug(targetOrgSlug);
      const result = await lightFetch<InstallResponse>(
        `/api/v1/orgs/${targetOrgSlug}/marketplace/experts/${applicationSlug}/install`,
        { method: "POST", authenticated: true },
      );
      onInstalled(targetOrgSlug, result.expert.slug, result.already_installed);
    } catch {
      toast.error("启用失败，请稍后重试");
      setInstalling(false);
    }
  }

  return (
    <Button className="w-full gap-2" onClick={install} disabled={installing}>
      {installing ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
      {installing ? "正在启用" : "立即启用"}
      {!installing ? <ArrowRight className="h-4 w-4" /> : null}
    </Button>
  );
}
