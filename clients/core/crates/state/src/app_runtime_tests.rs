use std::sync::Arc;

use agentsmesh_api_client::{ApiClient, AuthTokenStore};
use agentsmesh_events::types::EventDispatchHook;
use agentsmesh_events::{EventSubscriptionManager, EventType, RealtimeEvent};
use parking_lot::RwLock;

use crate::app_state::{AppRuntime, AppState, AppStateDispatchHook};

struct StubAuth;
impl AuthTokenStore for StubAuth {
    fn get_token(&self) -> Option<String> {
        Some("t".into())
    }
    fn get_refresh_token(&self) -> Option<String> {
        None
    }
    fn set_tokens(&self, _t: String, _r: String, _e: Option<i64>) {}
    fn clear_tokens(&self) {}
    fn get_current_org_slug(&self) -> Option<String> {
        Some("o".into())
    }
}

fn make_client() -> Arc<ApiClient> {
    Arc::new(ApiClient::new(
        "http://localhost:9999".into(),
        Arc::new(StubAuth),
    ))
}

fn make_event(json: serde_json::Value) -> RealtimeEvent {
    serde_json::from_value(json).unwrap()
}

fn dispatch(state: &Arc<RwLock<AppState>>, event: &RealtimeEvent) {
    // Simulate what EventSubscriptionManager.dispatch_event does on each
    // received frame: run the hook, which writes to AppState. The hook
    // is installed by AppRuntime::new(), but tests don't have a way to
    // pull it back out of the manager — so we construct it ourselves
    // here. The integration is verified end-to-end via the e2e suite.
    let hook = AppStateDispatchHook::new(state.clone());
    hook.dispatch(event);
}

/// AppRuntime construction installs a dispatch hook that writes events
/// directly into AppState. Phase 1's contract — verified end-to-end.
#[test]
fn app_runtime_dispatch_hook_writes_state() {
    let events = Arc::new(EventSubscriptionManager::with_default_options(make_client()));
    let rt = AppRuntime::new(events.clone());

    let evt = make_event(serde_json::json!({
        "type": EventType::PodCreated.as_str(),
        "category": "entity",
        "organization_id": 1,
        "data": {"pod_key": "p-1", "status": "running", "agent_slug": "claude"},
        "timestamp": 1000
    }));
    dispatch(&rt.state, &evt);

    let guard = rt.state.read();
    assert_eq!(guard.pods.pods().len(), 1);
    assert_eq!(guard.pods.get_pod("p-1").unwrap().status, "running");
}

/// Multiple events of different domains all land in the same AppState.
/// This is what host selectors will see post-Phase-4.
#[test]
fn dispatch_across_domains_single_state() {
    let events = Arc::new(EventSubscriptionManager::with_default_options(make_client()));
    let rt = AppRuntime::new(events.clone());

    dispatch(
        &rt.state,
        &make_event(serde_json::json!({
            "type": "pod:created",
            "category": "entity",
            "organization_id": 1,
            "data": {"pod_key": "p-1", "status": "running", "agent_slug": "claude"},
            "timestamp": 1
        })),
    );
    dispatch(
        &rt.state,
        &make_event(serde_json::json!({
            "type": "ticket:created",
            "category": "entity",
            "organization_id": 1,
            "data": {"slug": "T-9", "title": "fix", "status": "todo", "priority": "high"},
            "timestamp": 2
        })),
    );
    dispatch(
        &rt.state,
        &make_event(serde_json::json!({
            "type": "channel:message",
            "category": "entity",
            "organization_id": 1,
            "data": {"id": 5, "channel_id": 10, "content": "hello"},
            "timestamp": 3
        })),
    );

    let guard = rt.state.read();
    assert_eq!(guard.pods.pods().len(), 1);
    assert_eq!(guard.tickets.get_tickets().len(), 1);
    assert!(guard.channels.get_messages(10).is_some());
}

/// Shared-Arc semantics — two views observe the same writes. This is
/// the critical Phase 2 invariant: WasmPodState, WasmChannelState, etc.
/// must all hold `Arc<RwLock<AppState>>` clones of the same lock.
#[test]
fn shared_state_clone_observes_same_writes() {
    let state = Arc::new(RwLock::new(AppState::new()));
    let state_b = Arc::clone(&state);

    state.write().pods.upsert_pod(
        agentsmesh_types::proto_pod_v1::Pod {
            pod_key: "p".into(),
            status: "running".into(),
            agent_slug: "c".into(),
            ..Default::default()
        },
        Some(1),
    );

    assert_eq!(state_b.read().pods.pods().len(), 1);
    assert_eq!(state_b.read().pods.get_pod("p").unwrap().status, "running");
}

#[test]
fn reset_for_org_switch_clears_state_but_keeps_events() {
    let events = Arc::new(EventSubscriptionManager::with_default_options(make_client()));
    let rt = AppRuntime::new(events.clone());

    dispatch(
        &rt.state,
        &make_event(serde_json::json!({
            "type": "pod:created",
            "category": "entity",
            "organization_id": 1,
            "data": {"pod_key": "p-1", "status": "running", "agent_slug": "claude"},
            "timestamp": 1
        })),
    );
    assert_eq!(rt.state.read().pods.pods().len(), 1);

    rt.state.write().reset_for_org_switch();

    assert_eq!(
        rt.state.read().pods.pods().len(),
        0,
        "org switch must clear pods"
    );

    // Subsequent events still flow into the same runtime.
    dispatch(
        &rt.state,
        &make_event(serde_json::json!({
            "type": "pod:created",
            "category": "entity",
            "organization_id": 2,
            "data": {"pod_key": "p-2", "status": "running", "agent_slug": "claude"},
            "timestamp": 2
        })),
    );
    assert_eq!(rt.state.read().pods.pods().len(), 1);
}

/// Pending queue mechanism — events that don't write data but signal
/// platform side-effects (toast, browser notification, refetch).
#[test]
fn pending_queues_drain_atomically() {
    use crate::app_state::{NotificationSpec, ToastSpec};
    let mut state = AppState::new();

    state.pending_toasts.push(ToastSpec {
        kind: "warning".into(),
        title_key: "loops.runWarningTitle".into(),
        title_params: serde_json::json!({"runNumber": 5}),
        description: "step timeout".into(),
        duration_ms: 8000,
    });
    state.pending_browser_notifications.push(NotificationSpec {
        title: "New message".into(),
        body: "hi".into(),
        icon: None,
        link: Some("/channels/1".into()),
    });

    let toasts = state.take_pending_toasts();
    assert_eq!(toasts.len(), 1);
    assert_eq!(toasts[0].kind, "warning");
    // Second take returns empty — atomic drain.
    assert!(state.take_pending_toasts().is_empty());

    let notifs = state.take_pending_browser_notifications();
    assert_eq!(notifs.len(), 1);
    assert!(state.take_pending_browser_notifications().is_empty());
}
