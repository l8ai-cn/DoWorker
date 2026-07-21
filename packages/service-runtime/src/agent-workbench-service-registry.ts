import type {
  WasmAgentWorkbenchService,
  WasmAgentWorkbenchState,
} from "agent-cloud-wasm";

export interface AgentWorkbenchServiceRegistry {
  agentWorkbenchService: WasmAgentWorkbenchService;
  agentWorkbenchState: WasmAgentWorkbenchState;
}
