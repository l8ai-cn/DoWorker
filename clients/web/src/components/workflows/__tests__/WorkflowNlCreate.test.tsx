import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { NextIntlClientProvider } from "next-intl";
import commonMessages from "@/messages/en/common.json";
import workflowsMessages from "@/messages/en/workflows.json";
import { WorkflowNlCreate } from "../WorkflowNlCreate";
import { buildWorkflowAiGuidePrompt } from "../workflow-ai-guide-prompt";

const mockPush = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush, replace: vi.fn(), prefetch: vi.fn() }),
}));

const mockCreate = vi.fn();
vi.mock("@/lib/api/quickTaskApi", () => ({
  quickTaskApi: { create: (...args: unknown[]) => mockCreate(...args) },
}));

const mockGetApiErrorCode = vi.fn();
vi.mock("@/lib/api", () => ({
  getApiErrorCode: (...args: unknown[]) => mockGetApiErrorCode(...args),
  getLocalizedErrorMessage: () => "error",
}));

vi.mock("@/stores/auth", () => ({
  useCurrentOrg: () => ({ slug: "acme" }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), info: vi.fn(), error: vi.fn() },
}));

function Wrapper({ children }: { children: ReactNode }) {
  return (
    <NextIntlClientProvider
      locale="en"
      messages={{ ...commonMessages, ...workflowsMessages }}
    >
      {children}
    </NextIntlClientProvider>
  );
}

describe("WorkflowNlCreate", () => {
  const onNeedsWizard = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the AI guide section", () => {
    render(<WorkflowNlCreate onNeedsWizard={onNeedsWizard} />, { wrapper: Wrapper });
    expect(screen.getByText(workflowsMessages.workflows.aiGuideTitle)).toBeInTheDocument();
    expect(
      screen.getByPlaceholderText(workflowsMessages.workflows.aiGuidePlaceholder),
    ).toBeInTheDocument();
  });

  it("fills textarea from an example chip", () => {
    render(<WorkflowNlCreate onNeedsWizard={onNeedsWizard} />, { wrapper: Wrapper });
    fireEvent.click(screen.getByText(workflowsMessages.workflows.aiGuideExample1));
    const textarea = screen.getByPlaceholderText(
      workflowsMessages.workflows.aiGuidePlaceholder,
    ) as HTMLTextAreaElement;
    expect(textarea.value).toBe(workflowsMessages.workflows.aiGuideExample1);
  });

  it("submits workflower prompt and navigates to workspace", async () => {
    mockCreate.mockResolvedValue({ pod_key: "acme-guide-01" });
    render(<WorkflowNlCreate onNeedsWizard={onNeedsWizard} />, { wrapper: Wrapper });

    fireEvent.change(screen.getByPlaceholderText(workflowsMessages.workflows.aiGuidePlaceholder), {
      target: { value: "watch CI failures" },
    });
    fireEvent.click(screen.getByText(workflowsMessages.workflows.aiGuideStart));

    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledWith({
        prompt: buildWorkflowAiGuidePrompt("watch CI failures"),
        alias: workflowsMessages.workflows.aiGuidePodAlias,
      });
    });
    expect(mockCreate.mock.calls[0][0].prompt).toContain("LOOP-WORTHINESS GATE");
    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/acme/workspace?pod=acme-guide-01");
    });
    expect(onNeedsWizard).not.toHaveBeenCalled();
  });

  it("calls onNeedsWizard with idea when runner is unavailable", async () => {
    mockCreate.mockRejectedValue(new Error("no runner"));
    mockGetApiErrorCode.mockReturnValue("NO_RUNNER_FOR_AGENT");
    render(<WorkflowNlCreate onNeedsWizard={onNeedsWizard} />, { wrapper: Wrapper });

    fireEvent.change(screen.getByPlaceholderText(workflowsMessages.workflows.aiGuidePlaceholder), {
      target: { value: "some idea" },
    });
    fireEvent.click(screen.getByText(workflowsMessages.workflows.aiGuideStart));

    await waitFor(() => {
      expect(onNeedsWizard).toHaveBeenCalledWith("some idea");
    });
    expect(mockPush).not.toHaveBeenCalled();
  });
});
