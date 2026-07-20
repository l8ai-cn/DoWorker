import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { NlWorkerCreate } from "../NlWorkerCreate";

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}));

describe("NlWorkerCreate", () => {
  it("disables AI fill until the prompt has content", () => {
    const onFill = vi.fn();
    render(
      <NlWorkerCreate
        prompt=""
        filling={false}
        onPromptChange={vi.fn()}
        onFill={onFill}
      />,
    );

    expect(
      screen.getByRole("button", { name: "workers.create.nl.submit" }),
    ).toBeDisabled();
    expect(onFill).not.toHaveBeenCalled();
  });

  it("submits the trimmed prompt", () => {
    const onFill = vi.fn();
    render(
      <NlWorkerCreate
        prompt="  Configure the worker  "
        filling={false}
        onPromptChange={vi.fn()}
        onFill={onFill}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "workers.create.nl.submit" }));

    expect(onFill).toHaveBeenCalledWith("Configure the worker");
  });
});
