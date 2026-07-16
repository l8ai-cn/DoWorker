function findLineElement(node: Node, container: HTMLElement): HTMLElement | null {
  let current: Node | null = node;
  while (current && current !== container) {
    if (current instanceof HTMLElement && current.dataset.line) return current;
    current = current.parentElement;
  }
  return null;
}

function precedingContentLength(rawLines: string[], lineNumber: number): number {
  let length = 0;
  for (let index = 0; index < lineNumber - 1; index += 1) {
    length += (rawLines[index]?.length ?? 0) + 1;
  }
  return length;
}

function columnOffset(line: HTMLElement, node: Node, offset: number): number {
  const range = document.createRange();
  range.selectNodeContents(line);
  range.setEnd(node, offset);
  return range.toString().length;
}

export function getSelectionOffsets(
  range: Range,
  codeContainer: HTMLElement,
  rawLines: string[],
): { start_index: number; end_index: number } | null {
  const startLine = findLineElement(range.startContainer, codeContainer);
  const endLine = findLineElement(range.endContainer, codeContainer);
  if (!startLine || !endLine) return null;

  const startLineNumber = Number.parseInt(startLine.dataset.line ?? "0", 10);
  const endLineNumber = Number.parseInt(endLine.dataset.line ?? "0", 10);
  if (!startLineNumber || !endLineNumber) return null;

  return {
    start_index:
      precedingContentLength(rawLines, startLineNumber) +
      columnOffset(startLine, range.startContainer, range.startOffset),
    end_index:
      precedingContentLength(rawLines, endLineNumber) +
      columnOffset(endLine, range.endContainer, range.endOffset),
  };
}

export function indexToLine(index: number, rawLines: string[]): number {
  let remaining = index;
  for (let lineIndex = 0; lineIndex < rawLines.length; lineIndex += 1) {
    if (remaining <= rawLines[lineIndex].length) return lineIndex + 1;
    remaining -= rawLines[lineIndex].length + 1;
  }
  return rawLines.length;
}

export function lineOverlapsSelection(
  lineIndex: number,
  rawLines: string[],
  start: number,
  end: number,
): boolean {
  if (lineIndex < 0 || lineIndex >= rawLines.length) return false;
  let lineStart = 0;
  for (let index = 0; index < lineIndex; index += 1) {
    lineStart += rawLines[index].length + 1;
  }
  const lineEnd = lineStart + rawLines[lineIndex].length;
  return start <= lineEnd && end > lineStart;
}
