import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { NlWorkerCreate } from "@/components/workers/NlWorkerCreate";
import {
  createInitialWorkerDraftState,
  workerCreateDraftReducer,
} from "../../hooks/workerCreateDraft";
import { WorkerPreflightStep } from "../WorkerPreflightStep";
import { completeDraft, modelResource } from "./test-utils";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

const t = (key: string) => key;

describe("Worker preflight flow", () => {
  it("renders blocking errors separately from warnings and exposes one create action", () => {
    render(
      <WorkerPreflightStep
        creating={false}
        onCreate={vi.fn()}
        onRetry={vi.fn()}
        preflight={{
          status: "ready",
          data: {
            issues: [
              {
                code: "invalid-draft",
                field: "worker_spec.branch",
                message: "Branch is required",
                severity: "blocking",
              },
              {
                code: "large-profile",
                field: "worker_spec.resource_profile_id",
                message: "Large profile costs more",
                severity: "warning",
              },
            ],
            options_revision: "runtime-catalog-1",
          },
        }}
        t={t}
      />,
    );

    expect(screen.getByTestId("preflight-blocking")).toHaveTextContent(
      "Branch is required",
    );
    expect(screen.getByTestId("preflight-warnings")).toHaveTextContent(
      "Large profile costs more",
    );
    const createButtons = screen.getAllByRole("button", {
      name: "workerCreate.actions.create",
    });
    expect(createButtons).toHaveLength(1);
    expect(createButtons[0]).toBeDisabled();
  });

  it("does not report ready or enable create without a resolved spec", () => {
    render(
      <WorkerPreflightStep
        creating={false}
        onCreate={vi.fn()}
        onRetry={vi.fn()}
        preflight={{
          status: "ready",
          data: {
            issues: [],
            options_revision: "runtime-catalog-1",
          },
        }}
        t={t}
      />,
    );

    expect(
      screen.queryByText("workerCreate.preflight.ready"),
    ).not.toBeInTheDocument();
    expect(
      screen.getByText("workerCreate.preflight.missingResolvedSpec"),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "workerCreate.actions.create" }),
    ).toBeDisabled();
  });

  it("applies Fill with AI to the same reducer instance without creating", () => {
    const original = createInitialWorkerDraftState(completeDraft());
    const filledDraft = { ...original.draft, alias: "review-worker" };
    const requestId = "fill-request";
    const loading = workerCreateDraftReducer(original, {
      type: "fill_loading",
      requestId,
    });
    const next = workerCreateDraftReducer(loading, {
      type: "fill_succeeded",
      requestId,
      result: { draft: filledDraft, issues: [] },
    });

    expect(next.instanceId).toBe(original.instanceId);
    expect(next.draft.alias).toBe("review-worker");

    const onFill = vi.fn();
    const { rerender } = render(
      <NlWorkerCreate
        filling={false}
        generationModelResourceId={42}
        generationModels={{ status: "ready", data: [modelResource()] }}
        onFill={onFill}
        onGenerationModelChange={vi.fn()}
        onPromptChange={vi.fn()}
        prompt="Review authentication"
      />,
    );
    const panel = screen.getByTestId("worker-fill-panel");
    fireEvent.click(screen.getByRole("button", {
      name: "workers.create.nl.submit",
    }));
    expect(onFill).toHaveBeenCalledWith("Review authentication");

    rerender(
      <NlWorkerCreate
        filling={false}
        generationModelResourceId={42}
        generationModels={{ status: "ready", data: [modelResource()] }}
        onFill={onFill}
        onGenerationModelChange={vi.fn()}
        onPromptChange={vi.fn()}
        prompt="Review authentication"
      />,
    );
    expect(screen.getByTestId("worker-fill-panel")).toBe(panel);
  });
});
