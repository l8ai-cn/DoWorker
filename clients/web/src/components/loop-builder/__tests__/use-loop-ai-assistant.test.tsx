import { act, renderHook, waitFor } from "@testing-library/react";
import { create } from "@bufbuild/protobuf";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { LoopDiagnosticSchema } from "@proto/goalloop/v1/goalloop_pb";
import type { EffectiveResource } from "@/lib/api/facade/aiResource";
import type { LoopWorkbenchSnapshot } from "@/lib/viewModels/loop-program";
import { useLoopAIAssistant } from "../use-loop-ai-assistant";
import type { LoopAIMessages } from "../loop-workbench-messages";

const mocks = vi.hoisted(() => ({
  listResources: vi.fn(),
  requestDraft: vi.fn(),
  decodeDraft: vi.fn(),
  requestRepair: vi.fn(),
  decodeRepair: vi.fn(),
  applyDraft: vi.fn(),
}));

vi.mock("@/lib/api/facade/aiResourceConnect", () => ({
  listOrganizationEffectiveResources: mocks.listResources,
}));
vi.mock("@/lib/api/facade/loopProgramConnect", () => ({
  requestLoopAIDraft: mocks.requestDraft,
  decodeLoopAIDraft: mocks.decodeDraft,
  requestLoopAIRepair: mocks.requestRepair,
  decodeLoopAIRepair: mocks.decodeRepair,
  applyLoopAIDraft: mocks.applyDraft,
}));

const messages = {
  resourceError: "模型加载失败",
  generationError: "生成失败",
  unchanged: "没有变化",
  stale: "提案已过期",
  repair: { error: "修复失败" },
} as LoopAIMessages;

const snapshot: LoopWorkbenchSnapshot = {
  source: "loop current {}",
  canonicalSource: "loop current {}",
  diagnostics: [
    create(LoopDiagnosticSchema, {
      code: "loop.structure.verifier-count",
      message: "missing verifier",
      nodeId: "n-repeat",
      line: 4,
      column: 3,
    }),
  ],
  parseStatus: "valid",
  activeEditor: "blocks",
  revision: 7,
  semanticRevision: 3,
};

const resource: EffectiveResource = {
  selectable: true,
  blockingReason: "",
  connection: {
    id: 2,
    ownerScope: "organization",
    identifier: "team-ai",
    providerKey: "anthropic",
    name: "团队模型",
    baseUrl: "https://api.anthropic.com",
    configuredFields: ["api_key"],
    status: "valid",
    isEnabled: true,
    validationError: "",
    canManage: true,
    resources: [],
  },
  resource: {
    id: 42,
    providerConnectionId: 2,
    identifier: "claude-sonnet",
    modelId: "claude-sonnet",
    displayName: "Claude Sonnet",
    modalities: ["chat"],
    capabilities: ["text-generation"],
    defaultModalities: ["chat"],
    status: "valid",
    isEnabled: true,
    validationError: "",
  },
};

describe("useLoopAIAssistant", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.listResources.mockResolvedValue([resource]);
    mocks.requestDraft.mockResolvedValue(new Uint8Array([7]));
    mocks.decodeDraft.mockReturnValue({ canonicalSource: "loop generated {}" });
    mocks.requestRepair.mockResolvedValue(new Uint8Array([8]));
    mocks.decodeRepair.mockReturnValue({
      proposal: { canonicalSource: "loop repaired {}" },
      proposalBytes: new Uint8Array([9]),
      patch: {
        nodeId: "n-limits",
        fieldPath: "limits.iterations",
        oldValue: 100n,
        newValue: 20n,
      },
    });
    mocks.applyDraft.mockResolvedValue({ applied: true, snapshot });
  });

  it("loads only selectable text-generation resources and generates a preview", async () => {
    const onApplied = vi.fn();
    const { result } = renderHook(() => useLoopAIAssistant({
      orgSlug: "acme",
      locale: "zh-CN",
      snapshot,
      messages,
      onApplied,
    }));

    await waitFor(() => expect(result.current.resources).toEqual([
      { id: "42", label: "团队模型 · Claude Sonnet" },
    ]));
    act(() => {
      result.current.setPrompt("制作专业 PPT");
      result.current.setSelectedResourceId("42");
    });
    await act(() => result.current.submit());

    expect(mocks.requestDraft).toHaveBeenCalledWith({
      orgSlug: "acme",
      prompt: "制作专业 PPT",
      currentSource: "loop current {}",
      modelResourceId: 42,
      locale: "zh-CN",
      revision: 7,
    });
    expect(result.current.proposal?.proposedSource).toBe("loop generated {}");
    expect(mocks.applyDraft).not.toHaveBeenCalled();
    expect(onApplied).not.toHaveBeenCalled();
  });

  it("rejects unchanged and stale proposals without mutating the Loop", async () => {
    mocks.decodeDraft.mockReturnValue({ canonicalSource: "loop current {}" });
    mocks.applyDraft.mockResolvedValue({ applied: false, snapshot });
    const onApplied = vi.fn();
    const { result } = renderHook(() => useLoopAIAssistant({
      orgSlug: "acme",
      locale: "zh-CN",
      snapshot,
      messages,
      onApplied,
    }));
    await waitFor(() => expect(result.current.resourcesLoading).toBe(false));
    act(() => {
      result.current.setPrompt("保持不变");
      result.current.setSelectedResourceId("42");
    });
    await act(() => result.current.submit());
    expect(result.current.requestError).toBe("没有变化");
    expect(result.current.proposal).toBeUndefined();

    mocks.decodeDraft.mockReturnValue({ canonicalSource: "loop generated {}" });
    await act(() => result.current.submit());
    await act(() => result.current.confirm());
    expect(result.current.requestError).toBe("提案已过期");
    expect(onApplied).not.toHaveBeenCalled();
  });

  it("repairs only the selected diagnostic and applies the decoded proposal bytes", async () => {
    const onApplied = vi.fn();
    const { result } = renderHook(() => useLoopAIAssistant({
      orgSlug: "acme",
      locale: "zh-CN",
      snapshot,
      messages,
      onApplied,
    }));
    await waitFor(() => expect(result.current.resourcesLoading).toBe(false));

    act(() => {
      result.current.openRepair({
        diagnosticCode: "loop.value.out-of-range",
        diagnosticLabel: "积木参数超出允许范围",
        nodeId: "n-limits",
        fieldPath: "limits.iterations",
      });
      result.current.setSelectedResourceId("42");
      result.current.setPrompt("保持预算严格");
    });
    await act(() => result.current.submit());

    expect(mocks.requestRepair).toHaveBeenCalledWith({
      orgSlug: "acme",
      source: "loop current {}",
      modelResourceId: 42,
      locale: "zh-CN",
      revision: 7,
      diagnosticCode: "loop.value.out-of-range",
      nodeId: "n-limits",
      fieldPath: "limits.iterations",
      prompt: "保持预算严格",
    });
    expect(mocks.decodeRepair).toHaveBeenCalledWith(new Uint8Array([8]), {
      revision: 7,
      nodeId: "n-limits",
      fieldPath: "limits.iterations",
    });
    expect(result.current.proposal).toMatchObject({
      currentSource: "loop current {}",
      proposedSource: "loop repaired {}",
      repair: {
        fieldPath: "limits.iterations",
        oldValue: 100n,
        newValue: 20n,
      },
    });

    await act(() => result.current.confirm());

    expect(mocks.applyDraft).toHaveBeenCalledWith(new Uint8Array([9]));
    expect(onApplied).toHaveBeenCalledWith(snapshot);
  });
});
