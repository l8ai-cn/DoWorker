import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { WorkerRuntimeStep } from "../WorkerRuntimeStep";
import { completeDraft, createOptions } from "./test-utils";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("WorkerRuntimeStep", () => {
  it("explains why creation is blocked when no compatible model resource exists", () => {
    render(
      <WorkerRuntimeStep
        draft={completeDraft()}
        modelResources={{ status: "ready", data: [] }}
        options={{ status: "ready", data: createOptions() }}
        onPatch={vi.fn()}
        onWorkerTypeChange={vi.fn()}
        t={(key) => key}
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent(
      "ide.createPod.noModelResourcesAvailableHint",
    );
  });
});
