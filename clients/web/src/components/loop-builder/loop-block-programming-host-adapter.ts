import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import type { BlockProgrammingHostAdapter } from "@/components/block-programming/block-programming-host-adapter";
import {
  createLoopBlockCatalog,
  registerLoopBlocks,
} from "./loop-block-catalog";
import {
  customBlockType,
  type LoopResolvedCustomBlockDefinition,
} from "./loop-custom-block-types";
import {
  nodeIdForBlock,
  projectProgramToWorkspace,
  workspaceToLoopSource,
} from "./loop-block-projection";
import type { LoopBlockCatalogMessages } from "./loop-workbench-messages";

export const loopBlockProgrammingHostAdapter:
  BlockProgrammingHostAdapter<LoopProgram, LoopBlockCatalogMessages, LoopResolvedCustomBlockDefinition> = {
    namespace: "loop",
    customBlockType,
    createCatalog: createLoopBlockCatalog,
    registerBlocks: registerLoopBlocks,
    projectProgram: projectProgramToWorkspace,
    workspaceToSource: workspaceToLoopSource,
    nodeIdForBlock,
  };
