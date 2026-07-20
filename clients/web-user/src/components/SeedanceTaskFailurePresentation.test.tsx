import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { Bubble } from "@/lib/renderItems";
import {
  projectSeedanceTaskFailure,
  SeedanceTaskFailurePresentation,
} from "./SeedanceTaskFailurePresentation";

const QUOTA_ERROR =
  "Agent error: rate limit: API error (429): [API_KEY_QUOTA_EXHAUSTED] key=secret";

describe("SeedanceTaskFailurePresentation", () => {
  it("replaces the current Seedance raw quota error with the user projection", () => {
    const projection = projectSeedanceTaskFailure({
      agentId: "seedance-expert",
      agentLabel: "Seedance Expert",
      bubbles: currentTaskBubbles(),
      sessionId: "conv-seedance",
    });

    expect(projection).not.toBeNull();
    render(<SeedanceTaskFailurePresentation projection={projection} />);

    expect(screen.getByText("智能体主模型额度耗尽")).toBeVisible();
    expect(
      screen.getByText("未取得可验证视频文件。"),
    ).toBeVisible();
    expect(screen.getByText("智能体主模型额度不足；未取得可验证视频文件")).toBeVisible();
    expect(screen.queryByText(/API_KEY_QUOTA_EXHAUSTED|key=secret|429/)).not.toBeInTheDocument();
    expect(screen.queryByText(/Provider 尚未调用/)).not.toBeInTheDocument();
  });

  it("does not project a prior failure after a newer task begins", () => {
    const bubbles = [
      ...currentTaskBubbles(),
      userBubble("user-new"),
    ];

    expect(
      projectSeedanceTaskFailure({
        agentId: "seedance-expert",
        agentLabel: "Seedance Expert",
        bubbles,
        sessionId: "conv-seedance",
      }),
    ).toBeNull();
  });

  it("does not infer a video task for another agent", () => {
    expect(
      projectSeedanceTaskFailure({
        agentId: "e2e-echo",
        agentLabel: "E2E Echo",
        bubbles: currentTaskBubbles(),
        sessionId: "conv-echo",
      }),
    ).toBeNull();
  });
});

function currentTaskBubbles(): Bubble[] {
  return [
    userBubble("user-current"),
    {
      error: null,
      items: [{
        final: true,
        itemId: "assistant-failure",
        kind: "text",
        text: QUOTA_ERROR,
      }],
      kind: "assistant",
      lifecycle: "completed",
      responseId: "response-current",
      stableId: "assistant-current",
    },
  ];
}

function userBubble(itemId: string): Extract<Bubble, { kind: "user" }> {
  return {
    content: [{ text: "生成视频", type: "input_text" }],
    itemId,
    kind: "user",
  };
}
