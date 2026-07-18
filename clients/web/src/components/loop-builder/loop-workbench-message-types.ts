type MessageValues = Record<string, string | number>;

export type LoopMessageTranslator = (
  key: string,
  values?: MessageValues,
) => string;

export interface LoopBlockCatalogMessages {
  loop: { message0: string; message1: string; message2: string; message3: string };
  limits: { message0: string; message1: string };
  repeat: { message0: string; message1: string; message2: string };
  agent: { message0: string; message1: string; defaultPrompt: string };
  verifier: {
    message0: string;
    message1: string;
    message2: string;
    defaultAccept: string;
  };
  failure: { message0: string; pause: string; fail: string };
  toolbox: Record<"loop" | "control" | "agent" | "verifier" | "limits" | "failure", string>;
}

export interface LoopQuickInsertMessages {
  close: string;
  title: string;
  options: Record<"loop" | "repeat" | "agent" | "verifier" | "limits" | "failure", string>;
}

export interface LoopToolbarMessages {
  back: string;
  title: string;
  subtitle: string;
  blocks: string;
  code: string;
  run: string;
  parseStatusLabel: (status: string) => string;
}

export interface LoopStatusMessages {
  diagnosticsTitle: string;
  runTitle: string;
  valid: string;
  noRun: string;
  repairDiagnostic: string;
  repairingDiagnostic: string;
  nodeLabel: string;
  runStatusLabel: string;
  parseStatus: Record<string, string>;
  runStatus: Record<string, string>;
  diagnosticLocation: (line: number, column: number) => string;
  runInstance: (podKey: string) => string;
  parseStatusLabel: (status: string) => string;
  loopRunStatusLabel: (status: string) => string;
  diagnosticLabel: (code: string) => string;
}

export interface LoopRuntimeMessages {
  title: string;
  description: string;
  field: string;
  placeholder: string;
  unnamed: string;
  loading: string;
  retry: string;
  empty: string;
  cancel: string;
  start: string;
  snapshotLabel: (name: string, workerType: string, id: string) => string;
}

export interface LoopAIProjectionMessages {
  title: string;
  unavailable: string;
  overview: string;
  schemaVersion: string;
  loopName: string;
  limits: string;
  iterations: string;
  tokens: string;
  timeout: string;
  noProgress: string;
  sameError: string;
  repeat: string;
  repeatName: string;
  repeatMax: string;
  until: string;
  agent: string;
  agentName: string;
  agentPrompt: string;
  verifier: string;
  verifierName: string;
  verifierCommand: string;
  verifierAccept: string;
  failure: string;
  failurePolicy: string;
  empty: string;
  iterationsValue: (value: string) => string;
  tokensValue: (value: string) => string;
  minutesValue: (value: string) => string;
  timesValue: (value: string) => string;
  failurePolicyLabel: (value: string) => string;
}

export interface LoopAIRepairMessages {
  title: string;
  description: string;
  diagnostic: string;
  field: string;
  prompt: string;
  promptPlaceholder: string;
  repair: string;
  repairing: string;
  error: string;
  patch: string;
  patchValue: (oldValue: string, newValue: string) => string;
}

export interface LoopAIMessages {
  toolbar: string;
  generateMode: string;
  explainMode: string;
  generateTitle: string;
  generateDescription: string;
  explainTitle: string;
  explainDescription: string;
  resource: string;
  resourcePlaceholder: string;
  prompt: string;
  promptPlaceholder: string;
  loadingResources: string;
  resourceError: string;
  retry: string;
  noResources: string;
  generate: string;
  generating: string;
  generationError: string;
  unchanged: string;
  stale: string;
  current: string;
  proposed: string;
  cancel: string;
  close: string;
  back: string;
  confirm: string;
  projection: LoopAIProjectionMessages;
  repair: LoopAIRepairMessages;
}

export interface LoopWorkbenchMessages {
  shell: {
    canvasTitle: string;
    canvasHint: string;
    editorTitle: string;
    editorMetadata: (revision: number, semanticRevision: number) => string;
  };
  toolbar: LoopToolbarMessages;
  blockly: LoopBlockCatalogMessages;
  quickInsert: LoopQuickInsertMessages;
  status: LoopStatusMessages;
  runtime: LoopRuntimeMessages;
  ai: LoopAIMessages;
}
