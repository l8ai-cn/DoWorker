use crate::expert_state::ExpertState;
use crate::expert_types::{Expert, ExpertListResponse};

fn make_expert(id: i64, slug: &str) -> Expert {
    Expert {
        id,
        slug: slug.to_string(),
        name: slug.to_uppercase(),
        ..Default::default()
    }
}

#[test]
fn new_is_empty() {
    let s = ExpertState::new();
    assert!(s.experts().is_empty());
    assert!(s.current_expert().is_none());
    assert_eq!(s.total(), 0);
}

#[test]
fn set_and_get_experts() {
    let mut s = ExpertState::new();
    s.set_experts(vec![make_expert(1, "alpha"), make_expert(2, "beta")], 2);
    assert_eq!(s.experts().len(), 2);
    assert_eq!(s.total(), 2);
    assert_eq!(s.experts()[0].slug, "alpha");
}

#[test]
fn set_experts_replaces_previous() {
    let mut s = ExpertState::new();
    s.set_experts(vec![make_expert(1, "alpha")], 1);
    s.set_experts(vec![make_expert(2, "beta"), make_expert(3, "gamma")], 5);
    assert_eq!(s.experts().len(), 2);
    assert_eq!(s.total(), 5);
    assert_eq!(s.experts()[0].id, 2);
}

#[test]
fn set_current_expert() {
    let mut s = ExpertState::new();
    s.set_current_expert(Some(make_expert(1, "alpha")));
    assert_eq!(s.current_expert().unwrap().slug, "alpha");
    s.set_current_expert(None);
    assert!(s.current_expert().is_none());
}

#[test]
fn remove_expert_drops_from_list() {
    let mut s = ExpertState::new();
    s.set_experts(vec![make_expert(1, "alpha"), make_expert(2, "beta")], 2);
    s.remove_expert("alpha");
    assert_eq!(s.experts().len(), 1);
    assert_eq!(s.experts()[0].slug, "beta");
}

#[test]
fn remove_expert_clears_current_if_same() {
    let mut s = ExpertState::new();
    s.set_experts(vec![make_expert(1, "alpha")], 1);
    s.set_current_expert(Some(make_expert(1, "alpha")));
    s.remove_expert("alpha");
    assert!(s.current_expert().is_none());
}

#[test]
fn remove_expert_keeps_current_if_different() {
    let mut s = ExpertState::new();
    s.set_experts(vec![make_expert(1, "alpha"), make_expert(2, "beta")], 2);
    s.set_current_expert(Some(make_expert(2, "beta")));
    s.remove_expert("alpha");
    assert_eq!(s.current_expert().unwrap().slug, "beta");
}

#[test]
fn remove_nonexistent_is_noop() {
    let mut s = ExpertState::new();
    s.set_experts(vec![make_expert(1, "alpha")], 1);
    s.remove_expert("ghost");
    assert_eq!(s.experts().len(), 1);
}

#[test]
fn list_response_deserializes_backend_shape() {
    let raw = r#"{
        "experts": [{
            "id": 7, "slug": "reviewer", "name": "Reviewer",
            "agent_slug": "claude-code", "interaction_mode": "pty",
            "perpetual": false, "used_env_bundles": [], "skill_slugs": ["merge"],
            "knowledge_mounts": [{"slug": "kb", "mode": "ro"}],
            "config_overrides": {}, "worker_spec_snapshot_id": 41,
            "source_market_application_id": 12, "source_market_release_id": 19,
            "run_count": 3,
            "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"
        }],
        "total": 1
    }"#;
    let parsed: ExpertListResponse = serde_json::from_str(raw).unwrap();
    assert_eq!(parsed.total, 1);
    assert_eq!(parsed.experts.len(), 1);
    let e = &parsed.experts[0];
    assert_eq!(e.slug, "reviewer");
    assert_eq!(e.skill_slugs, vec!["merge".to_string()]);
    assert_eq!(e.knowledge_mounts[0]["slug"], "kb");
    assert_eq!(e.worker_spec_snapshot_id, Some(41));
    assert_eq!(e.source_market_application_id, Some(12));
    assert_eq!(e.source_market_release_id, Some(19));
    // Round-trip preserves the jsonb payload.
    let out = serde_json::to_value(e).unwrap();
    assert_eq!(out["knowledge_mounts"][0]["mode"], "ro");
    assert_eq!(out["worker_spec_snapshot_id"], 41);
    assert_eq!(out["run_count"], 3);
}
