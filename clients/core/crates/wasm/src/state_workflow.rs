use std::sync::Arc;

use agentsmesh_state::app_state::AppState;
use agentsmesh_state::workflow_state::WorkflowRunData;
use agentsmesh_types::proto_workflow_state_v1::{
    ClearCurrentWorkflowRequest, ClearWorkflowRunsRequest, InsertWorkflowRunRequest,
    PatchWorkflowFromActionRequest, PatchWorkflowRunStatusRequest, ReplaceCachedWorkflowsRequest,
    ReplaceCachedWorkflowRunsRequest, SetCurrentWorkflowRequest,
};
use agentsmesh_types::proto_workflow_v1::{
    ListWorkflowRunsResponse, ListWorkflowsResponse, Workflow as ProtoWorkflow,
};
use parking_lot::RwLock;
use prost::Message;
use wasm_bindgen::prelude::*;

use crate::state_workflow_proto::{workflow_from_proto, workflow_to_proto, run_from_proto, run_to_proto};

#[wasm_bindgen]
pub struct WasmWorkflowState {
    state: Arc<RwLock<AppState>>,
}

fn decode_err<E: std::fmt::Display>(e: E) -> JsValue {
    JsValue::from_str(&format!("decode: {e}"))
}

impl WasmWorkflowState {
    pub(crate) fn from_runtime(state: Arc<RwLock<AppState>>) -> Self {
        Self { state }
    }
}

#[wasm_bindgen]
impl WasmWorkflowState {
    #[wasm_bindgen(constructor)]
    pub fn new() -> Self {
        Self {
            state: Arc::new(RwLock::new(AppState::with_storage(
                crate::new_memory_backend(),
            ))),
        }
    }

    pub fn workflows_json(&self) -> String {
        serde_json::to_string(self.state.read().workflows.get_workflows()).unwrap_or_default()
    }

    pub fn current_workflow_json(&self) -> JsValue {
        match self.state.read().workflows.get_current_workflow() {
            Some(l) => JsValue::from_str(&serde_json::to_string(l).unwrap_or_default()),
            None => JsValue::NULL,
        }
    }

    pub fn runs_json(&self) -> String {
        serde_json::to_string(self.state.read().workflows.get_runs()).unwrap_or_default()
    }

    pub fn get_workflow_by_slug_json(&self, slug: &str) -> JsValue {
        match self.state.read().workflows.get_workflow_by_slug(slug) {
            Some(l) => JsValue::from_str(&serde_json::to_string(l).unwrap_or_default()),
            None => JsValue::NULL,
        }
    }

    // Read side (B, zero-JSON): prost-encode state into the same wrappers the
    // mutators decode, so the shared selectors decode bytes uniformly.
    pub fn workflows_bytes(&self) -> Vec<u8> {
        let workflows = self
            .state
            .read()
            .workflows
            .get_workflows()
            .iter()
            .map(workflow_to_proto)
            .collect();
        ReplaceCachedWorkflowsRequest { workflows }.encode_to_vec()
    }

    pub fn runs_bytes(&self) -> Vec<u8> {
        let runs = self
            .state
            .read()
            .workflows
            .get_runs()
            .iter()
            .map(run_to_proto)
            .collect();
        ReplaceCachedWorkflowRunsRequest { runs }.encode_to_vec()
    }

    pub fn current_workflow_bytes(&self) -> Vec<u8> {
        match self.state.read().workflows.get_current_workflow() {
            Some(l) => SetCurrentWorkflowRequest { workflow: Some(workflow_to_proto(l)) }.encode_to_vec(),
            None => Vec::new(),
        }
    }

    // Fetch→state (B): decode wire ListWorkflows/ListRuns response + fold into state
    // via workflow_from_proto/run_from_proto — no TS workflowToProtoWorkflow/fromProtoWorkflow.
    pub fn apply_fetched_workflows(&self, resp_bytes: &[u8]) -> Result<(), JsValue> {
        let resp = ListWorkflowsResponse::decode(resp_bytes).map_err(decode_err)?;
        let workflows = resp.items.into_iter().map(workflow_from_proto).collect();
        self.state.write().workflows.set_workflows(workflows);
        Ok(())
    }

    pub fn apply_fetched_runs(&self, resp_bytes: &[u8]) -> Result<(), JsValue> {
        let resp = ListWorkflowRunsResponse::decode(resp_bytes).map_err(decode_err)?;
        let runs = resp.items.into_iter().map(run_from_proto).collect();
        self.state.write().workflows.set_runs(runs);
        Ok(())
    }

    pub fn apply_appended_runs(&self, resp_bytes: &[u8]) -> Result<(), JsValue> {
        let resp = ListWorkflowRunsResponse::decode(resp_bytes).map_err(decode_err)?;
        let runs: Vec<WorkflowRunData> = resp.items.into_iter().map(run_from_proto).collect();
        self.state.write().workflows.append_runs(runs);
        Ok(())
    }

    pub fn set_current_workflow(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let req = SetCurrentWorkflowRequest::decode(req_bytes).map_err(decode_err)?;
        let workflow_data = req.workflow.map(workflow_from_proto);
        self.state.write().workflows.set_current_workflow(workflow_data);
        Ok(())
    }

    // Fetch→state (B): decode the full wire GetWorkflow response (Workflow) + fold via
    // workflow_from_proto — no TS workflowToProtoWorkflow round-trip (which dropped the
    // proto-only fields the lossy WorkflowData cannot carry).
    pub fn apply_fetched_current_workflow(&self, resp_bytes: &[u8]) -> Result<(), JsValue> {
        let proto = ProtoWorkflow::decode(resp_bytes).map_err(decode_err)?;
        self.state
            .write()
            .workflows
            .set_current_workflow(Some(workflow_from_proto(proto)));
        Ok(())
    }

    pub fn clear_current_workflow(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let _ = ClearCurrentWorkflowRequest::decode(req_bytes).map_err(decode_err)?;
        self.state.write().workflows.set_current_workflow(None);
        Ok(())
    }

    pub fn patch_workflow_from_action(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let req = PatchWorkflowFromActionRequest::decode(req_bytes).map_err(decode_err)?;
        let workflow_data = req
            .workflow
            .ok_or_else(|| JsValue::from_str("missing workflow"))?;
        self.state
            .write()
            .workflows
            .update_workflow(&req.slug, workflow_from_proto(workflow_data));
        Ok(())
    }

    pub fn insert_workflow_run(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let req = InsertWorkflowRunRequest::decode(req_bytes).map_err(decode_err)?;
        let run = req.run.ok_or_else(|| JsValue::from_str("missing run"))?;
        self.state.write().workflows.add_run(run_from_proto(run));
        Ok(())
    }

    pub fn patch_workflow_run_status(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let req = PatchWorkflowRunStatusRequest::decode(req_bytes).map_err(decode_err)?;
        self.state
            .write()
            .workflows
            .update_run_status(req.run_id, &req.status);
        Ok(())
    }

    pub fn clear_workflow_runs(&self, req_bytes: &[u8]) -> Result<(), JsValue> {
        let _ = ClearWorkflowRunsRequest::decode(req_bytes).map_err(decode_err)?;
        self.state.write().workflows.clear_runs();
        Ok(())
    }
}
