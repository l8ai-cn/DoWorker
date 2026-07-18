import type { AgentConnectionStatus, AgentSessionStatus } from "./contracts";
import type { ArtifactWorkspaceText } from "./artifactWorkspaceText";
import type { ToolActivityCount } from "./toolActivityGroupText";
import type { VideoTaskStatusText } from "./videoTaskStatusText";
import {
  chineseAgentWorkspaceText,
  englishAgentWorkspaceText,
} from "./agentWorkspaceTextValues";

export type AgentWorkspaceLocale = "en-US" | "zh-CN";
export interface AgentWorkspaceText {
  conversation: string;
  results: string;
  artifacts: string;
  workspaceViews: string;
  terminal: string;
  plan: string;
  agentPlan: string;
  loadEarlierActivity: string;
  readyForTask: string;
  startSession: string;
  you: string;
  agent: string;
  system: string;
  agentic: string;
  terminalMode: string;
  approvals: string;
  messageAgent: string;
  readOnly: string;
  stopAgent: string;
  sendMessage: string;
  addAttachment: string;
  removeAttachment: string;
  uploadingAttachment: string;
  reject: string;
  approve: string;
  submitAnswers: string;
  customAnswer: string;
  details: string;
  takeControl: string;
  releaseControl: string;
  taskFailed: string;
  unsupportedToolPreview: string;
  rawToolEvidence: string;
  videoTaskStatus: VideoTaskStatusText;
  artifact: ArtifactWorkspaceText;
  generatedArtifact: string;
  emptyHeading(agentLabel: string): string;
  composerPlaceholder(agentLabel: string): string;
  requiresArgument(commandLabel: string): string;
  slashCommandAttachmentsUnsupported: string;
  customAnswerFor(prompt: string): string;
  loadingArtifact(filename: string): string;
  previewArtifact(filename: string): string;
  videoPreview(filename: string): string;
  openArtifact(filename: string): string;
  downloadArtifact(filename: string): string;
  configurationOptions(label: string): string;
  configurationLabel(id: string, fallback: string): string;
  configurationOption(id: string, value: string, fallback: string): string;
  sessionStatus(
    status: AgentSessionStatus,
    connection: AgentConnectionStatus,
  ): string;
  activityStatus(
    status: "pending" | "running" | "completed" | "failed",
  ): string;
  artifactType(kind: string, fallback: string): string;
  toolText(value: string): string;
  toolActivityGroupSummary(counts: ToolActivityCount[]): string;
}

export function agentWorkspaceText(
  locale: AgentWorkspaceLocale,
): AgentWorkspaceText {
  return locale === "zh-CN"
    ? chineseAgentWorkspaceText
    : englishAgentWorkspaceText;
}
