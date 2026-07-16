import * as Blockly from "blockly";

export const loopTheme = Blockly.Theme.defineTheme("loop", {
  name: "loop",
  base: Blockly.Themes.Classic,
  blockStyles: {
    loop_control_blocks: { colourPrimary: "#2f4858" },
    loop_worker_blocks: { colourPrimary: "#2d6a4f" },
    loop_task_blocks: { colourPrimary: "#34699a" },
    loop_acceptance_blocks: { colourPrimary: "#7b5e2e" },
    loop_verifier_blocks: { colourPrimary: "#6b4c7a" },
    loop_limit_blocks: { colourPrimary: "#8a4f3d" },
    loop_escalation_blocks: { colourPrimary: "#8b3a3a" },
  },
  componentStyles: {
    workspaceBackgroundColour: "#f6f7f9",
    toolboxBackgroundColour: "#ffffff",
    toolboxForegroundColour: "#1f2933",
    flyoutBackgroundColour: "#eef1f4",
    flyoutForegroundColour: "#1f2933",
    scrollbarColour: "#aab2bd",
    insertionMarkerColour: "#111827",
    selectedGlowColour: "#0f766e",
  },
  fontStyle: {
    family: "Inter, ui-sans-serif, system-ui, sans-serif",
    size: 12,
    weight: "500",
  },
});
