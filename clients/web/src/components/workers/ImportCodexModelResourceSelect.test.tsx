import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ImportCodexModelResourceSelect } from "./ImportCodexModelResourceSelect";

const getRequirement = vi.fn();
const useWorkerModelResources = vi.fn();

vi.mock("@/lib/api/sessionImportWorkerPlan", () => ({
  getSessionImportWorkerRequirement: (...args: unknown[]) =>
    getRequirement(...args),
}));

vi.mock("@/components/pod/hooks/useWorkerModelResources", () => ({
  useWorkerModelResources: (...args: unknown[]) =>
    useWorkerModelResources(...args),
}));

vi.mock("@/components/pod/CreatePodForm/WorkerModelResourceSelect", () => ({
  WorkerModelResourceSelect: ({ validationError }: { validationError?: string }) => (
    <div>{validationError ?? "model resource select"}</div>
  ),
}));

describe("ImportCodexModelResourceSelect", () => {
  it("loads authoritative model requirements without clearing parent selection", async () => {
    const onSelect = vi.fn();
    getRequirement.mockResolvedValue({
      modelProtocolAdapters: ["openai-compatible"],
      requiresModelResource: true,
    });
    useWorkerModelResources.mockReturnValue({
      loadingModelResources: false,
      modelResourceError: null,
      modelResources: [],
      selectedModelResourceId: 42,
    });

    render(
      <ImportCodexModelResourceSelect
        open
        orgSlug="dev-org"
        workerTypeSlug="codex-cli"
        selectedResourceId={42}
        onSelect={onSelect}
        t={(key) => key}
      />,
    );

    expect(screen.getByText("common.loading")).toBeInTheDocument();
    expect(onSelect).not.toHaveBeenCalled();
    await screen.findByText("model resource select");
    expect(getRequirement).toHaveBeenCalledWith("dev-org", "codex-cli");
    expect(useWorkerModelResources).toHaveBeenLastCalledWith(
      "codex-cli",
      42,
      false,
      { protocolAdapters: ["openai-compatible"], required: true },
    );
  });

  it("ignores stale requirement responses when the worker changes", async () => {
    const first = deferredRequirement();
    const second = deferredRequirement();
    getRequirement.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    useWorkerModelResources.mockReturnValue({
      loadingModelResources: false,
      modelResourceError: null,
      modelResources: [],
      selectedModelResourceId: null,
    });

    const view = render(
      <ImportCodexModelResourceSelect
        open
        orgSlug="dev-org"
        workerTypeSlug="codex-cli"
        selectedResourceId={null}
        onSelect={() => {}}
        t={(key) => key}
      />,
    );

    view.rerender(
      <ImportCodexModelResourceSelect
        open
        orgSlug="dev-org"
        workerTypeSlug="pattern-designer"
        selectedResourceId={null}
        onSelect={() => {}}
        t={(key) => key}
      />,
    );
    first.resolve({ modelProtocolAdapters: ["stale"], requiresModelResource: true });
    await waitFor(() => expect(screen.getByText("common.loading")).toBeInTheDocument());
    second.resolve({ modelProtocolAdapters: [], requiresModelResource: false });
    await waitFor(() => expect(screen.queryByText("common.loading")).not.toBeInTheDocument());
    expect(screen.queryByText("model resource select")).not.toBeInTheDocument();
  });
});

function deferredRequirement() {
  let resolve!: (value: {
    modelProtocolAdapters: string[];
    requiresModelResource: boolean;
  }) => void;
  const promise = new Promise<{
    modelProtocolAdapters: string[];
    requiresModelResource: boolean;
  }>((next) => {
    resolve = next;
  });
  return { promise, resolve };
}
