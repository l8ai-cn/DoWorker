import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@/test/test-utils";

const navigation = vi.hoisted(() => ({
  search: "",
}));

vi.mock("next/navigation", () => ({
  useParams: () => ({ org: "acme" }),
  usePathname: () => "/acme/workers/new",
  useRouter: () => ({ push: vi.fn() }),
  useSearchParams: () => new URLSearchParams(navigation.search),
}));

vi.mock("@/components/workers/ImportCodexDialog", () => ({
  ImportCodexDialog: () => null,
}));

vi.mock("@/components/pod/CreatePodForm", () => ({
  CreatePodForm: () => <div data-testid="simple-worker-create-form" />,
}));

vi.mock("@/components/resource-editor/ResourceDependencyEditor", () => ({
  ResourceDependencyEditor: () => <div data-testid="resource-editor-resources" />,
}));

vi.mock("@/components/resource-editor/ResourceEditorShell", () => ({
  ResourceEditorShell: ({ kind }: { kind?: string }) => (
    <div data-testid={kind === "Worker"
      ? "resource-editor-run"
      : "resource-editor-template"}
    />
  ),
}));

import { CreateWorkerPageContent } from "./CreateWorkerPageContent";

describe("CreateWorkerPageContent", () => {
  beforeEach(() => {
    navigation.search = "";
    vi.spyOn(window.history, "replaceState").mockImplementation((
      _state,
      _unused,
      url,
    ) => {
      navigation.search = String(url).split("?")[1] ?? "";
    });
  });

  it("restores the selected mode after the page remounts", async () => {
    const user = userEvent.setup();
    const view = render(<CreateWorkerPageContent />);

    await user.click(screen.getByTestId("pill-tab-template"));
    expect(screen.getByTestId("resource-editor-template")).toBeInTheDocument();
    expect(window.history.replaceState).toHaveBeenCalledWith(
      window.history.state,
      "",
      "/acme/workers/new?mode=template",
    );

    view.unmount();
    render(<CreateWorkerPageContent />);

    expect(screen.getByTestId("pill-tab-template"))
      .toHaveAttribute("aria-selected", "true");
    expect(screen.getByTestId("resource-editor-template")).toBeInTheDocument();
  });

  it("uses the simple Worker creation form by default", () => {
    render(<CreateWorkerPageContent />);

    expect(screen.getByTestId("simple-worker-create-form")).toBeInTheDocument();
    expect(screen.queryByTestId("resource-editor-run")).not.toBeInTheDocument();
  });
});
