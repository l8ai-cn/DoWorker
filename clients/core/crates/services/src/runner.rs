use std::sync::Arc;
use std::sync::RwLock;

use agentsmesh_api_client::ApiClient;
use agentsmesh_state::runner_state::RunnerState;
use agentsmesh_types::{
    Runner, RunnerStatus, UpdateRunnerRequest, CreateRunnerTokenRequest,
};

use crate::parse_status;

pub struct RunnerService {
    client: Arc<ApiClient>,
    state: RwLock<RunnerState>,
}

impl RunnerService {
    pub fn new(client: Arc<ApiClient>, state: RunnerState) -> Self {
        Self { client, state: RwLock::new(state) }
    }

    pub fn runners_json(&self) -> String {
        serde_json::to_string(self.state.read().unwrap().runners()).unwrap_or_default()
    }

    pub fn available_runners_json(&self) -> String {
        serde_json::to_string(self.state.read().unwrap().available_runners()).unwrap_or_default()
    }

    pub fn current_runner_json(&self) -> Option<String> {
        self.state.read().unwrap().current_runner()
            .map(|r| serde_json::to_string(r).unwrap_or_default())
    }

    pub fn get_runner_json(&self, id: i64) -> Option<String> {
        self.state.read().unwrap().get_runner(id)
            .map(|r| serde_json::to_string(r).unwrap_or_default())
    }

    pub fn set_runners(&self, json: &str) {
        if let Ok(v) = serde_json::from_str::<Vec<Runner>>(json) {
            self.state.write().unwrap().set_runners(v);
        }
    }

    pub fn set_available_runners(&self, json: &str) {
        if let Ok(v) = serde_json::from_str::<Vec<Runner>>(json) {
            self.state.write().unwrap().set_available_runners(v);
        }
    }

    pub fn set_current_runner(&self, json: &str) {
        let r = if json.is_empty() { None } else { serde_json::from_str::<Runner>(json).ok() };
        self.state.write().unwrap().set_current_runner(r);
    }

    pub fn update_runner_local(&self, id: f64, json: &str) {
        if let Ok(r) = serde_json::from_str::<Runner>(json) {
            self.state.write().unwrap().update_runner(id as i64, r);
        }
    }

    pub fn update_runner_status(&self, id: i64, status: &str) {
        let parsed = parse_status::<RunnerStatus>(status);
        self.state.write().unwrap().update_runner_status(id, parsed);
    }

    pub fn remove_runner_local(&self, id: i64) {
        self.state.write().unwrap().remove_runner(id);
    }

    pub async fn fetch_runners(&self, status: Option<String>) -> Result<String, String> {
        let resp = self.client
            .list_runners(status.as_deref())
            .await.map_err(crate::wire)?;
        self.state.write().unwrap().set_runners(resp.runners.clone());
        serde_json::to_string(&resp).map_err(crate::wire)
    }

    pub async fn fetch_available_runners(&self) -> Result<String, String> {
        let resp = self.client
            .list_available_runners()
            .await.map_err(crate::wire)?;
        self.state.write().unwrap().set_available_runners(resp.runners.clone());
        serde_json::to_string(&resp).map_err(crate::wire)
    }

    pub async fn fetch_runner(&self, id: i64) -> Result<String, String> {
        let resp = self.client
            .get_runner(id)
            .await.map_err(crate::wire)?;
        self.state.write().unwrap().set_current_runner(Some(resp.runner.clone()));
        serde_json::to_string(&resp).map_err(crate::wire)
    }

    pub async fn update_runner(&self, id: i64, request_json: &str) -> Result<String, String> {
        let req: UpdateRunnerRequest = serde_json::from_str(request_json)
            .map_err(crate::wire)?;
        let runner: Runner = self.client
            .update_runner(id, &req)
            .await.map_err(crate::wire)?;
        self.state.write().unwrap().update_runner(id, runner.clone());
        serde_json::to_string(&runner).map_err(crate::wire)
    }

    pub async fn delete_runner(&self, id: i64) -> Result<(), String> {
        self.client.delete_runner(id).await.map_err(crate::wire)?;
        self.state.write().unwrap().remove_runner(id);
        Ok(())
    }

    pub async fn create_token(&self, request_json: &str) -> Result<String, String> {
        let req: CreateRunnerTokenRequest = serde_json::from_str(request_json)
            .map_err(crate::wire)?;
        let token = self.client
            .create_runner_token(&req)
            .await.map_err(crate::wire)?;
        serde_json::to_string(&token).map_err(crate::wire)
    }

    pub async fn fetch_tokens(&self) -> Result<String, String> {
        let resp = self.client
            .list_runner_tokens()
            .await.map_err(crate::wire)?;
        serde_json::to_string(&resp).map_err(crate::wire)
    }

    pub async fn delete_token(&self, id: i64) -> Result<(), String> {
        self.client.delete_runner_token(id).await.map_err(crate::wire)?;
        Ok(())
    }

    pub async fn list_runner_logs(&self, id: i64) -> Result<String, String> {
        let resp = self.client
            .list_runner_logs(id)
            .await.map_err(crate::wire)?;
        serde_json::to_string(&resp).map_err(crate::wire)
    }

    pub async fn request_log_upload(&self, id: i64) -> Result<(), String> {
        self.client
            .request_runner_log_upload(id)
            .await.map_err(crate::wire)?;
        Ok(())
    }

    pub async fn upgrade_runner(&self, id: i64, request_json: &str) -> Result<String, String> {
        let req: agentsmesh_types::UpgradeRunnerRequest = serde_json::from_str(request_json)
            .map_err(crate::wire)?;
        let resp = self.client
            .upgrade_runner(id, &req)
            .await.map_err(crate::wire)?;
        serde_json::to_string(&resp).map_err(crate::wire)
    }

    pub async fn list_runner_pods(
        &self, id: i64, status: Option<String>, limit: Option<u32>, offset: Option<u32>,
    ) -> Result<String, String> {
        let resp = self.client
            .list_runner_pods(id, status.as_deref(), limit, offset)
            .await.map_err(crate::wire)?;
        serde_json::to_string(&resp).map_err(crate::wire)
    }

    pub async fn query_runner_sandboxes(&self, id: i64, request_json: &str) -> Result<String, String> {
        let req: agentsmesh_types::SandboxQueryRequest = serde_json::from_str(request_json)
            .map_err(crate::wire)?;
        let resp = self.client
            .query_runner_sandboxes(id, &req)
            .await.map_err(crate::wire)?;
        serde_json::to_string(&resp).map_err(crate::wire)
    }

    pub async fn get_auth_status(&self, auth_key: &str) -> Result<String, String> {
        let resp = self.client
            .get_runner_auth_status(auth_key)
            .await.map_err(crate::wire)?;
        serde_json::to_string(&resp).map_err(crate::wire)
    }

    pub async fn authorize_runner(&self, request_json: &str) -> Result<String, String> {
        let req: agentsmesh_types::AuthorizeRunnerRequest = serde_json::from_str(request_json)
            .map_err(crate::wire)?;
        let resp = self.client
            .authorize_runner(&req)
            .await.map_err(crate::wire)?;
        serde_json::to_string(&resp).map_err(crate::wire)
    }
}
