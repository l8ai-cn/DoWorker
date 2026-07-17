use std::sync::Arc;

use agentsmesh_api_client::ApiClient;
use agentsmesh_types::proto_goalloop_v1 as lp;
use prost::Message;

pub struct GoalLoopService {
    client: Arc<ApiClient>,
}

impl GoalLoopService {
    pub fn new(client: Arc<ApiClient>) -> Self {
        Self { client }
    }

    pub async fn compile_loop_program_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::CompileLoopProgramRequest::decode(request)
            .map_err(|e| format!("decode compile_loop_program request: {e}"))?;
        let response = self
            .client
            .compile_loop_program_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(response.encode_to_vec())
    }

    pub async fn generate_loop_program_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::GenerateLoopProgramRequest::decode(request)
            .map_err(|e| format!("decode generate_loop_program request: {e}"))?;
        let response = self
            .client
            .generate_loop_program_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(response.encode_to_vec())
    }

    pub async fn repair_loop_program_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::RepairLoopProgramRequest::decode(request)
            .map_err(|e| format!("decode repair_loop_program request: {e}"))?;
        let response = self
            .client
            .repair_loop_program_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(response.encode_to_vec())
    }

    pub async fn run_loop_program_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::RunLoopProgramRequest::decode(request)
            .map_err(|e| format!("decode run_loop_program request: {e}"))?;
        let response = self
            .client
            .run_loop_program_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(response.encode_to_vec())
    }

    pub async fn list_worker_snapshots_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::ListWorkerSnapshotsRequest::decode(request)
            .map_err(|e| format!("decode list_worker_snapshots request: {e}"))?;
        let response = self
            .client
            .list_worker_snapshots_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(response.encode_to_vec())
    }

    pub async fn list_goal_loops_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::ListGoalLoopsRequest::decode(request)
            .map_err(|e| format!("decode list_goal_loops request: {e}"))?;
        let response = self
            .client
            .list_goal_loops_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(response.encode_to_vec())
    }

    pub async fn get_goal_loop_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::GetGoalLoopRequest::decode(request)
            .map_err(|e| format!("decode get_goal_loop request: {e}"))?;
        let response = self
            .client
            .get_goal_loop_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(response.encode_to_vec())
    }

    pub async fn create_goal_loop_connect(&self, request: &[u8]) -> Result<Vec<u8>, String> {
        let req = lp::CreateGoalLoopRequest::decode(request)
            .map_err(|e| format!("decode create_goal_loop request: {e}"))?;
        let response = self
            .client
            .create_goal_loop_connect(&req)
            .await
            .map_err(crate::wire)?;
        Ok(response.encode_to_vec())
    }

    pub async fn goal_loop_action_connect(
        &self,
        action: &str,
        request: &[u8],
    ) -> Result<Vec<u8>, String> {
        let req = lp::GoalLoopActionRequest::decode(request)
            .map_err(|e| format!("decode goal_loop_action request: {e}"))?;
        let response = match action {
            "start" => self.client.start_goal_loop_connect(&req).await,
            "verify" => self.client.verify_goal_loop_connect(&req).await,
            "cancel" => self.client.cancel_goal_loop_connect(&req).await,
            other => return Err(format!("unknown goal loop action: {other}")),
        }
        .map_err(crate::wire)?;
        Ok(response.encode_to_vec())
    }
}
