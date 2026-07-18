import { create } from "@bufbuild/protobuf";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  PlanResourceResponseSchema,
} from "@proto/orchestration_resource/v1/orchestration_resource_queries_pb";
import {
  IssueSeverity,
  PlanStatus,
  ResourceOperation,
  ResourceSchema,
} from "@proto/orchestration_resource/v1/orchestration_resource_types_pb";
import { render, screen, waitFor } from "@/test/test-utils";
import { createResourceDraft } from "./resource-draft-factory";

const api = vi.hoisted(() => ({
  validateResource: vi.fn(),
  planResource: vi.fn(),
  applyBindingResourcePlan: vi.fn(),
  applyExpertPlan: vi.fn(),
  applyPromptPlan: vi.fn(),
  applyWorkerTemplatePlan: vi.fn(),
  applyWorkflowPlan: vi.fn(),
  createGoalLoopFromPlan: vi.fn(),
  createWorkerFromPlan: vi.fn(),
  listResources: vi.fn(),
}));

vi.mock("@/lib/api/facade/orchestrationResource", () => ({ ...api }));

import { ResourceEditorShell } from "./ResourceEditorShell";

describe("ResourceEditorShell blocking states", () => {
  beforeEach(() => {
    Object.values(api).forEach((method) => method.mockReset());
    api.listResources.mockResolvedValue({ items: [] });
  });

  it("shows a typed permission error without the wire JSON envelope", async () => {
    const user = userEvent.setup();
    const wireError = JSON.stringify({
      kind: "http",
      status: 403,
      code: "permission_denied",
      message: "You cannot plan this resource. @ https://internal/api/plan",
    });
    api.planResource.mockRejectedValue(new Error(wireError));
    render(<ResourceEditorShell orgSlug="acme" kind="Prompt" />);

    await user.click(screen.getByRole("button", { name: "Generate plan" }));

    expect(await screen.findByText("You cannot plan this resource."))
      .toBeInTheDocument();
    expect(document.body).not.toHaveTextContent(wireError);
    expect(document.body).not.toHaveTextContent("https://internal/api/plan");
  });

  it("removes control-plane URLs from network errors", async () => {
    const user = userEvent.setup();
    api.planResource.mockRejectedValue(new Error(
      "error sending request for url (https://internal/api/plan) " +
      "@ https://internal/api/plan",
    ));
    render(<ResourceEditorShell orgSlug="acme" kind="Prompt" />);

    await user.click(screen.getByRole("button", { name: "Generate plan" }));

    expect(await screen.findByText("error sending request"))
      .toBeInTheDocument();
    expect(document.body).not.toHaveTextContent("https://internal/api/plan");
  });

  it("keeps blocking plan issues visible and disables Apply", async () => {
    const user = userEvent.setup();
    api.planResource.mockResolvedValue(create(PlanResourceResponseSchema, {
      operation: ResourceOperation.CREATE,
      issues: [{
        severity: IssueSeverity.BLOCKING,
        code: "REFERENCE_FORBIDDEN",
        path: "/spec/modelRef",
        message: "The model binding is not readable.",
      }],
    }));
    render(<ResourceEditorShell orgSlug="acme" kind="Prompt" />);

    await user.click(screen.getByRole("button", { name: "Generate plan" }));

    expect(await screen.findByText(
      "/spec/modelRef: The model binding is not readable.",
    )).toBeInTheDocument();
    expect(screen.queryByText("Unexpected end of JSON input"))
      .not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Apply resource" }))
      .toBeDisabled();
  });

  it("rejects a plan response without a canonical document", async () => {
    const user = userEvent.setup();
    api.planResource.mockResolvedValue(create(PlanResourceResponseSchema, {
      operation: ResourceOperation.CREATE,
      plan: {
        planId: "missing-canonical",
        expiresAt: "2099-07-16T00:00:00Z",
        status: PlanStatus.PENDING,
      },
    }));
    render(<ResourceEditorShell orgSlug="acme" kind="Prompt" />);

    await user.click(screen.getByRole("button", { name: "Generate plan" }));

    expect(await screen.findByText(
      "Resource planning response did not include a canonical document.",
    )).toBeInTheDocument();
    expect(screen.queryByText("missing-canonical")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Apply resource" }))
      .toBeDisabled();
  });

  it("replaces an older Validate error with the new Plan result", async () => {
    const user = userEvent.setup();
    api.validateResource.mockRejectedValue(new Error(
      "The validation request failed.",
    ));
    api.planResource.mockResolvedValue(readyPlan("validated-by-plan"));
    render(<ResourceEditorShell orgSlug="acme" kind="Prompt" />);

    await user.click(screen.getByRole("button", { name: "Validate" }));
    expect(await screen.findByText("The validation request failed."))
      .toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Generate plan" }));

    expect(await screen.findByText("Plan ready")).toBeInTheDocument();
    expect(screen.queryByText("The validation request failed."))
      .not.toBeInTheDocument();
  });

  it("expires Apply locally and accepts a newly generated plan", async () => {
    const user = userEvent.setup();
    api.planResource
      .mockResolvedValueOnce(readyPlan(
        "short-plan",
        new Date(Date.now() + 2_000).toISOString(),
      ))
      .mockResolvedValueOnce(readyPlan("replacement-plan"));
    render(<ResourceEditorShell orgSlug="acme" kind="Prompt" />);

    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    const apply = screen.getByRole("button", { name: "Apply resource" });
    expect(await screen.findByText("Plan ready")).toBeInTheDocument();
    expect(apply).toBeEnabled();

    expect(await screen.findByText(
      "This plan has expired. Generate a new plan before applying.",
      {},
      { timeout: 3_000 },
    )).toBeInTheDocument();
    expect(apply).toBeDisabled();
    expect(screen.getByText("Plan expired")).toBeInTheDocument();
    expect(screen.getByText("short-plan")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await waitFor(() => expect(apply).toBeEnabled());
  });

  it("retires a conflicted plan until the user generates a new one", async () => {
    const user = userEvent.setup();
    const wireError = JSON.stringify({
      kind: "http",
      status: 409,
      code: "plan_conflict",
      message: "The resource changed after this plan was generated.",
    });
    api.planResource.mockResolvedValue(readyPlan("conflicting-plan"));
    api.applyPromptPlan.mockRejectedValue(new Error(wireError));
    render(<ResourceEditorShell orgSlug="acme" kind="Prompt" />);

    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    const apply = screen.getByRole("button", { name: "Apply resource" });
    await user.click(apply);

    expect(await screen.findByText(
      "The resource changed after this plan was generated.",
    )).toBeInTheDocument();
    expect(document.body).not.toHaveTextContent(wireError);
    expect(screen.getByText("Plan ready")).toBeInTheDocument();
    expect(apply).toBeDisabled();

    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await waitFor(() => expect(apply).toBeEnabled());
  });

  it("retires a successfully applied plan until a new review", async () => {
    const user = userEvent.setup();
    api.planResource.mockResolvedValue(readyPlan("successful-plan"));
    api.applyPromptPlan.mockResolvedValue(create(ResourceSchema, {
      revision: 3n,
    }));
    render(<ResourceEditorShell orgSlug="acme" kind="Prompt" />);

    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    const apply = screen.getByRole("button", { name: "Apply resource" });
    await user.click(apply);

    expect(await screen.findByText("Revision 3")).toBeInTheDocument();
    expect(apply).toBeDisabled();

    await user.click(screen.getByRole("button", { name: "Generate plan" }));
    await waitFor(() => expect(apply).toBeEnabled());
  });

  it.each([
    [PlanStatus.UNSPECIFIED, "Plan unavailable"],
    [PlanStatus.APPLIED, "Plan applied"],
    [PlanStatus.CANCELLED, "Plan cancelled"],
    [PlanStatus.EXPIRED, "Plan expired"],
  ])("blocks Apply for protocol status %s", async (status, label) => {
    const user = userEvent.setup();
    api.planResource.mockResolvedValue(readyPlan(
      "terminal-plan",
      "2099-07-16T00:00:00Z",
      status,
    ));
    render(<ResourceEditorShell orgSlug="acme" kind="Prompt" />);

    await user.click(screen.getByRole("button", { name: "Generate plan" }));

    expect((await screen.findAllByText(label)).length).toBeGreaterThan(0);
    expect(screen.getByRole("button", { name: "Apply resource" }))
      .toBeDisabled();
  });
});

function readyPlan(
  planId: string,
  expiresAt = "2099-07-16T00:00:00Z",
  status = PlanStatus.PENDING,
) {
  return create(PlanResourceResponseSchema, {
    operation: ResourceOperation.CREATE,
    canonicalJson: promptCanonicalJson(),
    plan: {
      planId,
      operation: ResourceOperation.CREATE,
      expiresAt,
      status,
    },
  });
}

function promptCanonicalJson(): Uint8Array {
  const draft = createResourceDraft("Prompt", "acme");
  draft.metadata.name = "planned-prompt";
  draft.spec.content = "Review the release";
  return new TextEncoder().encode(JSON.stringify(draft));
}
