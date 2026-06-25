"use client";

import { useEffect } from "react";
import { useTranslations } from "next-intl";
import { relayPool } from "@/stores/relayConnection";
import { useDoAgentGoals } from "@/stores/doagentConsole";
import { doagentRpc, doagentControl } from "./doagentControl";
import { Button } from "@/components/ui/button";

// Fetch the goal list once per session, as soon as the relay is connected.
// Goal updates after that arrive via dispatchDoAgentRelayEvent -> setGoals.
// We must NOT re-fire on every relay response: goal/list's own response would
// re-trigger this effect, causing an infinite request/render loop.
export function useDoAgentGoalSync(podKey: string, active: boolean): void {
  useEffect(() => {
    if (!active || !podKey) return;
    let done = false;
    const tryFetch = () => {
      if (done || !relayPool.isConnected(podKey)) return;
      done = true;
      doagentRpc(podKey, "goal/list", {});
    };
    tryFetch();
    const iv = setInterval(tryFetch, 1500);
    return () => {
      done = true;
      clearInterval(iv);
    };
  }, [podKey, active]);
}

interface DoAgentGoalBarProps {
  podKey: string;
}

export function DoAgentGoalBar({ podKey }: DoAgentGoalBarProps) {
  const t = useTranslations("doagent");
  const goals = useDoAgentGoals(podKey);
  const active = goals.find((g) => g.status !== "complete" && g.status !== "stopped") ?? goals[0];

  const refresh = () => doagentRpc(podKey, "goal/list", {});

  return (
    <div className="flex items-center gap-2 panel-lift px-3 py-2 text-xs">
      <span className="text-muted-foreground">{t("goal.label")}</span>
      {active ? (
        <>
          <span className="truncate font-medium" title={active.title}>
            {active.title || active.id}
          </span>
          <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] uppercase">
            {active.status ?? t("goal.unknown")}
          </span>
          <Button
            variant="ghost"
            size="sm"
            className="h-6 px-2 text-[10px]"
            onClick={() => doagentControl(podKey, "goal/pause", { goalId: active.id })}
          >
            {t("goal.pause")}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="h-6 px-2 text-[10px]"
            onClick={() => doagentControl(podKey, "goal/resume", { goalId: active.id })}
          >
            {t("goal.resume")}
          </Button>
        </>
      ) : (
        <span className="text-muted-foreground">{t("goal.none")}</span>
      )}
      <Button variant="outline" size="sm" className="ml-auto h-6 px-2 text-[10px]" onClick={refresh}>
        {t("goal.refresh")}
      </Button>
    </div>
  );
}
