use std::sync::Arc;

use agentsmesh_persistence::{StorageBackend, WorkflowRepo};

pub use crate::workflow_types::{workflow_run_status, WorkflowData, WorkflowRunData};

pub struct WorkflowState {
    workflows: Vec<WorkflowData>,
    current_workflow: Option<WorkflowData>,
    runs: Vec<WorkflowRunData>,
    repo: Option<WorkflowRepo<WorkflowData, WorkflowRunData>>,
}

impl WorkflowState {
    pub fn new() -> Self {
        Self {
            workflows: Vec::new(),
            current_workflow: None,
            runs: Vec::new(),
            repo: None,
        }
    }

    pub fn with_storage(backend: Arc<dyn StorageBackend>) -> Self {
        let repo = WorkflowRepo::new(backend);
        let workflows = repo.list_workflows().unwrap_or_default();
        Self {
            workflows,
            current_workflow: None,
            runs: Vec::new(),
            repo: Some(repo),
        }
    }

    pub fn get_workflows(&self) -> &[WorkflowData] {
        &self.workflows
    }
    pub fn get_current_workflow(&self) -> Option<&WorkflowData> {
        self.current_workflow.as_ref()
    }
    pub fn get_runs(&self) -> &[WorkflowRunData] {
        &self.runs
    }

    pub fn get_workflow_by_slug(&self, slug: &str) -> Option<&WorkflowData> {
        self.workflows.iter().find(|l| l.slug == slug)
    }

    pub fn set_workflows(&mut self, workflows: Vec<WorkflowData>) {
        tracing::debug!(target: "workflow", count = workflows.len(), "set workflows (baseline)");
        self.workflows = workflows;
        if let Some(repo) = &self.repo {
            for l in &self.workflows {
                let _ = repo.save_workflow(l);
            }
        }
    }

    pub fn set_current_workflow(&mut self, workflow_data: Option<WorkflowData>) {
        self.current_workflow = workflow_data;
    }

    pub fn update_workflow(&mut self, slug: &str, workflow_data: WorkflowData) {
        tracing::info!(target: "workflow", slug, status = ?workflow_data.status, "update workflow");
        if let Some(l) = self.workflows.iter_mut().find(|l| l.slug == slug) {
            *l = workflow_data.clone();
            if let Some(repo) = &self.repo {
                let _ = repo.save_workflow(l);
            }
        }
        if self
            .current_workflow
            .as_ref()
            .is_some_and(|l| l.slug == slug)
        {
            self.current_workflow = Some(workflow_data);
        }
    }

    pub fn add_run(&mut self, run: WorkflowRunData) {
        tracing::info!(target: "workflow", run_id = run.id, workflow_slug = %run.workflow_slug, status = %run.status, "add run");
        if let Some(repo) = &self.repo {
            let _ = repo.save_run(&run);
        }
        self.runs.push(run);
    }

    pub fn set_runs(&mut self, runs: Vec<WorkflowRunData>) {
        tracing::debug!(target: "workflow", count = runs.len(), "set runs (baseline)");
        self.runs = runs;
    }

    pub fn append_runs(&mut self, runs: Vec<WorkflowRunData>) {
        tracing::debug!(target: "workflow", count = runs.len(), "append runs");
        self.runs.extend(runs);
    }

    pub fn update_run_status(&mut self, run_id: i64, status: &str) {
        tracing::info!(target: "workflow", run_id, status, "run status changed");
        if let Some(run) = self.runs.iter_mut().find(|r| r.id == run_id) {
            run.status = status.to_string();
            if let Some(repo) = &self.repo {
                let _ = repo.save_run(run);
            }
        }
    }

    pub fn clear_runs(&mut self) {
        tracing::debug!(target: "workflow", "clear runs");
        self.runs.clear();
    }
}

impl Default for WorkflowState {
    fn default() -> Self {
        Self::new()
    }
}
