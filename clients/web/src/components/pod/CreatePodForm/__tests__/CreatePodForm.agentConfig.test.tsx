import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { WorkerTypeConfigStep } from "../WorkerTypeConfigStep";
import { WorkerWorkspaceStep } from "../WorkerWorkspaceStep";
import {
  completeDraft,
  controllerFixture,
  createOptions,
  mockPatchDraft,
  mockSetLifecycle,
} from "./test-utils";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));
vi.mock("../WorkerRepositoryField", () => ({
  WorkerRepositoryField: ({ onChange }: { onChange: (id: number | null) => void }) => (
    <button type="button" onClick={() => onChange(52)}>select-repository</button>
  ),
}));
vi.mock("../WorkerWorkspaceCapabilities", () => ({
  WorkerWorkspaceCapabilities: () => <div>workspace-capabilities</div>,
}));

const t = (key: string) => key;

describe("Worker type and workspace configuration", () => {
  it("localizes declared Agent variable labels", () => {
    const options = createOptions();
    options.worker_types[0].config_schema = {
      version: 1,
      fields: {
        approval_mode: { kind: "select", options: ["never", "ask"] },
      },
    };
    const localized = (key: string) => (
      key === "workerCreate.typeConfig.fields.approvalMode" ? "审批方式" : key
    );

    render(
      <WorkerTypeConfigStep
        draft={completeDraft()}
        options={{ status: "ready", data: options }}
        credentialBundles={{ status: "ready", data: [] }}
        onPatch={mockPatchDraft}
        t={localized}
      />,
    );

    expect(screen.getByLabelText("审批方式")).toBeInTheDocument();
  });

  it("localizes select options and preserves the empty default option", () => {
    const options = createOptions();
    options.worker_types[0].config_schema = {
      version: 1,
      fields: {
        approval_mode: { kind: "select", options: ["", "never"] },
      },
    };
    const localized = (key: string) => ({
      "workerCreate.typeConfig.fields.approvalMode": "审批方式",
      "workerCreate.typeConfig.useDefault": "默认",
      "workerCreate.typeConfig.options.approvalNever": "不审批（全自动）",
      "workerCreate.typeConfig.options.empty": "默认",
    }[key] ?? key);

    render(
      <WorkerTypeConfigStep
        draft={completeDraft()}
        options={{ status: "ready", data: options }}
        credentialBundles={{ status: "ready", data: [] }}
        onPatch={mockPatchDraft}
        t={localized}
      />,
    );

    fireEvent.click(screen.getByLabelText("审批方式"));
    expect(screen.getByRole("option", { name: "默认" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "不审批（全自动）" })).toBeInTheDocument();
  });

  it("writes typed values and secret references into WorkerSpecDraft", () => {
    const options = createOptions();
    options.worker_types[0].config_schema = {
      version: 1,
      fields: {
        approval_mode: { kind: "select", options: ["never", "ask"] },
        retries: { kind: "number", options: [] },
        signing_key: { kind: "secret", options: [] },
      },
    };
    render(
      <WorkerTypeConfigStep
        draft={completeDraft()}
        options={{ status: "ready", data: options }}
        credentialBundles={{
          status: "ready",
          data: [{
            id: 8,
            name: "signing",
            kind: "credential",
            kind_primary: false,
            configured_fields: ["signing_key"],
          }],
        }}
        onPatch={mockPatchDraft}
        t={t}
      />,
    );

    fireEvent.change(screen.getByLabelText("Retries"), { target: { value: "3" } });
    expect(mockPatchDraft).toHaveBeenCalledWith({
      type_config_values: { retries: 3 },
    });

    fireEvent.click(screen.getByLabelText("Signing Key"));
    fireEvent.click(screen.getByText("signing"));
    expect(mockPatchDraft).toHaveBeenCalledWith({
      secret_refs: [{ field: "signing_key", kind: "env-bundle", id: 8 }],
    });
  });

  it("renders workspace fields and keeps lifecycle values on the draft contract", () => {
    const controller = controllerFixture({
      state: {
        step: 3,
        draft: { ...completeDraft(), repository_id: 51 },
      },
    });
    render(
      <WorkerWorkspaceStep
        controller={controller}
        promptPlaceholder="Describe the task"
        t={t}
      />,
    );

    expect(screen.getByText("workspace-capabilities")).toBeInTheDocument();
    fireEvent.click(screen.getByText("select-repository"));
    expect(mockPatchDraft).toHaveBeenCalledWith({
      repository_id: 52,
      branch: "",
      skill_ids: [],
    });

    fireEvent.change(screen.getByLabelText("ide.createPod.prompt"), {
      target: { value: "Review auth" },
    });
    expect(mockPatchDraft).toHaveBeenCalledWith({ initial_task: "Review auth" });

    fireEvent.change(screen.getByLabelText("ide.createPod.alias"), {
      target: { value: "review-worker" },
    });
    expect(mockPatchDraft).toHaveBeenCalledWith({ alias: "review-worker" });

    fireEvent.click(screen.getByLabelText("ide.createPod.lifecyclePolicyLabel"));
    fireEvent.click(screen.getByText("ide.createPod.lifecyclePolicy.idle"));
    expect(mockSetLifecycle).toHaveBeenCalledWith("idle", 30);
  });
});
