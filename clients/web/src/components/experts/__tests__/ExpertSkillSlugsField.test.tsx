import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { ExpertSkillSlugsField } from "../ExpertSkillSlugsField";

function setup(value: string[] = []) {
  const onChange = vi.fn();
  render(
    <ExpertSkillSlugsField
      value={value}
      onChange={onChange}
      emptyLabel="empty"
      addLabel="add"
      placeholder="slug"
      removeLabel="remove"
    />,
  );
  return { onChange };
}

describe("ExpertSkillSlugsField", () => {
  it("normalizes and adds a slug on Enter", () => {
    const { onChange } = setup([]);
    const input = screen.getByPlaceholderText("slug");
    fireEvent.change(input, { target: { value: "Code Review" } });
    fireEvent.keyDown(input, { key: "Enter" });
    expect(onChange).toHaveBeenCalledWith(["code-review"]);
  });

  it("adds via the add button", () => {
    const { onChange } = setup(["a"]);
    fireEvent.change(screen.getByPlaceholderText("slug"), { target: { value: "b" } });
    fireEvent.click(screen.getByText("add"));
    expect(onChange).toHaveBeenCalledWith(["a", "b"]);
  });

  it("skips duplicates", () => {
    const { onChange } = setup(["dup"]);
    fireEvent.change(screen.getByPlaceholderText("slug"), { target: { value: "dup" } });
    fireEvent.keyDown(screen.getByPlaceholderText("slug"), { key: "Enter" });
    expect(onChange).not.toHaveBeenCalled();
  });

  it("removes a slug via its remove button", () => {
    const { onChange } = setup(["keep", "drop"]);
    const removeButtons = screen.getAllByLabelText("remove");
    fireEvent.click(removeButtons[1]);
    expect(onChange).toHaveBeenCalledWith(["keep"]);
  });

  it("shows empty label when no slugs", () => {
    setup([]);
    expect(screen.getByText("empty")).toBeInTheDocument();
  });
});
