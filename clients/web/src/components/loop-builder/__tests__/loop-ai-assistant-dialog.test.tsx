import { create } from "@bufbuild/protobuf";
import { fireEvent, render, screen } from "@/test/test-utils";
import { describe, expect, it, vi } from "vitest";
import { LoopProgramSchema } from "@proto/goalloop/v1/goalloop_pb";
import zhMessages from "@/messages/zh/app.json";
import { LoopAIAssistantDialog } from "../loop-ai-assistant-dialog";
import {
  createLoopWorkbenchMessages,
  type LoopMessageTranslator,
} from "../loop-workbench-messages";

function translator(messages: Record<string, unknown>): LoopMessageTranslator {
  return (key, values) => {
    const value = key.split(".").reduce<unknown>(
      (current, segment) => (current as Record<string, unknown>)[segment],
      messages,
    );
    return Object.entries(values ?? {}).reduce(
      (text, [name, replacement]) => text.replace(`{${name}}`, String(replacement)),
      String(value),
    );
  };
}

const messages = createLoopWorkbenchMessages(
  translator(zhMessages.loopWorkbench),
).ai;

const resources = [{ id: "42", label: "团队模型 · Claude Sonnet" }];
const program = create(LoopProgramSchema, {
  schemaVersion: 1,
  loop: { nodeId: "n-loop", localId: "ppt-loop" },
  limits: {
    iterations: 5n,
    tokens: 80000n,
    timeoutMinutes: 60n,
    noProgress: 3n,
    sameError: 2n,
  },
  repeat: {
    identity: { nodeId: "n-repeat", localId: "build-cycle" },
    max: 5n,
    until: { localId: "verify-ppt", field: "passed" },
    agent: {
      identity: { nodeId: "n-agent", localId: "build-ppt" },
      prompt: "制作专业 PPT",
    },
    verifier: {
      identity: { nodeId: "n-verifier", localId: "verify-ppt" },
      command: "test -f output.pptx",
      accept: "PPTX 文件存在且可打开",
    },
  },
  failurePolicy: "pause",
});

function renderDialog(overrides: Partial<React.ComponentProps<typeof LoopAIAssistantDialog>> = {}) {
  const props: React.ComponentProps<typeof LoopAIAssistantDialog> = {
    open: true,
    mode: "generate",
    parseStatus: "valid",
    program,
    prompt: "",
    selectedResourceId: "",
    resources,
    resourcesLoading: false,
    busy: false,
    messages,
    onOpenChange: vi.fn(),
    onModeChange: vi.fn(),
    onPromptChange: vi.fn(),
    onResourceChange: vi.fn(),
    onRetryResources: vi.fn(),
    onSubmit: vi.fn(),
    onBack: vi.fn(),
    onConfirm: vi.fn(),
    ...overrides,
  };
  render(<LoopAIAssistantDialog {...props} />);
  return props;
}

describe("LoopAIAssistantDialog", () => {
  it("shows resource loading without a false empty state", () => {
    renderDialog({ resources: [], resourcesLoading: true });

    expect(screen.getByText("正在加载 AI 模型")).toBeInTheDocument();
    expect(screen.queryByText("没有可用的文本生成模型")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成草稿" })).toBeDisabled();
  });

  it("shows retryable resource and generation failures", () => {
    const props = renderDialog({
      resources: [],
      resourceError: "AI 模型加载失败",
      requestError: "生成失败",
    });

    expect(screen.getByText("AI 模型加载失败")).toBeInTheDocument();
    expect(screen.getByText("生成失败")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "重新加载" }));
    expect(props.onRetryResources).toHaveBeenCalledOnce();
  });

  it("requires an explicit model and prompt before generation", () => {
    const props = renderDialog();
    const generate = screen.getByRole("button", { name: "生成草稿" });
    expect(generate).toBeDisabled();

    fireEvent.click(screen.getByRole("button", { name: "选择 AI 模型" }));
    fireEvent.click(screen.getByRole("option", { name: /Claude Sonnet/ }));
    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "制作专业 PPT" },
    });

    expect(props.onResourceChange).toHaveBeenCalledWith("42");
    expect(props.onPromptChange).toHaveBeenCalledWith("制作专业 PPT");
  });

  it("projects a valid LoopProgram without model, prompt, runtime, or Worker controls", () => {
    renderDialog({
      mode: "explain",
      selectedResourceId: "",
    });

    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    expect(screen.queryByText("AI 模型")).not.toBeInTheDocument();
    expect(screen.queryByText(/运行环境/)).not.toBeInTheDocument();
    expect(screen.queryByText(/Worker/i)).not.toBeInTheDocument();
    expect(screen.getByText("执行预算")).toBeInTheDocument();
    expect(screen.getByText("5 轮")).toBeInTheDocument();
    expect(screen.getByText("test -f output.pptx")).toBeInTheDocument();
    expect(screen.getByText("PPTX 文件存在且可打开")).toBeInTheDocument();
  });

  it("previews both sources and confirms only on explicit action", () => {
    const props = renderDialog({
      proposal: {
        currentSource: "loop current {}",
        proposedSource: "loop generated {}",
      },
    });

    expect(screen.getByText("loop current {}")).toBeInTheDocument();
    expect(screen.getByText("loop generated {}")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "返回" }));
    expect(props.onBack).toHaveBeenCalledOnce();
    expect(props.onConfirm).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "确认应用" }));
    expect(props.onConfirm).toHaveBeenCalledOnce();
  });

  it("does not project stale or invalid source semantics", () => {
    renderDialog({
      mode: "explain",
      parseStatus: "syntax-error",
    });

    expect(screen.getByText("当前循环尚未通过校验，无法生成结构说明。")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "确认应用" })).not.toBeInTheDocument();
  });

  it("requires a model for targeted repair but keeps the extra instruction optional", () => {
    const props = renderDialog({
      repairTarget: {
        diagnosticCode: "loop.value.out-of-range",
        diagnosticLabel: "积木参数超出允许范围",
        nodeId: "n-limits",
        fieldPath: "limits.iterations",
      },
    });

    expect(screen.getByText("修复诊断")).toBeInTheDocument();
    expect(screen.getByText("积木参数超出允许范围")).toBeInTheDocument();
    expect(screen.getByText("limits.iterations")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "生成修复" })).toBeDisabled();

    fireEvent.click(screen.getByRole("button", { name: "选择 AI 模型" }));
    fireEvent.click(screen.getByRole("option", { name: /Claude Sonnet/ }));

    expect(props.onResourceChange).toHaveBeenCalledWith("42");
  });

  it("previews the exact repair patch before explicit confirmation", () => {
    const props = renderDialog({
      repairTarget: {
        diagnosticCode: "loop.value.out-of-range",
        diagnosticLabel: "积木参数超出允许范围",
        nodeId: "n-limits",
        fieldPath: "limits.iterations",
      },
      proposal: {
        response: new Uint8Array([9]),
        currentSource: "limits(iterations: 100)",
        proposedSource: "limits(iterations: 20)",
        repair: {
          nodeId: "n-limits",
          fieldPath: "limits.iterations",
          oldValue: 100n,
          newValue: 20n,
        },
      },
    });

    expect(screen.getByText("100 调整为 20")).toBeInTheDocument();
    expect(screen.getByText("limits(iterations: 100)")).toBeInTheDocument();
    expect(screen.getByText("limits(iterations: 20)")).toBeInTheDocument();
    expect(props.onConfirm).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "确认应用" }));
    expect(props.onConfirm).toHaveBeenCalledOnce();
  });
});
