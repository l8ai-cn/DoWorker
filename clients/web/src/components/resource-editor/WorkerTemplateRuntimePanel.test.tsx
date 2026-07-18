import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";
import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";
import { createWorkerTemplateDraft } from "./worker-template-draft";
import { WorkerTemplateRuntimePanel } from "./WorkerTemplateRuntimePanel";

describe("WorkerTemplateRuntimePanel", () => {
  it("uses live Worker options instead of a raw runtime image ID", async () => {
    const user = userEvent.setup();
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.workerType = "codex-cli";
    draft.spec.runtime.runtimeImageId = 11;
    const onChange = vi.fn();

    render(
      <WorkerTemplateRuntimePanel
        draft={draft}
        catalog={{
          loading: false,
          error: null,
          errorsByKind: {},
          byKind: {},
        }}
        workerOptions={{ status: "ready", data: workerOptions() }}
        onChange={onChange}
      />,
    );

    expect(screen.queryByRole("spinbutton", {
      name: "Runtime image ID",
    })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Worker type" }));
    expect(screen.getByRole("option", {
      name: /Gemini CLI.*No online Runner/i,
    })).toHaveAttribute("aria-disabled", "true");
  });

  it("selects a compatible image and PTY mode when the Worker type changes", async () => {
    const user = userEvent.setup();
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.workerType = "codex-cli";
    draft.spec.optionsRevision = "catalog-old";
    draft.spec.runtime.runtimeImageId = 11;
    const onChange = vi.fn();

    render(
      <WorkerTemplateRuntimePanel
        draft={draft}
        catalog={{
          loading: false,
          error: null,
          errorsByKind: {},
          byKind: {},
        }}
        workerOptions={{ status: "ready", data: workerOptions() }}
        onChange={onChange}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Worker type" }));
    await user.click(screen.getByRole("option", { name: "MiniMax CLI" }));

    expect(onChange).toHaveBeenLastCalledWith(expect.objectContaining({
      spec: expect.objectContaining({
        optionsRevision: "catalog-current",
        workerType: "minimax-cli",
        runtime: expect.objectContaining({ runtimeImageId: 42 }),
        typeConfig: expect.objectContaining({
          interactionMode: "pty",
          schemaVersion: 3,
        }),
      }),
    }));
  });

  it("associates runtime controls with their labels", () => {
    const draft = createWorkerTemplateDraft("acme");
    draft.spec.workerType = "codex-cli";
    draft.spec.runtime.customResources = {
      cpuRequestMilliCPU: 500,
      cpuLimitMilliCPU: 1000,
      memoryRequestBytes: 536870912,
      memoryLimitBytes: 1073741824,
      storageRequestBytes: 1073741824,
      storageLimitBytes: 10737418240,
    };

    render(
      <WorkerTemplateRuntimePanel
        draft={draft}
        catalog={{
          loading: false,
          error: null,
          errorsByKind: {},
          byKind: {},
        }}
        workerOptions={{ status: "ready", data: workerOptions() }}
        onChange={vi.fn()}
      />,
    );

    expect(screen.getByLabelText("Placement policy")).toHaveAttribute(
      "id",
      "placement-policy",
    );
    expect(screen.getByLabelText("Resource allocation")).toHaveAttribute(
      "id",
      "resource-mode",
    );
    expect(screen.getByLabelText("CPU request (millicpu)")).toHaveAttribute(
      "id",
      "cpu-request",
    );
    expect(screen.getByLabelText("GPU limit")).toHaveAttribute(
      "id",
      "gpu-limit",
    );
  });
});

function workerOptions(): WorkerCreateOptions {
  return {
    revision: "catalog-current",
    worker_types: [
      workerType("codex-cli", "Codex CLI", 2, ["pty", "acp"], true),
      workerType(
        "gemini-cli",
        "Gemini CLI",
        1,
        ["acp"],
        false,
        "No online Runner currently supports this worker type",
      ),
      workerType("minimax-cli", "MiniMax CLI", 3, ["pty"], true),
    ],
    runtime_images: [
      runtimeImage(11, "Codex stable", ["codex-cli"]),
      runtimeImage(12, "Gemini stable", ["gemini-cli"]),
      runtimeImage(42, "MiniMax stable", ["minimax-cli"]),
    ],
    compute_targets: [],
    deployment_modes: [],
    resource_profiles: [],
  };
}

function workerType(
  slug: string,
  name: string,
  schemaVersion: number,
  supportedInteractionModes: string[],
  selectable: boolean,
  blockingReason = "",
) {
  return {
    slug,
    name,
    description: "",
    schema_version: schemaVersion,
    config_schema: {},
    supported_interaction_modes: supportedInteractionModes,
    requires_model_resource: false,
    model_protocol_adapters: [],
    tool_model_requirements: [],
    credential_requirements: [],
    config_document_requirements: [],
    selectable,
    blocking_reason: blockingReason,
  };
}

function runtimeImage(id: number, name: string, workerTypeSlugs: string[]) {
  return {
    id,
    slug: name.toLowerCase().replaceAll(" ", "-"),
    name,
    reference: "",
    digest: "",
    worker_type_slugs: workerTypeSlugs,
    selectable: true,
    blocking_reason: "",
  };
}
