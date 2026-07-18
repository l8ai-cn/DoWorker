import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { EmbedSessionClient } from "@/embed-session-api";
import { EmbeddedSessionTimeline } from "./EmbeddedSessionTimeline";

class ResizeObserverStub {
  disconnect() {}
  observe() {}
  unobserve() {}
}

vi.stubGlobal("ResizeObserver", ResizeObserverStub);

vi.mock("./useEmbeddedSessionTimeline", () => ({
  useEmbeddedSessionTimeline: () => ({
    state: {
      activeResponse: null,
      blocks: [],
      error: null,
      isLoading: false,
      isSending: false,
      session: { id: "session-1", status: "idle", title: "Embedded review" },
      status: "idle",
    },
    sendMessage: vi.fn(),
  }),
}));

const readOnlyClient: EmbedSessionClient = {
  getItems: vi.fn(),
  getSession: vi.fn(),
  openStream: vi.fn(),
};

describe("EmbeddedSessionTimeline", () => {
  it("keeps the composer disabled for a read-only session", () => {
    render(<EmbeddedSessionTimeline client={readOnlyClient} />);

    expect(screen.getByLabelText("Message the agent")).toBeDisabled();
    expect(screen.getByLabelText("Send message")).toBeDisabled();
    expect(screen.getByPlaceholderText("Read-only session")).toBeInTheDocument();
  });
});
