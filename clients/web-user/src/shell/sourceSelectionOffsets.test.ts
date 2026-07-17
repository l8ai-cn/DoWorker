import { describe, expect, it } from "vitest";
import {
  getSelectionOffsets,
  indexToLine,
  lineOverlapsSelection,
} from "./sourceSelectionOffsets";

describe("indexToLine", () => {
  const lines = ["hello", "world", "foo"];

  it("maps the first character", () => {
    expect(indexToLine(0, lines)).toBe(1);
  });

  it("maps the last character of the first line", () => {
    expect(indexToLine(4, lines)).toBe(1);
  });

  it("attributes a newline to the preceding line", () => {
    expect(indexToLine(5, lines)).toBe(1);
  });

  it("maps the first character of the second line", () => {
    expect(indexToLine(6, lines)).toBe(2);
  });

  it("maps the final line", () => {
    expect(indexToLine(13, lines)).toBe(3);
  });

  it("clamps beyond EOF", () => {
    expect(indexToLine(999, lines)).toBe(3);
  });

  it("handles a single line", () => {
    expect(indexToLine(3, ["abcdef"])).toBe(1);
  });

  it("handles an empty file", () => {
    expect(indexToLine(0, [])).toBe(0);
  });

  it("handles empty lines", () => {
    expect(indexToLine(0, ["", "x"])).toBe(1);
    expect(indexToLine(1, ["", "x"])).toBe(2);
  });
});

describe("lineOverlapsSelection", () => {
  const lines = ["ab", "cd", "ef"];

  it("detects a fully covered line", () => {
    expect(lineOverlapsSelection(0, lines, 0, 8)).toBe(true);
  });

  it("detects a same-line selection", () => {
    expect(lineOverlapsSelection(0, lines, 0, 2)).toBe(true);
  });

  it("detects a multi-line selection", () => {
    expect(lineOverlapsSelection(1, lines, 1, 4)).toBe(true);
  });

  it("respects an exclusive end at the line start", () => {
    expect(lineOverlapsSelection(1, lines, 0, 3)).toBe(false);
  });

  it("rejects a selection before the line", () => {
    expect(lineOverlapsSelection(2, lines, 0, 2)).toBe(false);
  });

  it("rejects a selection after the line", () => {
    expect(lineOverlapsSelection(0, lines, 3, 5)).toBe(false);
  });

  it("detects a single-character selection", () => {
    expect(lineOverlapsSelection(1, lines, 3, 4)).toBe(true);
  });

  it("detects all lines in a file-wide selection", () => {
    expect(lineOverlapsSelection(0, lines, 0, 8)).toBe(true);
    expect(lineOverlapsSelection(1, lines, 0, 8)).toBe(true);
    expect(lineOverlapsSelection(2, lines, 0, 8)).toBe(true);
  });
});

function buildContainer(rawLines: string[]): HTMLElement {
  const container = document.createElement("div");
  rawLines.forEach((line, index) => {
    const element = document.createElement("div");
    element.dataset.line = String(index + 1);
    element.textContent = line;
    container.appendChild(element);
  });
  document.body.appendChild(container);
  return container;
}

function lineTextNode(container: HTMLElement, lineIndex: number): Text {
  return container.children[lineIndex].firstChild as Text;
}

describe("getSelectionOffsets", () => {
  it("computes a single-line selection", () => {
    const rawLines = ["hello", "world", "foo"];
    const container = buildContainer(rawLines);
    const range = document.createRange();
    range.setStart(lineTextNode(container, 0), 1);
    range.setEnd(lineTextNode(container, 0), 4);
    expect(getSelectionOffsets(range, container, rawLines)).toEqual({
      start_index: 1,
      end_index: 4,
    });
    container.remove();
  });

  it("sums preceding lines for a multi-line selection", () => {
    const rawLines = ["hello", "world", "foo"];
    const container = buildContainer(rawLines);
    const range = document.createRange();
    range.setStart(lineTextNode(container, 0), 2);
    range.setEnd(lineTextNode(container, 1), 3);
    expect(getSelectionOffsets(range, container, rawLines)).toEqual({
      start_index: 2,
      end_index: 9,
    });
    container.remove();
  });

  it("rejects a boundary outside a data-line element", () => {
    const rawLines = ["hello"];
    const container = buildContainer(rawLines);
    const stray = document.createElement("span");
    stray.textContent = "outside";
    document.body.appendChild(stray);
    const range = document.createRange();
    range.setStart(stray.firstChild as Text, 0);
    range.setEnd(stray.firstChild as Text, 3);
    expect(getSelectionOffsets(range, container, rawLines)).toBeNull();
    container.remove();
    stray.remove();
  });

  it("rejects a zero line number", () => {
    const container = document.createElement("div");
    const line = document.createElement("div");
    line.dataset.line = "0";
    line.textContent = "abc";
    container.appendChild(line);
    document.body.appendChild(container);
    const range = document.createRange();
    range.setStart(line.firstChild as Text, 0);
    range.setEnd(line.firstChild as Text, 2);
    expect(getSelectionOffsets(range, container, ["abc"])).toBeNull();
    container.remove();
  });
});
