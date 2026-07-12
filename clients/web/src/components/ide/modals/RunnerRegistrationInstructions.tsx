"use client";

import { useState } from "react";
import { AlertCircle, Check, Terminal } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";

interface RunnerRegistrationInstructionsProps {
  command: string;
  onDone: () => void;
}

export function RunnerRegistrationInstructions({
  command,
  onDone,
}: RunnerRegistrationInstructionsProps) {
  const t = useTranslations();
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const serverUrl = command.match(/(?:^|\s)--server\s+(\S+)/)?.[1];

  const copyText = (text: string, key: string) => {
    navigator.clipboard.writeText(text);
    setCopiedKey(key);
    setTimeout(() => setCopiedKey(null), 2000);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-start gap-2 p-3 bg-warning-bg border border-warning/30 rounded-lg">
        <AlertCircle className="w-5 h-5 text-warning flex-shrink-0 mt-0.5" />
        <p className="text-sm text-warning">
          {t("runners.addRunnerModal.tokenWarning")}
        </p>
      </div>

      {serverUrl && (
        <InstallInstructions
          serverUrl={serverUrl}
          copiedKey={copiedKey}
          onCopy={copyText}
          copyLabel={t("runners.addRunnerModal.copyCommand")}
        />
      )}

      <CommandPanel
        label={t("runners.addRunnerModal.usageTitle")}
        command={command}
        copyId="command"
        copiedKey={copiedKey}
        onCopy={copyText}
        copyLabel={t("runners.addRunnerModal.copyCommand")}
      />
      <CommandPanel
        label={t("runners.addRunnerModal.serviceTitle")}
        hint={t("runners.addRunnerModal.serviceHint")}
        command={
          "do-worker-runner service install\ndo-worker-runner service start"
        }
        copyId="service"
        copiedKey={copiedKey}
        onCopy={copyText}
        copyLabel={t("runners.addRunnerModal.copyCommand")}
      />

      <div className="flex justify-end pt-2">
        <Button onClick={onDone}>{t("runners.addRunnerModal.done")}</Button>
      </div>
    </div>
  );
}

function InstallInstructions({
  serverUrl,
  copiedKey,
  onCopy,
  copyLabel,
}: {
  serverUrl: string;
  copiedKey: string | null;
  onCopy: (text: string, key: string) => void;
  copyLabel: string;
}) {
  const t = useTranslations();
  return (
    <div>
      <label className="block text-sm font-medium mb-1">
        {t("runners.addRunnerModal.installTitle")}
      </label>
      <div className="space-y-2">
        <p className="text-xs text-muted-foreground">
          {t("runners.addRunnerModal.installHint")}
        </p>
        <CommandPanel
          label="# macOS / Linux"
          command={`curl -fsSL ${serverUrl}/install.sh | sh`}
          copyId="install-mac"
          copiedKey={copiedKey}
          onCopy={onCopy}
          copyLabel={copyLabel}
          terminal={false}
        />
        <CommandPanel
          label="# Windows (PowerShell)"
          command={`irm ${serverUrl}/install.ps1 | iex`}
          copyId="install-win"
          copiedKey={copiedKey}
          onCopy={onCopy}
          copyLabel={copyLabel}
          terminal={false}
        />
      </div>
    </div>
  );
}

function CommandPanel({
  label,
  hint,
  command,
  copyId,
  copiedKey,
  onCopy,
  copyLabel,
  terminal = true,
}: {
  label: string;
  hint?: string;
  command: string;
  copyId: string;
  copiedKey: string | null;
  onCopy: (text: string, key: string) => void;
  copyLabel: string;
  terminal?: boolean;
}) {
  return (
    <div>
      <label className="block text-sm font-medium mb-1">{label}</label>
      {hint && <p className="text-xs text-muted-foreground mb-2">{hint}</p>}
      <div className="bg-muted rounded-lg p-4 relative">
        {terminal && (
          <div className="flex items-center gap-2 text-muted-foreground text-xs mb-2">
            <Terminal className="w-4 h-4" />
            <span>Terminal</span>
          </div>
        )}
        <code className="text-success text-sm font-mono block whitespace-pre-wrap pr-24">
          {command}
        </code>
        <CopyButton
          text={command}
          id={copyId}
          copiedKey={copiedKey}
          onCopy={onCopy}
          label={copyLabel}
        />
      </div>
    </div>
  );
}

function CopyButton({
  text,
  id,
  copiedKey,
  onCopy,
  label,
}: {
  text: string;
  id: string;
  copiedKey: string | null;
  onCopy: (text: string, key: string) => void;
  label: string;
}) {
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={() => onCopy(text, id)}
      className="absolute top-2 right-2 h-7 text-xs text-muted-foreground hover:text-foreground"
    >
      {copiedKey === id ? <Check className="w-3 h-3 text-success" /> : label}
    </Button>
  );
}
