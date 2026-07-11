// Facade re-export of the workflow Connect-RPC adapter. Business code imports
// from here (or from the `@/lib/api` barrel) so the wire-shape layer stays
// internal to the facade boundary. Tests mock this path.

export {
  listWorkflows,
  listWorkflowsRaw,
  getWorkflow,
  getWorkflowRaw,
  createWorkflow,
  updateWorkflow,
  deleteWorkflow,
  enableWorkflow,
  disableWorkflow,
  triggerWorkflow,
  listWorkflowRuns,
  listWorkflowRunsRaw,
  cancelWorkflowRun,
  type TriggerWorkflowResult,
} from "../connect/workflowConnect";
