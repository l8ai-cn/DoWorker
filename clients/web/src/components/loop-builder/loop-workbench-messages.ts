"use client";

import { useTranslations } from "next-intl";
import { useMemo } from "react";
import type {
  LoopMessageTranslator,
  LoopWorkbenchMessages,
} from "./loop-workbench-message-types";
import { createLoopAIMessages } from "./loop-ai-messages";

export type {
  LoopAIMessages,
  LoopAIProjectionMessages,
  LoopAIRepairMessages,
  LoopBlockCatalogMessages,
  LoopCustomBlockMessages,
  LoopMessageTranslator,
  LoopQuickInsertMessages,
  LoopRuntimeMessages,
  LoopStatusMessages,
  LoopToolbarMessages,
  LoopWorkbenchMessages,
} from "./loop-workbench-message-types";

const parseStatuses = ["empty", "parsing", "valid", "syntax-error"] as const;
const runStatuses = [
  "draft", "active", "paused", "verifying", "completed", "failed", "cancelled",
] as const;
const diagnosticKeys: Record<string, string> = {
  "loop.program.nil": "programNil",
  "loop.schema-version.unsupported": "schemaVersionUnsupported",
  "loop.syntax.invalid-token": "syntaxInvalidToken",
  "loop.syntax.invalid-number": "syntaxInvalidNumber",
  "loop.syntax.unexpected-token": "syntaxUnexpectedToken",
  "loop.syntax.unknown": "syntaxUnknown",
  "loop.node-id.missing": "nodeIdMissing",
  "loop.node-id.duplicate": "nodeIdDuplicate",
  "loop.local-id.duplicate": "localIdDuplicate",
  "loop.identifier.invalid": "identifierInvalid",
  "loop.structure.limits-count": "limitsCount",
  "loop.structure.repeat-count": "repeatCount",
  "loop.structure.agent-count": "agentCount",
  "loop.structure.verifier-count": "verifierCount",
  "loop.structure.failure-count": "failureCount",
  "loop.reference.until-invalid": "untilInvalid",
  "loop.repeat.max-exceeds-limit": "repeatMaxExceedsLimit",
  "loop.failure-policy.invalid": "failurePolicyInvalid",
  "loop.value.out-of-range": "valueOutOfRange",
  "loop.text.empty": "textEmpty",
  "loop.secret.literal-forbidden": "secretLiteralForbidden",
};

export function createLoopWorkbenchMessages(t: LoopMessageTranslator): LoopWorkbenchMessages {
  const parseLabels = Object.fromEntries(
    parseStatuses.map((status) => [status, t(`status.parseStatus.${status}`)]),
  );
  const runLabels = Object.fromEntries(
    runStatuses.map((status) => [status, t(`status.runStatus.${status}`)]),
  );
  const diagnostics = Object.fromEntries(
    Object.entries(diagnosticKeys).map(([code, key]) => [code, t(`status.diagnostics.${key}`)]),
  );
  const parseStatusLabel = (status: string) => parseLabels[status] ?? t("status.unknownStatus");
  const loopRunStatusLabel = (status: string) => runLabels[status] ?? t("status.unknownStatus");
  const diagnosticLabel = (code: string) =>
    diagnostics[code] ?? t("status.diagnostics.unknown");

  return {
    shell: {
      canvasTitle: t("shell.canvasTitle"),
      canvasHint: t("shell.canvasHint"),
      editorTitle: t("shell.editorTitle"),
      editorMetadata: (revision, semanticRevision) =>
        t("shell.editorMetadata", { revision, semanticRevision }),
    },
    toolbar: {
      back: t("toolbar.back"),
      title: t("toolbar.title"),
      subtitle: t("toolbar.subtitle"),
      blocks: t("toolbar.blocks"),
      code: t("toolbar.code"),
      run: t("toolbar.run"),
      parseStatusLabel,
    },
    blockly: {
      loop: {
        message0: t("blockly.loop.message0"), message1: t("blockly.loop.message1"),
        message2: t("blockly.loop.message2"), message3: t("blockly.loop.message3"),
      },
      limits: { message0: t("blockly.limits.message0"), message1: t("blockly.limits.message1") },
      repeat: {
        message0: t("blockly.repeat.message0"), message1: t("blockly.repeat.message1"),
        message2: t("blockly.repeat.message2"),
      },
      agent: {
        message0: t("blockly.agent.message0"),
        message1: t("blockly.agent.message1"),
        defaultPrompt: t("blockly.agent.defaultPrompt"),
      },
      verifier: {
        message0: t("blockly.verifier.message0"), message1: t("blockly.verifier.message1"),
        message2: t("blockly.verifier.message2"),
        defaultAccept: t("blockly.verifier.defaultAccept"),
      },
      failure: {
        message0: t("blockly.failure.message0"),
        pause: t("blockly.failure.pause"),
        fail: t("blockly.failure.fail"),
      },
      toolbox: {
        loop: t("blockly.toolbox.loop"), control: t("blockly.toolbox.control"),
        agent: t("blockly.toolbox.agent"), verifier: t("blockly.toolbox.verifier"),
        limits: t("blockly.toolbox.limits"), failure: t("blockly.toolbox.failure"),
        custom: t("blockly.toolbox.custom"),
      },
    },
    quickInsert: {
      close: t("quickInsert.close"),
      title: t("quickInsert.title"),
      createCustom: t("quickInsert.createCustom"),
      customEmpty: t("quickInsert.customEmpty"),
      customTitle: t("quickInsert.customTitle"),
      options: {
        loop: t("quickInsert.options.loop"), repeat: t("quickInsert.options.repeat"),
        agent: t("quickInsert.options.agent"), verifier: t("quickInsert.options.verifier"),
        limits: t("quickInsert.options.limits"), failure: t("quickInsert.options.failure"),
      },
    },
    status: {
      diagnosticsTitle: t("status.diagnosticsTitle"), runTitle: t("status.runTitle"),
      valid: t("status.valid"), noRun: t("status.noRun"), nodeLabel: t("status.nodeLabel"),
      repairDiagnostic: t("status.repairDiagnostic"),
      repairingDiagnostic: t("status.repairingDiagnostic"),
      runStatusLabel: t("status.runStatusLabel"), parseStatus: parseLabels,
      runStatus: runLabels, parseStatusLabel, loopRunStatusLabel,
      diagnosticLabel,
      diagnosticLocation: (line, column) => t("status.diagnosticLocation", { line, column }),
      runInstance: (podKey) => t("status.runInstance", { podKey }),
    },
    runtime: {
      title: t("runtime.title"), description: t("runtime.description"),
      field: t("runtime.field"), placeholder: t("runtime.placeholder"),
      unnamed: t("runtime.unnamed"), loading: t("runtime.loading"),
      retry: t("runtime.retry"), empty: t("runtime.empty"),
      cancel: t("runtime.cancel"), start: t("runtime.start"),
      snapshotLabel: (name, workerType, id) =>
        t("runtime.snapshotLabel", { name, workerType, id }),
    },
    customBlock: {
      title: t("customBlock.title"),
      description: t("customBlock.description"),
      label: t("customBlock.label"),
      slug: t("customBlock.slug"),
      promptTemplate: t("customBlock.promptTemplate"),
      commandTemplate: t("customBlock.commandTemplate"),
      acceptTemplate: t("customBlock.acceptTemplate"),
      cancel: t("customBlock.cancel"),
      create: t("customBlock.create"),
      duplicate: t("customBlock.duplicate"),
      required: t("customBlock.required"),
      identifier: t("customBlock.identifier"),
    },
    ai: createLoopAIMessages(t),
  };
}

export function useLoopWorkbenchMessages(): LoopWorkbenchMessages {
  const translate = useTranslations("loopWorkbench");
  return useMemo(
    () => createLoopWorkbenchMessages(
      (key, values) => translate(key as never, values as never),
    ),
    [translate],
  );
}
