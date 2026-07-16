import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";
import type { WorkerCreateOptions } from "@/lib/api/facade/podConnect";
import { createWorkerTemplateDraft } from "./worker-template-draft";
import { WorkerTemplateRuntimePanel } from "./WorkerTemplateRuntimePanel";

vi.mock("@/components/pod/hooks/useWorkerCreateOptions", () => ({
  useWorkerCreateOptions: () => ({ status: "ready", data: workerOptions() }),
}));

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
        catalog={{ loading: false, error: null, byKind: {} }}
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
        catalog={{ loading: false, error: null, byKind: {} }}
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
    tool_model_requirements: [],
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
