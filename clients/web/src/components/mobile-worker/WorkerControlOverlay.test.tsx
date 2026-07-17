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

function lease() {
  return { ...mocks.lease, acquire: mocks.acquire };
}

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
    render(<WorkerControlOverlay lease={lease()} />);

    fireEvent.click(screen.getByRole("button", { name: "Take control" }));

    expect(mocks.acquire).toHaveBeenCalledTimes(1);
  });

  it("disables control while the relay is disconnected", () => {
    mocks.lease.connected = false;

    render(<WorkerControlOverlay lease={lease()} />);

    expect(screen.getByRole("button", { name: "Take control" })).toBeDisabled();
    expect(screen.getByText("Waiting for the Worker connection.")).toBeInTheDocument();
  });

  it("shows the current controller conflict", () => {
    mocks.lease.status = "busy";

    render(<WorkerControlOverlay lease={lease()} />);

    expect(screen.getByText("Another device has control")).toBeInTheDocument();
  });

  it("does not cover a client that owns the lease", () => {
    mocks.lease.status = "granted";

    const { container } = render(
      <WorkerControlOverlay lease={lease()} />,
    );

    expect(container).toBeEmptyDOMElement();
  });

  it("keeps the panel header interactive in observer mode", () => {
    const { container } = render(
      <WorkerControlOverlay
        lease={lease()}
        preserveHeader
      />,
    );

    expect(container.firstChild).toHaveClass("top-8");
    expect(container.firstChild).not.toHaveClass("top-0");
  });

  it("keeps the workbench browsable in compact observer mode", () => {
    const { container } = render(
      <WorkerControlOverlay blocking={false} lease={lease()} />,
    );

    expect(container.firstChild).toHaveClass("pointer-events-none");
    expect(container.querySelector(".pointer-events-auto")).toBeInTheDocument();
  });
});
