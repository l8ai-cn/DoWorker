use agentcloud_state::workflow_state::{WorkflowData, WorkflowRunData};
use agentcloud_types::proto_workflow_v1::{
    Workflow as ProtoWorkflow, WorkflowRun as ProtoWorkflowRun,
};

pub(crate) fn workflow_from_proto(p: ProtoWorkflow) -> WorkflowData {
    WorkflowData {
        id: p.id,
        slug: p.slug,
        name: p.name,
        description: p.description,
        schedule: None,
        is_enabled: false,
        status: Some(p.status),
        agent_slug: Some(p.agent_slug),
        permission_mode: Some(p.permission_mode),
        prompt_template: Some(p.prompt_template),
        config_overrides: serde_json::from_str(&p.config_overrides_json).ok(),
        prompt_variables: serde_json::from_str(&p.prompt_variables_json).ok(),
        execution_mode: Some(p.execution_mode),
        autopilot_config: serde_json::from_str(&p.autopilot_config_json).ok(),
        sandbox_strategy: Some(p.sandbox_strategy),
        session_persistence: Some(p.session_persistence),
        concurrency_policy: Some(p.concurrency_policy),
        max_concurrent_runs: Some(p.max_concurrent_runs),
        max_retained_runs: Some(p.max_retained_runs),
        timeout_minutes: Some(p.timeout_minutes),
        idle_timeout_sec: Some(p.idle_timeout_sec),
        total_runs: Some(p.total_runs),
        successful_runs: Some(p.successful_runs),
        failed_runs: Some(p.failed_runs),
        active_run_count: Some(p.active_run_count),
        last_run_at: p.last_run_at,
        created_at: Some(p.created_at),
        updated_at: Some(p.updated_at),
        cron_expression: p.cron_expression,
        callback_url: p.callback_url,
        repository_id: p.repository_id,
        runner_id: p.runner_id,
        branch_name: p.branch_name,
        ticket_id: p.ticket_id,
        model_resource_id: p.model_resource_id,
        avg_duration_sec: p.avg_duration_sec,
        used_env_bundles: p.used_env_bundles,
    }
}

pub(crate) fn run_from_proto(p: ProtoWorkflowRun) -> WorkflowRunData {
    WorkflowRunData {
        id: p.id,
        workflow_slug: String::new(),
        run_number: Some(p.run_number),
        status: p.status,
        pod_key: p.pod_key,
        started_at: p.started_at,
        completed_at: p.completed_at,
        error_message: p.error_message,
        created_at: Some(p.created_at),
    }
}

fn json_str(v: &Option<serde_json::Value>) -> String {
    v.as_ref().map(|x| x.to_string()).unwrap_or_default()
}

pub(crate) fn workflow_to_proto(l: &WorkflowData) -> ProtoWorkflow {
    ProtoWorkflow {
        id: l.id,
        slug: l.slug.clone(),
        name: l.name.clone(),
        description: l.description.clone(),
        agent_slug: l.agent_slug.clone().unwrap_or_default(),
        permission_mode: l.permission_mode.clone().unwrap_or_default(),
        prompt_template: l.prompt_template.clone().unwrap_or_default(),
        config_overrides_json: json_str(&l.config_overrides),
        prompt_variables_json: json_str(&l.prompt_variables),
        execution_mode: l.execution_mode.clone().unwrap_or_default(),
        autopilot_config_json: json_str(&l.autopilot_config),
        status: l.status.clone().unwrap_or_default(),
        sandbox_strategy: l.sandbox_strategy.clone().unwrap_or_default(),
        session_persistence: l.session_persistence.unwrap_or_default(),
        concurrency_policy: l.concurrency_policy.clone().unwrap_or_default(),
        max_concurrent_runs: l.max_concurrent_runs.unwrap_or_default(),
        max_retained_runs: l.max_retained_runs.unwrap_or_default(),
        timeout_minutes: l.timeout_minutes.unwrap_or_default(),
        idle_timeout_sec: l.idle_timeout_sec.unwrap_or_default(),
        total_runs: l.total_runs.unwrap_or_default(),
        successful_runs: l.successful_runs.unwrap_or_default(),
        failed_runs: l.failed_runs.unwrap_or_default(),
        active_run_count: l.active_run_count.unwrap_or_default(),
        last_run_at: l.last_run_at.clone(),
        created_at: l.created_at.clone().unwrap_or_default(),
        updated_at: l.updated_at.clone().unwrap_or_default(),
        cron_expression: l.cron_expression.clone(),
        callback_url: l.callback_url.clone(),
        repository_id: l.repository_id,
        runner_id: l.runner_id,
        branch_name: l.branch_name.clone(),
        ticket_id: l.ticket_id,
        model_resource_id: l.model_resource_id,
        avg_duration_sec: l.avg_duration_sec,
        used_env_bundles: l.used_env_bundles.clone(),
        ..Default::default()
    }
}

pub(crate) fn run_to_proto(r: &WorkflowRunData) -> ProtoWorkflowRun {
    ProtoWorkflowRun {
        id: r.id,
        run_number: r.run_number.unwrap_or_default(),
        status: r.status.clone(),
        pod_key: r.pod_key.clone(),
        started_at: r.started_at.clone(),
        completed_at: r.completed_at.clone(),
        error_message: r.error_message.clone(),
        created_at: r.created_at.clone().unwrap_or_default(),
        ..Default::default()
    }
}
