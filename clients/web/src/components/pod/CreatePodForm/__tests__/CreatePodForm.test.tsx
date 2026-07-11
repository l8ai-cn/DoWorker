import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { CreatePodForm } from "../index";
import {
  controllerFixture,
  mockFillWithAI,
  mockRepository,
  mockSetFillPrompt,
} from "./test-utils";

const mockUseWorkerCreateDraft = vi.fn((params?: unknown) => {
  void params;
  return controllerFixture();
});
const mockFetchRepositories = vi.fn(async () => undefined);

vi.mock("../../hooks", () => ({
  useWorkerCreateDraft: (params: unknown) => mockUseWorkerCreateDraft(params),
}));
vi.mock("@/stores/repository", () => ({
  useRepositories: () => [mockRepository],
  useRepositoryStore: (selector: (state: unknown) => unknown) =>
    selector({
      fetched: true,
      isLoading: false,
      fetchRepositories: mockFetchRepositories,
    }),
}));
vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("CreatePodForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseWorkerCreateDraft.mockReturnValue(controllerFixture());
  });

  it("renders one four-step workflow and preserves the supplied class", () => {
    const { container } = render(
      <CreatePodForm
        className="worker-form"
        config={{ scenario: "workspace" }}
      />,
    );

    expect(container.firstChild).toHaveClass("worker-form");
    expect(screen.getAllByText("workerCreate.steps.runtime").length).toBeGreaterThan(0);
    expect(screen.getAllByText("workerCreate.steps.typeConfig").length).toBeGreaterThan(0);
    expect(screen.getAllByText("workerCreate.steps.workspace").length).toBeGreaterThan(0);
    expect(screen.getAllByText("workerCreate.steps.preflight").length).toBeGreaterThan(0);
  });

  it("uses the natural-language panel to fill the same controller", () => {
    render(<CreatePodForm config={{ scenario: "workspace" }} />);

    fireEvent.change(screen.getByPlaceholderText("workers.create.nl.placeholder"), {
      target: { value: "Review authentication" },
    });
    expect(mockSetFillPrompt).toHaveBeenCalledWith("Review authentication");

    mockUseWorkerCreateDraft.mockReturnValue(controllerFixture({
      state: { fillPrompt: "Review authentication" },
    }));
    const { rerender } = render(
      <CreatePodForm config={{ scenario: "workspace" }} />,
    );
    rerender(<CreatePodForm config={{ scenario: "workspace" }} />);
    fireEvent.click(screen.getAllByText("workers.create.nl.submit").at(-1)!);
    expect(mockFillWithAI).toHaveBeenCalledWith("Review authentication");
  });

  it("passes ticket context and generated task into the WorkerSpec hook", () => {
    render(
      <CreatePodForm
        config={{
          scenario: "ticket",
          context: {
            ticket: {
              id: 7,
              slug: "TASK-7",
              title: "Fix flaky test",
              repositoryId: 51,
            },
          },
        }}
      />,
    );

    expect(mockUseWorkerCreateDraft).toHaveBeenLastCalledWith(
      expect.objectContaining({
        initialTask: "Work on ticket TASK-7: Fix flaky test",
        initialRepositoryId: 51,
        ticketSlug: "TASK-7",
      }),
    );
  });

  it("renders and invokes cancel only when configured", () => {
    const onCancel = vi.fn();
    const { rerender } = render(
      <CreatePodForm config={{ scenario: "workspace", onCancel }} />,
    );
    fireEvent.click(screen.getByText("ide.createPod.cancel"));
    expect(onCancel).toHaveBeenCalledOnce();

    rerender(<CreatePodForm config={{ scenario: "workspace" }} />);
    expect(screen.queryByText("ide.createPod.cancel")).not.toBeInTheDocument();
  });
});
