"use client";

import { useTranslations } from "next-intl";
import { DocsHorizontalScroll } from "./DocsHorizontalScroll";

export default function ArchitectureDiagram() {
  const t = useTranslations("docs.architecture");

  return (
    <DocsHorizontalScroll className="my-8" hintPosition="top">
      <div className="min-w-[600px] max-w-3xl mx-auto flex flex-col items-center gap-0">
        <LayerBox
          label={t("clientLayer")}
          color="blue"
          items={[
            { icon: "🖥️", text: "Web" },
            { icon: "📱", text: "Mobile" },
            { icon: "📟", text: "Tablet" },
          ]}
          subtitle={t("clientSubtitle")}
        />

        <Arrow label="HTTPS / WebSocket" />

        <div className="w-full border-2 border-success/40 rounded-xl bg-success/5 overflow-hidden">
          <div className="bg-success/10 px-4 py-1.5">
            <span className="text-[11px] font-semibold text-success uppercase tracking-wider">
              {t("cloudLayer")}
            </span>
          </div>
          <div className="p-4 pt-2">
            <div className="text-center mb-3">
              <span className="text-sm font-semibold text-success">
                Do Worker Cloud
              </span>
              <p className="text-xs text-muted-foreground mt-0.5">
                {t("cloudDesc")}
              </p>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-center text-xs">
              {(["orchestration", "monitoring", "collaboration", "security"] as const).map(
                (key) => (
                  <div
                    key={key}
                    className="bg-success/10 rounded-md px-2 py-1.5 font-medium text-success"
                  >
                    {t(`cloud.${key}`)}
                  </div>
                )
              )}
            </div>
          </div>
        </div>

        <div className="w-full grid grid-cols-2 gap-4 my-1">
          <div className="flex flex-col items-center">
            <div className="h-6 w-px border-l-2 border-dashed border-warning/60" />
            <div className="border border-warning/40 rounded-lg bg-warning/5 px-3 py-2.5 text-center w-full">
              <div className="text-[11px] font-bold text-warning uppercase tracking-wider mb-1">
                {t("controlPlane")}
              </div>
              <div className="text-xs text-muted-foreground">{t("controlPlaneDesc")}</div>
              <div className="mt-1.5 inline-flex items-center gap-1 rounded-full bg-warning/15 px-2.5 py-0.5 text-[10px] font-semibold text-warning">
                🔒 gRPC + mTLS
              </div>
            </div>
            <div className="h-6 w-px border-l-2 border-dashed border-warning/60" />
          </div>

          <div className="flex flex-col items-center">
            <div className="h-6 w-px border-l-2 border-dashed border-primary/60" />
            <div className="border border-primary/40 rounded-lg bg-primary/5 px-3 py-2.5 text-center w-full">
              <div className="text-[11px] font-bold text-primary uppercase tracking-wider mb-1">
                {t("dataPlane")}
              </div>
              <div className="text-xs text-muted-foreground">{t("dataPlaneDesc")}</div>
              <div className="mt-1.5 inline-flex items-center gap-1 rounded-full bg-primary/15 px-2.5 py-0.5 text-[10px] font-semibold text-primary">
                ⚡ Relay {t("cluster")}
              </div>
            </div>
            <div className="h-6 w-px border-l-2 border-dashed border-primary/60" />
          </div>
        </div>

        <div className="w-full border-2 border-info/40 rounded-xl bg-info/5 overflow-hidden">
          <div className="bg-info/10 px-4 py-1.5">
            <span className="text-[11px] font-semibold text-info uppercase tracking-wider">
              {t("runnerLayer")}
            </span>
          </div>
          <div className="p-4 pt-2">
            <div className="text-center mb-3">
              <span className="text-sm font-semibold text-info">
                {t("selfHostedRunners")}
              </span>
              <p className="text-xs text-muted-foreground mt-0.5">
                {t("runnerDesc")}
              </p>
            </div>
            <div className="grid grid-cols-3 gap-2 text-center text-xs">
              {[
                { icon: "🖥️", label: t("runnerMac") },
                { icon: "🐧", label: t("runnerLinux") },
                { icon: "☁️", label: t("runnerCloud") },
              ].map((r) => (
                <div
                  key={r.label}
                  className="bg-info/10 rounded-md px-2 py-1.5 font-medium text-info"
                >
                  {r.icon} {r.label}
                </div>
              ))}
            </div>
          </div>
        </div>

        <Arrow label="PTY + Sandbox + Git Worktree" />

        <LayerBox
          label={t("agentLayer")}
          color="rose"
          items={[
            { icon: "🤖", text: "Claude Code" },
            { icon: "🤖", text: "Codex CLI" },
            { icon: "🤖", text: "Gemini CLI" },
            { icon: "🤖", text: "Aider" },
          ]}
          subtitle={t("agentSubtitle")}
        />

        <div className="mt-4 w-full border border-warning/30 rounded-lg bg-warning/5 px-4 py-3">
          <div className="flex items-start gap-2">
            <span className="text-base mt-0.5">🔐</span>
            <div>
              <div className="text-xs font-semibold text-warning mb-1">
                {t("securityTitle")}
              </div>
              <p className="text-xs text-muted-foreground leading-relaxed">
                {t("securityDesc")}
              </p>
            </div>
          </div>
        </div>
      </div>
    </DocsHorizontalScroll>
  );
}

function Arrow({ label }: { label: string }) {
  return (
    <div className="flex flex-col items-center py-1">
      <div className="h-4 w-px bg-border" />
      <div className="text-[10px] text-muted-foreground font-medium px-2 py-0.5 rounded-full bg-muted/50 border border-border">
        {label}
      </div>
      <div className="h-2 w-px bg-border" />
      <div className="w-0 h-0 border-l-[5px] border-l-transparent border-r-[5px] border-r-transparent border-t-[6px] border-t-border" />
    </div>
  );
}

function LayerBox({
  label,
  color,
  items,
  subtitle,
}: {
  label: string;
  color: "blue" | "rose";
  items: { icon: string; text: string }[];
  subtitle: string;
}) {
  const colorMap = {
    blue: {
      border: "border-info/40",
      bg: "bg-info/5",
      headerBg: "bg-info/10",
      tag: "text-info",
      itemBg: "bg-info/10",
      itemText: "text-info",
    },
    rose: {
      border: "border-danger/40",
      bg: "bg-danger/5",
      headerBg: "bg-danger/10",
      tag: "text-danger",
      itemBg: "bg-danger/10",
      itemText: "text-danger",
    },
  };
  const c = colorMap[color];

  return (
    <div className={`w-full border-2 ${c.border} rounded-xl ${c.bg} overflow-hidden`}>
      <div className={`${c.headerBg} px-4 py-1.5`}>
        <span
          className={`text-[11px] font-semibold ${c.tag} uppercase tracking-wider`}
        >
          {label}
        </span>
      </div>
      <div className="p-4 pt-2">
        <div className="text-center mb-2">
          <p className="text-xs text-muted-foreground">{subtitle}</p>
        </div>
        <div
          className="grid gap-2 text-center text-xs"
          style={{ gridTemplateColumns: `repeat(${items.length}, 1fr)` }}
        >
          {items.map((item) => (
            <div
              key={item.text}
              className={`${c.itemBg} rounded-md px-2 py-1.5 font-medium ${c.itemText}`}
            >
              {item.icon} {item.text}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
