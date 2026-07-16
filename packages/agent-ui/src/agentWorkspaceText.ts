import type { AgentConnectionStatus, AgentSessionStatus } from "./contracts";
import {
  localizeConfigurationLabel,
  localizeConfigurationOption,
} from "./configurationLocalization";
import {
  chineseActivityStatus,
  chineseSessionStatus,
  englishActivityStatus,
  englishSessionStatus,
} from "./agentStatusText";
import {
  englishFileChangeVerb,
  localizeFileChangeVerb,
  localizeToolText,
} from "./toolLocalization";
import {
  chineseToolActivityGroupSummary,
  englishToolActivityGroupSummary,
  type ToolActivityCount,
} from "./toolActivityGroupText";
import {
  artifactWorkspaceText,
  type ArtifactWorkspaceText,
} from "./artifactWorkspaceText";

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
  reject: string;
  approve: string;
  submitAnswers: string;
  customAnswer: string;
  details: string;
  takeControl: string;
  releaseControl: string;
  artifact: ArtifactWorkspaceText;
  emptyHeading(agentLabel: string): string;
  composerPlaceholder(agentLabel: string): string;
  requiresArgument(commandLabel: string): string;
  customAnswerFor(prompt: string): string;
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
  fileChangeVerb(kind: string): string;
}

export function agentWorkspaceText(
  locale: AgentWorkspaceLocale,
): AgentWorkspaceText {
  return locale === "zh-CN" ? zhCN : enUS;
}

const enUS: AgentWorkspaceText = {
  conversation: "Conversation",
  results: "Results",
  artifacts: "Artifacts",
  workspaceViews: "Workspace views",
  terminal: "Terminal",
  plan: "Plan",
  agentPlan: "Agent plan",
  loadEarlierActivity: "Load earlier activity",
  readyForTask: "Ready for a task",
  startSession: "Send a message to start this agent session.",
  you: "You",
  agent: "Agent",
  system: "System",
  agentic: "Agentic",
  terminalMode: "Terminal mode",
  approvals: "Approvals",
  messageAgent: "Message the agent",
  readOnly: "This session is read-only",
  stopAgent: "Stop agent",
  sendMessage: "Send message",
  reject: "Reject",
  approve: "Approve",
  submitAnswers: "Submit answers",
  customAnswer: "Custom answer",
  details: "Details",
  takeControl: "Take control",
  releaseControl: "Release control",
  artifact: artifactWorkspaceText("en-US"),
  emptyHeading: (agentLabel) => `${agentLabel}, what should we work on?`,
  composerPlaceholder: (agentLabel) => `Ask ${agentLabel} to work on a task…`,
  requiresArgument: (commandLabel) => `${commandLabel} requires an argument`,
  customAnswerFor: (prompt) => `Custom answer for ${prompt}`,
  configurationOptions: (label) => `${label} options`,
  configurationLabel: (_id, fallback) => fallback,
  configurationOption: (_id, _value, fallback) => fallback,
  sessionStatus: englishSessionStatus,
  activityStatus: englishActivityStatus,
  artifactType: (_kind, fallback) => fallback,
  toolText: (value) => value,
  toolActivityGroupSummary: englishToolActivityGroupSummary,
  fileChangeVerb: englishFileChangeVerb,
};

const zhCN: AgentWorkspaceText = {
  conversation: "对话",
  results: "成果",
  artifacts: "成果列表",
  workspaceViews: "工作区视图",
  terminal: "终端",
  plan: "执行计划",
  agentPlan: "智能体执行计划",
  loadEarlierActivity: "加载更早记录",
  readyForTask: "可以开始任务",
  startSession: "发送消息以启动智能体会话。",
  you: "你",
  agent: "智能体",
  system: "系统",
  agentic: "智能体模式",
  terminalMode: "终端模式",
  approvals: "请求审批",
  messageAgent: "给智能体发送消息",
  readOnly: "此会话为只读状态",
  stopAgent: "停止智能体",
  sendMessage: "发送消息",
  reject: "拒绝",
  approve: "批准",
  submitAnswers: "提交回答",
  customAnswer: "自定义回答",
  details: "详细信息",
  takeControl: "接管终端",
  releaseControl: "释放控制",
  artifact: artifactWorkspaceText("zh-CN"),
  emptyHeading: (agentLabel) => `${agentLabel}，我能为你做什么？`,
  composerPlaceholder: (agentLabel) => `让 ${agentLabel} 帮你完成任务…`,
  requiresArgument: (commandLabel) => `${commandLabel} 需要填写参数`,
  customAnswerFor: (prompt) => `${prompt}的自定义回答`,
  configurationOptions: (label) => `${label}选项`,
  configurationLabel: localizeConfigurationLabel,
  configurationOption: localizeConfigurationOption,
  sessionStatus: chineseSessionStatus,
  activityStatus: chineseActivityStatus,
  artifactType: (kind, fallback) =>
    ({
      "HTML document": "HTML 文档",
      "SVG document": "SVG 文档",
      Image: "图片",
      Video: "视频",
      PDF: "PDF 文档",
      PowerPoint: "PowerPoint 演示文稿",
      "Code file": "代码文件",
      "Text file": "文本文件",
      File: "文件",
    })[kind] ?? fallback,
  toolText: localizeToolText,
  toolActivityGroupSummary: chineseToolActivityGroupSummary,
  fileChangeVerb: localizeFileChangeVerb,
};
