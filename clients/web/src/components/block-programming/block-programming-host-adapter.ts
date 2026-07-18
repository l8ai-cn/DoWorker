import type * as Blockly from "blockly";

export interface BlockProgrammingSourceResult {
  source: string;
  complete: boolean;
  issues: string[];
  nodeIndex: Map<string, string>;
}

export interface BlockProgrammingHostAdapter<
  TProgram,
  TMessages,
  TCustomDefinition,
> {
  namespace: string;
  customBlockType: (definition: TCustomDefinition) => string;
  createCatalog: (
    messages: TMessages,
    customDefinitions: readonly TCustomDefinition[],
  ) => { toolbox: Blockly.utils.toolbox.ToolboxDefinition; definitions: object[] };
  registerBlocks: (
    messages?: TMessages,
    customDefinitions?: readonly TCustomDefinition[],
  ) => void;
  projectProgram: (
    workspace: Blockly.Workspace,
    program: TProgram,
    customDefinitions: readonly TCustomDefinition[],
  ) => Map<string, string>;
  workspaceToSource: (
    workspace: Blockly.Workspace,
    customDefinitions: readonly TCustomDefinition[],
  ) => BlockProgrammingSourceResult;
  nodeIdForBlock: (block: Blockly.Block) => string | undefined;
}
