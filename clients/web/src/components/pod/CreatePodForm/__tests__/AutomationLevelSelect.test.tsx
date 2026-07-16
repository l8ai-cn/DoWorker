import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { AutomationLevelSelect } from "../AutomationLevelSelect";

const t = (key: string) =>
  key === "ide.createPod.automationLevel.autonomousPtyHint"
    ? "终端全自动运行"
    : key;

describe("AutomationLevelSelect", () => {
  it("uses terminal wording for PTY-only Workers", () => {
    render(
      <AutomationLevelSelect
        value="autonomous"
        onChange={() => undefined}
        supportedModes={["pty"]}
        t={t}
      />,
    );

    expect(screen.getByText("终端全自动运行")).toBeInTheDocument();
  });
});
