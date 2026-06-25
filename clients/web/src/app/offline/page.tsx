"use client";

import { WifiOff, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useTranslations } from "next-intl";

export default function OfflinePage() {
  const t = useTranslations();

  return (
    <div className="azure-theme min-h-screen flex items-center justify-center bg-background p-4">
      <div className="text-center max-w-md">
        <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-muted flex items-center justify-center">
          <WifiOff className="w-10 h-10 text-muted-foreground" />
        </div>

        <h1 className="text-2xl font-bold mb-2">{t("offline.title")}</h1>
        <p className="text-muted-foreground mb-6">{t("offline.description")}</p>

        <Button onClick={() => window.location.reload()} className="gap-2">
          <RefreshCw className="w-4 h-4" />
          {t("offline.retry")}
        </Button>

        <div className="mt-8 surface-card p-4 text-sm text-muted-foreground">
          <p className="font-medium mb-2">{t("offline.whileOffline")}</p>
          <ul className="text-left space-y-1">
            <li>• {t("offline.tip1")}</li>
            <li>• {t("offline.tip2")}</li>
            <li>• {t("offline.tip3")}</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
