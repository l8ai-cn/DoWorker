import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@/test/test-utils";
import {
  PaneLoadingState,
  PaneErrorState,
  PaneReconnectingState,
} from "../PaneStateViews";

describe("PaneLoadingState", () => {
  const defaultProps = {
    podStatus: "initializing",
    onClose: vi.fn(),
  };

  describe("loading state (non-completed)", () => {
    it("shows spinner and waiting message", () => {
      render(<PaneLoadingState {...defaultProps} />);

      expect(screen.getByText("Waiting for Pod to be ready...")).toBeInTheDocument();
      expect(screen.queryByText("Pod completed")).not.toBeInTheDocument();
    });

    it("shows status text with yellow styling", () => {
      render(<PaneLoadingState {...defaultProps} podStatus="initializing" />);

      const statusText = screen.getByText("initializing");
      expect(statusText).toBeInTheDocument();
      expect(statusText).toHaveClass("text-warning");
    });

    it("does not show close button for initializing status", () => {
      render(<PaneLoadingState {...defaultProps} podStatus="initializing" />);

      expect(screen.queryByText("Close")).not.toBeInTheDocument();
    });

    it("does not show close button for running status", () => {
      render(<PaneLoadingState {...defaultProps} podStatus="running" />);

      expect(screen.queryByText("Close")).not.toBeInTheDocument();
    });

    it("shows init progress when provided", () => {
      const initProgress = { progress: 50, phase: "Cloning", message: "Cloning repository..." };
      render(<PaneLoadingState {...defaultProps} initProgress={initProgress} />);

      expect(screen.getByText("Cloning repository...")).toBeInTheDocument();
      expect(screen.getByText("Cloning - 50%")).toBeInTheDocument();
    });
  });

  describe("unknown status", () => {
    it("shows close button", () => {
      render(<PaneLoadingState {...defaultProps} podStatus="unknown" />);

      expect(screen.getByText("Close")).toBeInTheDocument();
    });

    it("calls onClose when close button is clicked", () => {
      const onClose = vi.fn();
      render(<PaneLoadingState {...defaultProps} podStatus="unknown" onClose={onClose} />);

      fireEvent.click(screen.getByText("Close"));
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });

  describe("completed status", () => {
    it("shows 'Pod completed' text instead of waiting message", () => {
      render(<PaneLoadingState {...defaultProps} podStatus="completed" />);

      expect(screen.getByText("Pod completed")).toBeInTheDocument();
      expect(screen.queryByText("Waiting for Pod to be ready...")).not.toBeInTheDocument();
    });

    it("shows status text with green styling", () => {
      render(<PaneLoadingState {...defaultProps} podStatus="completed" />);

      const statusText = screen.getByText("completed");
      expect(statusText).toBeInTheDocument();
      expect(statusText).toHaveClass("text-success");
    });

    it("shows close button", () => {
      render(<PaneLoadingState {...defaultProps} podStatus="completed" />);

      expect(screen.getByText("Close")).toBeInTheDocument();
    });

    it("calls onClose when close button is clicked", () => {
      const onClose = vi.fn();
      render(<PaneLoadingState {...defaultProps} podStatus="completed" onClose={onClose} />);

      fireEvent.click(screen.getByText("Close"));
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it("does not show close button when onClose is not provided", () => {
      render(
        <PaneLoadingState
          podStatus="completed"
        />
      );

      expect(screen.queryByText("Close")).not.toBeInTheDocument();
    });
  });
});

describe("PaneErrorState", () => {
  it("shows error message", () => {
    render(<PaneErrorState error="Pod failed" />);

    expect(screen.getByText("Pod failed")).toBeInTheDocument();
    expect(
      screen.getByText("The pod cannot be connected. Please check the pod status or create a new one.")
    ).toBeInTheDocument();
  });

  it("shows close button when onClose is provided", () => {
    render(<PaneErrorState error="Pod failed" onClose={vi.fn()} />);

    expect(screen.getByText("Close")).toBeInTheDocument();
  });

  it("calls onClose when close button is clicked", () => {
    const onClose = vi.fn();
    render(<PaneErrorState error="Pod failed" onClose={onClose} />);

    fireEvent.click(screen.getByText("Close"));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("does not show close button when onClose is not provided", () => {
    render(<PaneErrorState error="Pod terminated" />);

    expect(screen.queryByText("Close")).not.toBeInTheDocument();
  });
});

describe("PaneReconnectingState", () => {
  it("renders the workspace reconnecting messages", () => {
    render(<PaneReconnectingState />);

    expect(screen.getByText("Runner is restarting...")).toBeInTheDocument();
    expect(
      screen.getByText("Session will resume automatically. Your work is preserved.")
    ).toBeInTheDocument();
  });
});
