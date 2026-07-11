import { fireEvent, render, screen } from "@/test/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { WorkerControlOverlay } from "./WorkerControlOverlay";

const mocks = vi.hoisted(() => ({
  acquire: vi.fn(),
  lease: {
    status: "observer",
    connected: true,
    acquiring: false,
    error: null,
  },
}));

vi.mock("@/hooks/useWorkerControlLease", () => ({
  useWorkerControlLease: () => ({
    ...mocks.lease,
    acquire: mocks.acquire,
  }),
}));

describe("WorkerControlOverlay", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.lease = {
      status: "observer",
      connected: true,
      acquiring: false,
      error: null,
    };
  });

  it("requests control from observer mode", () => {
    render(<WorkerControlOverlay podKey="pod-1" clientLabel="mobile" />);

    fireEvent.click(screen.getByRole("button", { name: "Take control" }));

    expect(mocks.acquire).toHaveBeenCalledTimes(1);
  });

  it("disables control while the relay is disconnected", () => {
    mocks.lease.connected = false;

    render(<WorkerControlOverlay podKey="pod-1" clientLabel="mobile" />);

    expect(screen.getByRole("button", { name: "Take control" })).toBeDisabled();
    expect(screen.getByText("Waiting for the Worker connection.")).toBeInTheDocument();
  });

  it("shows the current controller conflict", () => {
    mocks.lease.status = "busy";

    render(<WorkerControlOverlay podKey="pod-1" clientLabel="mobile" />);

    expect(screen.getByText("Another device has control")).toBeInTheDocument();
  });

  it("does not cover a client that owns the lease", () => {
    mocks.lease.status = "granted";

    const { container } = render(
      <WorkerControlOverlay podKey="pod-1" clientLabel="mobile" />,
    );

    expect(container).toBeEmptyDOMElement();
  });

  it("keeps the panel header interactive in observer mode", () => {
    const { container } = render(
      <WorkerControlOverlay
        podKey="pod-1"
        clientLabel="desktop"
        preserveHeader
      />,
    );

    expect(container.firstChild).toHaveClass("top-8");
    expect(container.firstChild).not.toHaveClass("top-0");
  });
});
