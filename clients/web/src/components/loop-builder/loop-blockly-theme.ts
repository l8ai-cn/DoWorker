import * as Blockly from "blockly";

export const loopBlocklyTheme = Blockly.Theme.defineTheme("loop", {
  name: "loop",
  base: Blockly.Themes.Classic,
  componentStyles: {
    workspaceBackgroundColour: "#f8fafc",
    toolboxBackgroundColour: "#ffffff",
    toolboxForegroundColour: "#334155",
    flyoutBackgroundColour: "#f1f5f9",
    flyoutForegroundColour: "#334155",
    scrollbarColour: "#94a3b8",
    insertionMarkerColour: "#0f766e",
    insertionMarkerOpacity: 0.35,
  },
});
