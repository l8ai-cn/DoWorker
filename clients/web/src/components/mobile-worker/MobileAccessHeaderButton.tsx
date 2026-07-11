"use client";

import { useState } from "react";
import { Smartphone } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { PodMobileAccessDialog } from "@/components/mobile/PodMobileAccessDialog";
import { useCurrentOrg } from "@/stores/auth";
import { usePod } from "@/stores/pod";

export function MobileAccessHeaderButton({ podKey }: { podKey: string }) {
  const t = useTranslations("mobile.access");
  const [open, setOpen] = useState(false);
  const currentOrg = useCurrentOrg();
  const pod = usePod(podKey);

  return (
    <>
      <Button
        type="button"
        variant="ghost"
        size="sm"
        className="h-5 w-5 p-0 text-terminal-text hover:bg-terminal-bg-active"
        title={t("mobileAccess")}
        aria-label={t("mobileAccess")}
        disabled={!pod || !currentOrg?.slug}
        onClick={(event) => {
          event.stopPropagation();
          setOpen(true);
        }}
      >
        <Smartphone className="h-3 w-3" />
      </Button>
      <PodMobileAccessDialog
        open={open}
        onOpenChange={setOpen}
        orgSlug={currentOrg?.slug ?? ""}
        pod={pod ?? null}
      />
    </>
  );
}
