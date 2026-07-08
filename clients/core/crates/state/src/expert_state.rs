use crate::expert_types::Expert;

/// In-memory cache for the Expert domain (list + current + total). Mirrors the
/// `RepoState` list+current pattern; org-scoped, cleared on org switch. No
/// persistence backend — experts are cheap to refetch and not needed offline.
#[derive(Default)]
pub struct ExpertState {
    experts: Vec<Expert>,
    current_expert: Option<Expert>,
    total: i64,
}

impl ExpertState {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn experts(&self) -> &[Expert] {
        &self.experts
    }
    pub fn current_expert(&self) -> Option<&Expert> {
        self.current_expert.as_ref()
    }
    pub fn total(&self) -> i64 {
        self.total
    }

    pub fn set_experts(&mut self, experts: Vec<Expert>, total: i64) {
        tracing::debug!(target: "expert", count = experts.len(), total, "set experts (baseline)");
        self.experts = experts;
        self.total = total;
    }

    pub fn set_current_expert(&mut self, expert: Option<Expert>) {
        self.current_expert = expert;
    }

    pub fn remove_expert(&mut self, slug: &str) {
        tracing::info!(target: "expert", %slug, "remove expert");
        self.experts.retain(|e| e.slug != slug);
        if self.current_expert.as_ref().is_some_and(|e| e.slug == slug) {
            self.current_expert = None;
        }
    }
}
