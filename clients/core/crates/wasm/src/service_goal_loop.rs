use std::sync::Arc;

use agentsmesh_api_client::ApiClient;
use agentsmesh_services::GoalLoopService;
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub struct WasmGoalLoopService(GoalLoopService);

#[wasm_bindgen]
impl WasmGoalLoopService {
    pub(crate) fn new(client: Arc<ApiClient>) -> Self {
        Self(GoalLoopService::new(client))
    }

    #[wasm_bindgen(js_name = compileLoopProgramConnect)]
    pub async fn compile_loop_program_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.compile_loop_program_connect(request).await
    }

    #[wasm_bindgen(js_name = runLoopProgramConnect)]
    pub async fn run_loop_program_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.run_loop_program_connect(request).await
    }

    #[wasm_bindgen(js_name = listWorkerSnapshotsConnect)]
    pub async fn list_worker_snapshots_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_worker_snapshots_connect(request).await
    }

    #[wasm_bindgen(js_name = listGoalLoopsConnect)]
    pub async fn list_goal_loops_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.list_goal_loops_connect(request).await
    }

    #[wasm_bindgen(js_name = getGoalLoopConnect)]
    pub async fn get_goal_loop_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.get_goal_loop_connect(request).await
    }

    #[wasm_bindgen(js_name = createGoalLoopConnect)]
    pub async fn create_goal_loop_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.create_goal_loop_connect(request).await
    }

    #[wasm_bindgen(js_name = startGoalLoopConnect)]
    pub async fn start_goal_loop_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.goal_loop_action_connect("start", request).await
    }

    #[wasm_bindgen(js_name = verifyGoalLoopConnect)]
    pub async fn verify_goal_loop_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.goal_loop_action_connect("verify", request).await
    }

    #[wasm_bindgen(js_name = cancelGoalLoopConnect)]
    pub async fn cancel_goal_loop_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        self.0.goal_loop_action_connect("cancel", request).await
    }
}
