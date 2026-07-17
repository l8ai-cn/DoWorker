import type {
  WasmAgentWorkbenchService,
  WasmAgentWorkbenchState,
} from "do-worker-wasm";

export interface AgentWorkbenchServiceRegistry {
  agentWorkbenchService: WasmAgentWorkbenchService;
  agentWorkbenchState: WasmAgentWorkbenchState;
}
