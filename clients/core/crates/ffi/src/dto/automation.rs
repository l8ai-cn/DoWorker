use agentsmesh_types::{
    ApproveAutopilotRequest, AutopilotController, AutopilotIteration, AutopilotListResponse,
    AutopilotStatus, CreateAutopilotRequest, CreateLoopRequest, LoopData, LoopListResponse,
    LoopRunData, LoopRunListResponse, LoopRunStatus, UpdateLoopRequest,
};

// ── Enums ─────────────────────────────────────────────────

#[derive(Clone, Copy, Debug, uniffi::Enum)]
pub enum AutopilotStatusDto {
    Idle,
    Running,
    Paused,
    WaitingApproval,
    Completed,
    Failed,
    Terminated,
    Unknown,
}

impl From<AutopilotStatus> for AutopilotStatusDto {
    fn from(s: AutopilotStatus) -> Self {
        match s {
            AutopilotStatus::Idle => Self::Idle,
            AutopilotStatus::Running => Self::Running,
            AutopilotStatus::Paused => Self::Paused,
            AutopilotStatus::WaitingApproval => Self::WaitingApproval,
            AutopilotStatus::Completed => Self::Completed,
            AutopilotStatus::Failed => Self::Failed,
            AutopilotStatus::Terminated => Self::Terminated,
            AutopilotStatus::Unknown => Self::Unknown,
        }
    }
}

#[derive(Clone, Copy, Debug, uniffi::Enum)]
pub enum LoopRunStatusDto {
    Pending,
    Running,
    Completed,
    Failed,
    Cancelled,
    Unknown,
}

impl From<LoopRunStatus> for LoopRunStatusDto {
    fn from(s: LoopRunStatus) -> Self {
        match s {
            LoopRunStatus::Pending => Self::Pending,
            LoopRunStatus::Running => Self::Running,
            LoopRunStatus::Completed => Self::Completed,
            LoopRunStatus::Failed => Self::Failed,
            LoopRunStatus::Cancelled => Self::Cancelled,
            LoopRunStatus::Unknown => Self::Unknown,
        }
    }
}

// ── Autopilot ─────────────────────────────────────────────

#[derive(Clone, Debug, uniffi::Record)]
pub struct AutopilotControllerDto {
    pub autopilot_controller_key: String,
    pub pod_key: String,
    pub status: Option<AutopilotStatusDto>,
    pub phase: Option<String>,
    pub prompt: Option<String>,
    pub max_iterations: Option<i64>,
    pub iteration_timeout_sec: Option<i64>,
    pub no_progress_threshold: Option<i64>,
    pub same_error_threshold: Option<i64>,
    pub approval_timeout_min: Option<i64>,
    pub current_iteration: Option<i64>,
    pub control_agent_slug: Option<String>,
    pub circuit_breaker_state: Option<String>,
    pub circuit_breaker_reason: Option<String>,
    pub created_at: Option<String>,
    pub updated_at: Option<String>,
}

impl From<AutopilotController> for AutopilotControllerDto {
    fn from(c: AutopilotController) -> Self {
        Self {
            autopilot_controller_key: c.autopilot_controller_key,
            pod_key: c.pod_key,
            status: c.status.map(Into::into),
            phase: c.phase,
            prompt: c.prompt,
            max_iterations: c.max_iterations,
            iteration_timeout_sec: c.iteration_timeout_sec,
            no_progress_threshold: c.no_progress_threshold,
            same_error_threshold: c.same_error_threshold,
            approval_timeout_min: c.approval_timeout_min,
            current_iteration: c.current_iteration,
            control_agent_slug: c.control_agent_slug,
            circuit_breaker_state: c.circuit_breaker_state,
            circuit_breaker_reason: c.circuit_breaker_reason,
            created_at: c.created_at,
            updated_at: c.updated_at,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct AutopilotListResponseDto {
    pub controllers: Vec<AutopilotControllerDto>,
}

impl From<AutopilotListResponse> for AutopilotListResponseDto {
    fn from(r: AutopilotListResponse) -> Self {
        Self {
            controllers: r
                .controllers
                .into_iter()
                .map(AutopilotControllerDto::from)
                .collect(),
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct AutopilotIterationDto {
    pub id: i64,
    pub controller_key: String,
    pub iteration_number: Option<i64>,
    pub status: Option<String>,
    pub result: Option<String>,
    pub started_at: Option<String>,
    pub completed_at: Option<String>,
}

impl From<AutopilotIteration> for AutopilotIterationDto {
    fn from(i: AutopilotIteration) -> Self {
        Self {
            id: i.id,
            controller_key: i.controller_key,
            iteration_number: i.iteration_number,
            status: i.status,
            result: i.result,
            started_at: i.started_at,
            completed_at: i.completed_at,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct AutopilotIterationListResponseDto {
    pub iterations: Vec<AutopilotIterationDto>,
}

impl From<Vec<AutopilotIteration>> for AutopilotIterationListResponseDto {
    fn from(iterations: Vec<AutopilotIteration>) -> Self {
        Self {
            iterations: iterations.into_iter().map(AutopilotIterationDto::from).collect(),
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct CreateAutopilotRequestDto {
    pub pod_key: String,
    pub prompt: Option<String>,
    pub max_iterations: Option<i64>,
    pub iteration_timeout_sec: Option<i64>,
    pub no_progress_threshold: Option<i64>,
    pub same_error_threshold: Option<i64>,
    pub approval_timeout_min: Option<i64>,
    pub control_agent_slug: Option<String>,
    pub control_prompt_template: Option<String>,
    pub mcp_config_json: Option<String>,
}

impl From<CreateAutopilotRequestDto> for CreateAutopilotRequest {
    fn from(d: CreateAutopilotRequestDto) -> Self {
        Self {
            pod_key: d.pod_key,
            prompt: d.prompt,
            max_iterations: d.max_iterations,
            iteration_timeout_sec: d.iteration_timeout_sec,
            no_progress_threshold: d.no_progress_threshold,
            same_error_threshold: d.same_error_threshold,
            approval_timeout_min: d.approval_timeout_min,
            control_agent_slug: d.control_agent_slug,
            control_prompt_template: d.control_prompt_template,
            mcp_config_json: d.mcp_config_json,
        }
    }
}

pub(crate) fn approve_autopilot_req(
    continue_execution: Option<bool>,
    additional_iterations: Option<i64>,
) -> ApproveAutopilotRequest {
    ApproveAutopilotRequest {
        continue_execution,
        additional_iterations,
    }
}

// ── Loop ──────────────────────────────────────────────────

fn value_opt_to_json(v: Option<serde_json::Value>) -> Option<String> {
    v.and_then(|x| serde_json::to_string(&x).ok())
}

fn json_opt_to_value(s: Option<String>) -> Option<serde_json::Value> {
    s.and_then(|x| serde_json::from_str(&x).ok())
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct LoopDataDto {
    pub id: i64,
    pub slug: String,
    pub name: String,
    pub description: Option<String>,
    pub schedule: Option<String>,
    pub is_enabled: bool,
    pub status: Option<String>,
    pub agent_slug: Option<String>,
    pub permission_mode: Option<String>,
    pub prompt_template: Option<String>,
    /// Opaque JSON — per-agent config overrides.
    pub config_overrides_json: Option<String>,
    /// Opaque JSON — prompt template variable bindings.
    pub prompt_variables_json: Option<String>,
    pub execution_mode: Option<String>,
    /// Opaque JSON — nested autopilot controller config.
    pub autopilot_config_json: Option<String>,
    pub sandbox_strategy: Option<String>,
    pub session_persistence: Option<bool>,
    pub concurrency_policy: Option<String>,
    pub max_concurrent_runs: Option<i32>,
    pub max_retained_runs: Option<i32>,
    pub timeout_minutes: Option<i32>,
    pub idle_timeout_sec: Option<i32>,
    pub total_runs: Option<i64>,
    pub successful_runs: Option<i64>,
    pub failed_runs: Option<i64>,
    pub active_run_count: Option<i64>,
    pub last_run_at: Option<String>,
    pub created_at: Option<String>,
    pub updated_at: Option<String>,
}

impl From<LoopData> for LoopDataDto {
    fn from(l: LoopData) -> Self {
        Self {
            id: l.id,
            slug: l.slug,
            name: l.name,
            description: l.description,
            schedule: l.schedule,
            is_enabled: l.is_enabled,
            status: l.status,
            agent_slug: l.agent_slug,
            permission_mode: l.permission_mode,
            prompt_template: l.prompt_template,
            config_overrides_json: value_opt_to_json(l.config_overrides),
            prompt_variables_json: value_opt_to_json(l.prompt_variables),
            execution_mode: l.execution_mode,
            autopilot_config_json: value_opt_to_json(l.autopilot_config),
            sandbox_strategy: l.sandbox_strategy,
            session_persistence: l.session_persistence,
            concurrency_policy: l.concurrency_policy,
            max_concurrent_runs: l.max_concurrent_runs,
            max_retained_runs: l.max_retained_runs,
            timeout_minutes: l.timeout_minutes,
            idle_timeout_sec: l.idle_timeout_sec,
            total_runs: l.total_runs,
            successful_runs: l.successful_runs,
            failed_runs: l.failed_runs,
            active_run_count: l.active_run_count,
            last_run_at: l.last_run_at,
            created_at: l.created_at,
            updated_at: l.updated_at,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct LoopListResponseDto {
    pub loops: Vec<LoopDataDto>,
    pub total: Option<i64>,
}

impl From<LoopListResponse> for LoopListResponseDto {
    fn from(r: LoopListResponse) -> Self {
        Self {
            loops: r.loops.into_iter().map(LoopDataDto::from).collect(),
            total: r.total,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct LoopRunDataDto {
    pub id: i64,
    pub loop_slug: String,
    pub run_number: Option<i64>,
    pub status: LoopRunStatusDto,
    pub pod_key: Option<String>,
    pub started_at: Option<String>,
    pub completed_at: Option<String>,
    pub error_message: Option<String>,
    pub created_at: Option<String>,
}

impl From<LoopRunData> for LoopRunDataDto {
    fn from(r: LoopRunData) -> Self {
        Self {
            id: r.id,
            loop_slug: r.loop_slug,
            run_number: r.run_number,
            status: r.status.into(),
            pod_key: r.pod_key,
            started_at: r.started_at,
            completed_at: r.completed_at,
            error_message: r.error_message,
            created_at: r.created_at,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct LoopRunListResponseDto {
    pub runs: Vec<LoopRunDataDto>,
    pub total: Option<i64>,
    pub limit: Option<i64>,
    pub offset: Option<i64>,
}

impl From<LoopRunListResponse> for LoopRunListResponseDto {
    fn from(r: LoopRunListResponse) -> Self {
        Self {
            runs: r.runs.into_iter().map(LoopRunDataDto::from).collect(),
            total: r.total,
            limit: r.limit,
            offset: r.offset,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct CreateLoopRequestDto {
    pub name: String,
    pub slug: Option<String>,
    pub description: Option<String>,
    pub agent_slug: Option<String>,
    pub custom_agent_slug: Option<String>,
    pub permission_mode: Option<String>,
    pub prompt_template: Option<String>,
    pub prompt_variables_json: Option<String>,
    pub repository_id: Option<i64>,
    pub runner_id: Option<i64>,
    pub branch_name: Option<String>,
    pub ticket_id: Option<String>,
    pub credential_profile_id: Option<i64>,
    pub config_overrides_json: Option<String>,
    pub execution_mode: Option<String>,
    pub cron_expression: Option<String>,
    pub autopilot_config_json: Option<String>,
    pub callback_url: Option<String>,
    pub sandbox_strategy: Option<String>,
    pub session_persistence: Option<String>,
    pub concurrency_policy: Option<String>,
    pub max_concurrent_runs: Option<i64>,
    pub max_retained_runs: Option<i64>,
    pub timeout_minutes: Option<i64>,
}

impl From<CreateLoopRequestDto> for CreateLoopRequest {
    fn from(d: CreateLoopRequestDto) -> Self {
        Self {
            name: d.name,
            slug: d.slug,
            description: d.description,
            agent_slug: d.agent_slug,
            custom_agent_slug: d.custom_agent_slug,
            permission_mode: d.permission_mode,
            prompt_template: d.prompt_template,
            prompt_variables: json_opt_to_value(d.prompt_variables_json),
            repository_id: d.repository_id,
            runner_id: d.runner_id,
            branch_name: d.branch_name,
            ticket_id: d.ticket_id,
            credential_profile_id: d.credential_profile_id,
            config_overrides: json_opt_to_value(d.config_overrides_json),
            execution_mode: d.execution_mode,
            cron_expression: d.cron_expression,
            autopilot_config: json_opt_to_value(d.autopilot_config_json),
            callback_url: d.callback_url,
            sandbox_strategy: d.sandbox_strategy,
            session_persistence: d.session_persistence,
            concurrency_policy: d.concurrency_policy,
            max_concurrent_runs: d.max_concurrent_runs,
            max_retained_runs: d.max_retained_runs,
            timeout_minutes: d.timeout_minutes,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct UpdateLoopRequestDto {
    pub name: Option<String>,
    pub description: Option<String>,
    pub agent_slug: Option<String>,
    pub prompt_template: Option<String>,
    pub prompt_variables_json: Option<String>,
    pub repository_id: Option<i64>,
    pub runner_id: Option<i64>,
    pub branch_name: Option<String>,
    pub cron_expression: Option<String>,
    pub autopilot_config_json: Option<String>,
    pub sandbox_strategy: Option<String>,
    pub session_persistence: Option<String>,
    pub concurrency_policy: Option<String>,
    pub max_concurrent_runs: Option<i64>,
    pub max_retained_runs: Option<i64>,
    pub timeout_minutes: Option<i64>,
}

impl From<UpdateLoopRequestDto> for UpdateLoopRequest {
    fn from(d: UpdateLoopRequestDto) -> Self {
        Self {
            name: d.name,
            description: d.description,
            agent_slug: d.agent_slug,
            prompt_template: d.prompt_template,
            prompt_variables: json_opt_to_value(d.prompt_variables_json),
            repository_id: d.repository_id,
            runner_id: d.runner_id,
            branch_name: d.branch_name,
            cron_expression: d.cron_expression,
            autopilot_config: json_opt_to_value(d.autopilot_config_json),
            sandbox_strategy: d.sandbox_strategy,
            session_persistence: d.session_persistence,
            concurrency_policy: d.concurrency_policy,
            max_concurrent_runs: d.max_concurrent_runs,
            max_retained_runs: d.max_retained_runs,
            timeout_minutes: d.timeout_minutes,
        }
    }
}
