use std::marker::PhantomData;
use std::sync::Arc;

use serde::Serialize;
use serde::de::DeserializeOwned;

use crate::backend::StorageBackend;
use crate::error::Result;

pub trait WorkflowRow: Serialize + DeserializeOwned {
    fn slug(&self) -> &str;
}

pub trait WorkflowRunRow: Serialize + DeserializeOwned {
    fn id(&self) -> i64;
    fn workflow_slug(&self) -> &str;
}

pub struct WorkflowRepo<L: WorkflowRow, R: WorkflowRunRow> {
    backend: Arc<dyn StorageBackend>,
    _phantom: PhantomData<fn(L, R)>,
}

impl<L: WorkflowRow, R: WorkflowRunRow> WorkflowRepo<L, R> {
    pub fn new(backend: Arc<dyn StorageBackend>) -> Self {
        Self { backend, _phantom: PhantomData }
    }

    pub fn save_workflow(&self, data: &L) -> Result<()> {
        let bytes = serde_json::to_vec(data)?;
        self.backend.put_raw("workflows", data.slug(), &[], &bytes)
    }

    pub fn get_workflow(&self, slug: &str) -> Result<Option<L>> {
        match self.backend.get_raw("workflows", slug)? {
            Some(data) => Ok(Some(serde_json::from_slice(&data)?)),
            None => Ok(None),
        }
    }

    pub fn delete_workflow(&self, slug: &str) -> Result<()> {
        self.backend.delete_raw("workflows", slug)
    }

    pub fn list_workflows(&self) -> Result<Vec<L>> {
        super::deserialize_rows(self.backend.list_raw("workflows")?)
    }

    pub fn save_run(&self, run: &R) -> Result<()> {
        let data = serde_json::to_vec(run)?;
        let fields: &[(&str, &str)] = &[("workflow_slug", run.workflow_slug())];
        self.backend
            .put_raw("workflow_runs", &run.id().to_string(), fields, &data)
    }

    pub fn get_run(&self, id: i64) -> Result<Option<R>> {
        match self.backend.get_raw("workflow_runs", &id.to_string())? {
            Some(data) => Ok(Some(serde_json::from_slice(&data)?)),
            None => Ok(None),
        }
    }

    pub fn get_runs_by_workflow(&self, workflow_slug: &str) -> Result<Vec<R>> {
        let rows = self
            .backend
            .query_raw("workflow_runs", "workflow_slug", workflow_slug)?;
        let mut runs: Vec<R> = super::deserialize_rows(rows)?;
        runs.sort_by(|a, b| b.id().cmp(&a.id()));
        Ok(runs)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::backend::InMemoryBackend;
    use serde::{Deserialize, Serialize};

    #[derive(Debug, Clone, Default, Serialize, Deserialize)]
    struct TestWorkflow {
        slug: String,
        name: String,
    }
    impl WorkflowRow for TestWorkflow {
        fn slug(&self) -> &str { &self.slug }
    }

    #[derive(Debug, Clone, Default, Serialize, Deserialize)]
    struct TestRun {
        id: i64,
        workflow_slug: String,
    }
    impl WorkflowRunRow for TestRun {
        fn id(&self) -> i64 { self.id }
        fn workflow_slug(&self) -> &str { &self.workflow_slug }
    }

    fn make_repo() -> WorkflowRepo<TestWorkflow, TestRun> {
        WorkflowRepo::new(Arc::new(InMemoryBackend::new()))
    }

    #[test]
    fn workflow_crud() {
        let repo = make_repo();
        let ld = TestWorkflow { slug: "workflow-1".into(), name: "Hourly".into() };
        repo.save_workflow(&ld).unwrap();
        let loaded = repo.get_workflow("workflow-1").unwrap().unwrap();
        assert_eq!(loaded.name, "Hourly");
        repo.delete_workflow("workflow-1").unwrap();
        assert!(repo.get_workflow("workflow-1").unwrap().is_none());
    }

    #[test]
    fn list_workflows() {
        let repo = make_repo();
        let ld = TestWorkflow { slug: "l".into(), name: "n".into() };
        repo.save_workflow(&ld).unwrap();
        assert_eq!(repo.list_workflows().unwrap().len(), 1);
    }

    #[test]
    fn runs_by_workflow_sorted_desc() {
        let repo = make_repo();
        for i in 1..=3 {
            repo.save_run(&TestRun { id: i, workflow_slug: "workflow-1".into() }).unwrap();
        }
        let runs = repo.get_runs_by_workflow("workflow-1").unwrap();
        assert_eq!(runs.len(), 3);
        assert!(runs[0].id > runs[1].id);
    }

    #[test]
    fn runs_filtered_by_workflow_slug() {
        let repo = make_repo();
        repo.save_run(&TestRun { id: 1, workflow_slug: "workflow-1".into() }).unwrap();
        repo.save_run(&TestRun { id: 2, workflow_slug: "workflow-2".into() }).unwrap();
        assert_eq!(repo.get_runs_by_workflow("workflow-1").unwrap().len(), 1);
        assert_eq!(repo.get_runs_by_workflow("workflow-2").unwrap().len(), 1);
    }

    #[test]
    fn get_workflow_nonexistent() {
        let repo = make_repo();
        assert!(repo.get_workflow("nope").unwrap().is_none());
    }

    #[test]
    fn get_run_nonexistent() {
        let repo = make_repo();
        assert!(repo.get_run(999).unwrap().is_none());
    }

    #[test]
    fn get_run_roundtrip() {
        let repo = make_repo();
        repo.save_run(&TestRun { id: 42, workflow_slug: "workflow-x".into() }).unwrap();
        let loaded = repo.get_run(42).unwrap().unwrap();
        assert_eq!(loaded.workflow_slug, "workflow-x");
    }

    #[test]
    fn delete_workflow_nonexistent_is_noop() {
        let repo = make_repo();
        repo.delete_workflow("nope").unwrap();
    }

    #[test]
    fn save_workflow_overwrites() {
        let repo = make_repo();
        let mut ld = TestWorkflow { slug: "l1".into(), name: "A".into() };
        repo.save_workflow(&ld).unwrap();
        ld.name = "B".into();
        repo.save_workflow(&ld).unwrap();
        let loaded = repo.get_workflow("l1").unwrap().unwrap();
        assert_eq!(loaded.name, "B");
        assert_eq!(repo.list_workflows().unwrap().len(), 1);
    }

    #[test]
    fn runs_empty_for_unknown_workflow() {
        let repo = make_repo();
        assert!(repo.get_runs_by_workflow("unknown").unwrap().is_empty());
    }
}
