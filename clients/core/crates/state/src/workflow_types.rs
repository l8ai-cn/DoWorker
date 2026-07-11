use agentsmesh_persistence::{WorkflowRow, WorkflowRunRow};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct WorkflowData {
    #[serde(default)]
    pub id: i64,
    pub slug: String,
    pub name: String,
    #[serde(default)]
    pub description: Option<String>,
    #[serde(default)]
    pub schedule: Option<String>,
    #[serde(default)]
    pub is_enabled: bool,
    #[serde(default)]
    pub status: Option<String>,
    #[serde(default)]
    pub agent_slug: Option<String>,
    #[serde(default)]
    pub permission_mode: Option<String>,
    #[serde(default)]
    pub prompt_template: Option<String>,
    #[serde(default)]
    pub config_overrides: Option<serde_json::Value>,
    #[serde(default)]
    pub prompt_variables: Option<serde_json::Value>,
    #[serde(default)]
    pub execution_mode: Option<String>,
    #[serde(default)]
    pub autopilot_config: Option<serde_json::Value>,
    #[serde(default)]
    pub sandbox_strategy: Option<String>,
    #[serde(default)]
    pub session_persistence: Option<bool>,
    #[serde(default)]
    pub concurrency_policy: Option<String>,
    #[serde(default)]
    pub max_concurrent_runs: Option<i32>,
    #[serde(default)]
    pub max_retained_runs: Option<i32>,
    #[serde(default)]
    pub timeout_minutes: Option<i32>,
    #[serde(default)]
    pub idle_timeout_sec: Option<i32>,
    #[serde(default)]
    pub total_runs: Option<i64>,
    #[serde(default)]
    pub successful_runs: Option<i64>,
    #[serde(default)]
    pub failed_runs: Option<i64>,
    #[serde(default)]
    pub active_run_count: Option<i64>,
    #[serde(default)]
    pub last_run_at: Option<String>,
    #[serde(default)]
    pub created_at: Option<String>,
    #[serde(default)]
    pub updated_at: Option<String>,
    #[serde(default)]
    pub cron_expression: Option<String>,
    #[serde(default)]
    pub callback_url: Option<String>,
    #[serde(default)]
    pub repository_id: Option<i64>,
    #[serde(default)]
    pub runner_id: Option<i64>,
    #[serde(default)]
    pub branch_name: Option<String>,
    #[serde(default)]
    pub ticket_id: Option<i64>,
    #[serde(default)]
    pub model_resource_id: Option<i64>,
    #[serde(default)]
    pub avg_duration_sec: Option<f64>,
    #[serde(default)]
    pub used_env_bundles: Vec<String>,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct WorkflowRunData {
    pub id: i64,
    #[serde(default)]
    pub workflow_slug: String,
    #[serde(default)]
    pub run_number: Option<i64>,
    #[serde(default)]
    pub status: String,
    #[serde(default)]
    pub pod_key: Option<String>,
    #[serde(default)]
    pub started_at: Option<String>,
    #[serde(default)]
    pub completed_at: Option<String>,
    #[serde(default)]
    pub error_message: Option<String>,
    #[serde(default)]
    pub created_at: Option<String>,
}

pub mod workflow_run_status {
    pub const PENDING: &str = "pending";
    pub const RUNNING: &str = "running";
    pub const COMPLETED: &str = "completed";
    pub const FAILED: &str = "failed";
    pub const CANCELLED: &str = "cancelled";
}

impl WorkflowRow for WorkflowData {
    fn slug(&self) -> &str {
        &self.slug
    }
}

impl WorkflowRunRow for WorkflowRunData {
    fn id(&self) -> i64 {
        self.id
    }
    fn workflow_slug(&self) -> &str {
        &self.workflow_slug
    }
}
