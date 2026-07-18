import { useState } from "react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { render, screen } from "@/test/test-utils";
import { ResourceStringMapField } from "./ResourceStringMapField";

function StringMapHarness() {
  const [value, setValue] = useState({ "input-1": "" });
  return (
    <ResourceStringMapField
      label="Inputs"
      value={value}
      onChange={setValue}
    />
  );
}

function TwoRowHarness() {
  const [value, setValue] = useState({
    "input-1": "first",
    "input-2": "second",
  });
  return (
    <ResourceStringMapField
      label="Inputs"
      value={value}
      onChange={setValue}
    />
  );
}

describe("resource collection row identity", () => {
  it("keeps the edited row mounted while its business key changes", async () => {
    const user = userEvent.setup();
    render(<StringMapHarness />);
    const keyInput = screen.getAllByRole("textbox")[0];

    await user.type(keyInput, "-renamed");

    expect(screen.getAllByRole("textbox")[0]).toBe(keyInput);
    expect(keyInput).toHaveValue("input-1-renamed");
    expect(keyInput).toHaveFocus();
  });

  it("preserves a later row when an earlier row is removed", async () => {
    const user = userEvent.setup();
    render(<TwoRowHarness />);
    const secondKeyInput = screen.getAllByRole("textbox")[2];

    await user.click(screen.getByRole("button", {
      name: "Remove input-1",
    }));

    expect(screen.getAllByRole("textbox")[0]).toBe(secondKeyInput);
    expect(secondKeyInput).toHaveValue("input-2");
  });
});
