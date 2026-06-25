import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { RelayStatusOverlay } from "../RelayStatusOverlay";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      web: "Web",
      relay: "Relay",
      runner: "Runner",
      connected: "Connected",
      connecting: "Connecting",
      disconnected: "Disconnected",
      error: "Error",
      unknown: "Unknown",
    };
    return translations[key] || key;
  },
}));

describe("RelayStatusOverlay", () => {
  describe("all connected (green-green)", () => {
    it("shows Web, Relay, Runner labels", () => {
      render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={false} />
      );
      expect(screen.getByText("Web")).toBeInTheDocument();
      expect(screen.getByText("Relay")).toBeInTheDocument();
      expect(screen.getByText("Runner")).toBeInTheDocument();
    });

    it("both segment dots are green", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots).toHaveLength(2);
      dots.forEach((dot) => expect(dot).toHaveClass("bg-success"));
    });

    it("Web and Runner labels are green", () => {
      render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={false} />
      );
      expect(screen.getByText("Web")).toHaveClass("text-success");
      expect(screen.getByText("Runner")).toHaveClass("text-success");
    });

    it("Relay label is neutral gray", () => {
      render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={false} />
      );
      expect(screen.getByText("Relay")).toHaveClass("text-muted-foreground");
    });

    it("applies green overall background", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={false} />
      );
      expect(container.querySelector(".bg-success\\/15")).toBeInTheDocument();
    });
  });

  describe("connecting state (yellow-gray)", () => {
    it("web-relay dot is yellow with pulse", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connecting" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[0]).toHaveClass("bg-warning", "animate-pulse");
    });

    it("web-relay tooltip says Connecting", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connecting" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[0]).toHaveAttribute("title", "Connecting");
    });

    it("relay-runner dot is gray (unknown)", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connecting" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[1]).toHaveClass("bg-muted-foreground");
    });

    it("relay-runner tooltip says Unknown", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connecting" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[1]).toHaveAttribute("title", "Unknown");
    });

    it("applies yellow overall background", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connecting" isRunnerDisconnected={false} />
      );
      expect(container.querySelector(".bg-warning\\/15")).toBeInTheDocument();
    });
  });

  describe("relay disconnected (red-gray)", () => {
    it("web-relay dot is red", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="disconnected" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[0]).toHaveClass("bg-danger");
    });

    it("web-relay tooltip says Disconnected", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="disconnected" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[0]).toHaveAttribute("title", "Disconnected");
    });

    it("relay-runner dot is gray", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="disconnected" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[1]).toHaveClass("bg-muted-foreground");
    });

    it("applies red overall background", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="disconnected" isRunnerDisconnected={false} />
      );
      expect(container.querySelector(".bg-danger\\/15")).toBeInTheDocument();
    });
  });

  describe("relay error (red-gray)", () => {
    it("web-relay dot is red", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="error" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[0]).toHaveClass("bg-danger");
    });

    it("web-relay tooltip says Error (not Disconnected)", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="error" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[0]).toHaveAttribute("title", "Error");
    });

    it("relay-runner dot is gray", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="error" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[1]).toHaveClass("bg-muted-foreground");
    });

    it("applies red overall background", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="error" isRunnerDisconnected={false} />
      );
      expect(container.querySelector(".bg-danger\\/15")).toBeInTheDocument();
    });
  });

  describe("runner disconnected (green-red)", () => {
    it("Web label is green (its segment is ok)", () => {
      render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={true} />
      );
      expect(screen.getByText("Web")).toHaveClass("text-success");
    });

    it("Runner label is red (its segment is broken)", () => {
      render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={true} />
      );
      expect(screen.getByText("Runner")).toHaveClass("text-danger");
    });

    it("web-relay dot is green", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={true} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[0]).toHaveClass("bg-success");
    });

    it("web-relay tooltip says Connected", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={true} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[0]).toHaveAttribute("title", "Connected");
    });

    it("relay-runner dot is red", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={true} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[1]).toHaveClass("bg-danger");
    });

    it("relay-runner tooltip says Disconnected", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={true} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[1]).toHaveAttribute("title", "Disconnected");
    });

    it("applies red overall background (worst segment)", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={true} />
      );
      expect(container.querySelector(".bg-danger\\/15")).toBeInTheDocument();
    });
  });

  describe("overlay positioning", () => {
    it("renders as absolute positioned overlay", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={false} />
      );
      const overlay = container.firstChild as HTMLElement;
      expect(overlay).toHaveClass("absolute", "top-0", "left-0", "right-0", "z-10");
    });

    it("is not interactive (pointer-events-none)", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={false} />
      );
      const overlay = container.firstChild as HTMLElement;
      expect(overlay).toHaveClass("pointer-events-none");
    });
  });

  describe("className prop", () => {
    it("applies custom className", () => {
      const { container } = render(
        <RelayStatusOverlay
          connectionStatus="connected"
          isRunnerDisconnected={false}
          className="custom-class"
        />
      );
      const overlay = container.firstChild as HTMLElement;
      expect(overlay).toHaveClass("custom-class");
    });
  });

  describe("segment dot accessibility", () => {
    it("dots have role=status and aria-label", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots).toHaveLength(2);
      dots.forEach((dot) => expect(dot).toHaveAttribute("aria-label"));
    });

    it("dots have title tooltips", () => {
      const { container } = render(
        <RelayStatusOverlay connectionStatus="connected" isRunnerDisconnected={false} />
      );
      const dots = container.querySelectorAll("[role='status']");
      expect(dots[0]).toHaveAttribute("title", "Connected");
      expect(dots[1]).toHaveAttribute("title", "Connected");
    });
  });
});
