import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { NextIntlClientProvider } from "next-intl";
import commonMessages from "@/messages/en/common.json";
import loopsMessages from "@/messages/en/loops.json";
import { LoopNlCreate } from "../LoopNlCreate";
import { buildLoopAiGuidePrompt } from "../loop-ai-guide-prompt";

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
      messages={{ ...commonMessages, ...loopsMessages }}
    >
      {children}
    </NextIntlClientProvider>
  );
}

describe("LoopNlCreate", () => {
  const onNeedsWizard = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the AI guide section", () => {
    render(<LoopNlCreate onNeedsWizard={onNeedsWizard} />, { wrapper: Wrapper });
    expect(screen.getByText(loopsMessages.loops.aiGuideTitle)).toBeInTheDocument();
    expect(
      screen.getByPlaceholderText(loopsMessages.loops.aiGuidePlaceholder),
    ).toBeInTheDocument();
  });

  it("fills textarea from an example chip", () => {
    render(<LoopNlCreate onNeedsWizard={onNeedsWizard} />, { wrapper: Wrapper });
    fireEvent.click(screen.getByText(loopsMessages.loops.aiGuideExample1));
    const textarea = screen.getByPlaceholderText(
      loopsMessages.loops.aiGuidePlaceholder,
    ) as HTMLTextAreaElement;
    expect(textarea.value).toBe(loopsMessages.loops.aiGuideExample1);
  });

  it("submits looper prompt and navigates to workspace", async () => {
    mockCreate.mockResolvedValue({ pod_key: "acme-guide-01" });
    render(<LoopNlCreate onNeedsWizard={onNeedsWizard} />, { wrapper: Wrapper });

    fireEvent.change(screen.getByPlaceholderText(loopsMessages.loops.aiGuidePlaceholder), {
      target: { value: "watch CI failures" },
    });
    fireEvent.click(screen.getByText(loopsMessages.loops.aiGuideStart));

    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledWith({
        prompt: buildLoopAiGuidePrompt("watch CI failures"),
        alias: loopsMessages.loops.aiGuidePodAlias,
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
    render(<LoopNlCreate onNeedsWizard={onNeedsWizard} />, { wrapper: Wrapper });

    fireEvent.change(screen.getByPlaceholderText(loopsMessages.loops.aiGuidePlaceholder), {
      target: { value: "some idea" },
    });
    fireEvent.click(screen.getByText(loopsMessages.loops.aiGuideStart));

    await waitFor(() => {
      expect(onNeedsWizard).toHaveBeenCalledWith("some idea");
    });
    expect(mockPush).not.toHaveBeenCalled();
  });
});
