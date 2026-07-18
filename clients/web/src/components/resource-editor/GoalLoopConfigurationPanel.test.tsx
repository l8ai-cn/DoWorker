import { useState } from "react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@/test/test-utils";
import { createGoalLoopDraft } from "./resource-definition-drafts";
import { createResourceDraft } from "./resource-draft-factory";
import type {
  GoalLoopDraft,
  ResourceDraft,
} from "./resource-editor-types";

vi.mock("./use-resource-reference-options", () => ({
  useResourceReferenceOptions: () => ({
    loading: false,
    error: null,
    errorsByKind: {},
    byKind: {
      WorkerTemplate: [{
        name: "code-reviewer",
        displayName: "Code reviewer",
        revision: 7,
      }],
    },
  }),
}));

import { ResourceConfigurationPanel } from "./ResourceConfigurationPanel";
import { ResourceEditorShell } from "./ResourceEditorShell";

describe("GoalLoopConfigurationPanel", () => {
  it("creates the GoalLoop form defaults", () => {
    expect(createResourceDraft("GoalLoop", "acme")).toEqual({
      apiVersion: "agentsmesh.io/v1alpha1",
      kind: "GoalLoop",
      metadata: {
        name: "",
        namespace: "acme",
        displayName: "",
        labels: {},
      },
      spec: {
        workerTemplateRef: { kind: "WorkerTemplate", name: "" },
        description: "",
        objective: "",
        acceptanceCriteria: [""],
        verificationCommand: "",
        maxIterations: 10,
        tokenBudget: undefined,
        timeoutMinutes: 60,
        noProgressLimit: 3,
        sameErrorLimit: 2,
        escalationPolicy: "pause",
      },
    });
  });

  it("updates the objective, verification, policy, and numeric draft fields", async () => {
    const user = userEvent.setup();
    render(<GoalLoopPanelHarness />);

    await user.type(screen.getByLabelText(/^Objective/), "Ship the release");
    await user.type(
      screen.getByLabelText(/^Verification command/),
      "pnpm test",
    );
    await user.selectOptions(screen.getByLabelText("Escalation policy"), "fail");
    fireEvent.change(screen.getByLabelText(/^Maximum iterations/), {
      target: { value: "25" },
    });
    fireEvent.change(screen.getByLabelText("Token budget"), {
      target: { value: "50000" },
    });

    expect(currentDraft().spec).toMatchObject({
      objective: "Ship the release",
      verificationCommand: "pnpm test",
      escalationPolicy: "fail",
      maxIterations: 25,
      tokenBudget: 50000,
    });
  });

  it("adds, edits, and removes acceptance criteria", async () => {
    const user = userEvent.setup();
    render(<GoalLoopPanelHarness />);

    await user.type(
      screen.getByLabelText("Acceptance criterion 1"),
      "Focused tests pass",
    );
    await user.click(screen.getByRole("button", {
      name: "Add acceptance criteria",
    }));
    await user.type(
      screen.getByLabelText("Acceptance criterion 2"),
      "Typecheck passes",
    );
    await user.click(screen.getByRole("button", {
      name: "Remove acceptance criterion 1",
    }));

    expect(currentDraft().spec.acceptanceCriteria).toEqual([
      "Typecheck passes",
    ]);
    expect(screen.getAllByLabelText(/Acceptance criterion \d/)).toHaveLength(1);
  });

  it("updates the WorkerTemplate reference and exposes numeric bounds", async () => {
    const user = userEvent.setup();
    render(<GoalLoopPanelHarness />);

    await user.selectOptions(
      screen.getByRole("combobox", { name: /^Worker template/ }),
      "code-reviewer",
    );

    expect(currentDraft().spec.workerTemplateRef).toEqual({
      kind: "WorkerTemplate",
      name: "code-reviewer",
    });
    expectBounds(/^Maximum iterations/, "1", "100");
    expectBounds(/^Run timeout \(minutes\)/, "1", "1440");
    expectBounds(/^No-progress limit/, "1", "20");
    expectBounds(/^Same-error limit/, "1", "20");
    expect(screen.getByLabelText("Token budget")).toHaveAttribute("min", "1");
  });

  it("keeps invalid numeric text visible and reports why it cannot be used", () => {
    render(<GoalLoopPanelHarness />);

    fireEvent.change(screen.getByLabelText(/^Maximum iterations/), {
      target: { value: "" },
    });

    expect(currentDraft().spec.maxIterations).toBe("");
    expect(screen.getByText("Enter a whole number.")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/^Maximum iterations/), {
      target: { value: "1.5" },
    });

    expect(currentDraft().spec.maxIterations).toBe("1.5");
    expect(screen.getByText("Use a whole number without decimals."))
      .toBeInTheDocument();
  });

  it("does not round an unsafe token budget", () => {
    render(<GoalLoopPanelHarness />);

    fireEvent.change(screen.getByLabelText("Token budget"), {
      target: { value: "9007199254740993" },
    });

    expect(currentDraft().spec.tokenBudget).toBe("9007199254740993");
    expect(screen.getByText("Use an integer within the safe numeric range."))
      .toBeInTheDocument();
  });

  it("blocks Validate and Plan while GoalLoop numeric fields are invalid", () => {
    render(<ResourceEditorShell orgSlug="acme" kind="GoalLoop" />);

    fireEvent.change(screen.getByLabelText(/^Maximum iterations/), {
      target: { value: "" },
    });

    expect(screen.getByRole("button", { name: "Validate" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Generate plan" })).toBeDisabled();
  });

  it("shows the GoalLoop resource heading", () => {
    render(<ResourceEditorShell orgSlug="acme" kind="GoalLoop" />);

    expect(screen.getByRole("heading", { name: "Goal loop" }))
      .toBeInTheDocument();
    expect(screen.getByText(
      "Define bounded autonomous work with explicit completion evidence.",
    )).toBeInTheDocument();
  });
});

function GoalLoopPanelHarness() {
  const [draft, setDraft] = useState(createGoalLoopDraft("acme"));
  return (
    <>
      <ResourceConfigurationPanel
        orgSlug="acme"
        draft={draft}
        onChange={(next: ResourceDraft) => setDraft(next as GoalLoopDraft)}
        onPlanBlockChange={vi.fn()}
      />
      <output data-testid="goal-loop-draft">{JSON.stringify(draft)}</output>
    </>
  );
}

function currentDraft(): GoalLoopDraft {
  return JSON.parse(screen.getByTestId("goal-loop-draft").textContent ?? "");
}

function expectBounds(label: RegExp, min: string, max: string) {
  const input = screen.getByLabelText(label);
  expect(input).toHaveAttribute("min", min);
  expect(input).toHaveAttribute("max", max);
}
