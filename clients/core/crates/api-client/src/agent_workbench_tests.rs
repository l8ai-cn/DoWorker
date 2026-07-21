use std::sync::{Arc, Mutex};

use agentcloud_types::proto_agent_workbench_v2 as v2;
use futures::StreamExt;
use prost::Message;
use wiremock::matchers::{body_bytes, header, method, path};
use wiremock::{Mock, MockServer, ResponseTemplate};

use crate::{AgentWorkbenchAccessScope, ApiClient, ApiError, AuthTokenStore};

struct MutableTokenStore {
    token: Mutex<Option<String>>,
    org_slug: Mutex<Option<String>>,
}

impl MutableTokenStore {
    fn new(token: &str, org_slug: &str) -> Arc<Self> {
        Arc::new(Self {
            token: Mutex::new(Some(token.into())),
            org_slug: Mutex::new(Some(org_slug.into())),
        })
    }

    fn set(&self, token: &str, org_slug: &str) {
        *self.token.lock().unwrap() = Some(token.into());
        *self.org_slug.lock().unwrap() = Some(org_slug.into());
    }
}

impl AuthTokenStore for MutableTokenStore {
    fn get_token(&self) -> Option<String> {
        self.token.lock().unwrap().clone()
    }

    fn get_refresh_token(&self) -> Option<String> {
        None
    }

    fn set_tokens(&self, _: String, _: String, _: Option<i64>) {}

    fn clear_tokens(&self) {}

    fn get_current_org_slug(&self) -> Option<String> {
        self.org_slug.lock().unwrap().clone()
    }
}

#[tokio::test]
async fn unary_calls_use_explicit_access_scope_instead_of_token_store() {
    let server = MockServer::start().await;
    let store = MutableTokenStore::new("stale-token", "stale-org");
    let client = ApiClient::new(server.uri(), store.clone());
    let snapshot = snapshot();
    let access = AgentWorkbenchAccessScope::new("org-a", "token-a").unwrap();

    mount_unary(
        &server,
        "/proto.agent_workbench.v2.AgentWorkbenchService/GetSessionSnapshot",
        "token-a",
        v2::GetSessionSnapshotRequest {
            org_slug: "org-a".into(),
            session_id: "session-1".into(),
        },
        snapshot.clone(),
    )
    .await;

    assert_eq!(
        client
            .get_agent_workbench_session_snapshot_connect(&access, "session-1")
            .await
            .unwrap(),
        snapshot
    );

    store.set("other-stale-token", "other-stale-org");
    let access = AgentWorkbenchAccessScope::new("org-b", "token-b").unwrap();
    let command = v2::CommandEnvelope {
        session_id: "session-1".into(),
        command_id: "command-1".into(),
        ..Default::default()
    };
    let receipt = v2::CommandReceipt {
        session_id: "session-1".into(),
        command_id: "command-1".into(),
        state: v2::CommandReceiptState::Received as i32,
        payload_digest: "digest-1".into(),
        ..Default::default()
    };
    mount_unary(
        &server,
        "/proto.agent_workbench.v2.AgentWorkbenchService/ExecuteCommand",
        "token-b",
        v2::ExecuteCommandRequest {
            org_slug: "org-b".into(),
            command: Some(command.clone()),
        },
        receipt.clone(),
    )
    .await;

    assert_eq!(
        client
            .execute_agent_workbench_command_connect(&access, command)
            .await
            .unwrap(),
        receipt
    );
}

#[tokio::test]
async fn delta_stream_uses_connect_frames_and_decodes_batches() {
    let server = MockServer::start().await;
    let client = ApiClient::new(
        server.uri(),
        MutableTokenStore::new("stale-token", "stale-org"),
    );
    let access = AgentWorkbenchAccessScope::new("stream-org", "stream-token").unwrap();

    let cursor = v2::SessionCursor {
        session_id: "session-1".into(),
        stream_epoch: "epoch-1".into(),
        revision: 4,
        sequence: 9,
    };
    let request = v2::StreamSessionDeltasRequest {
        org_slug: "stream-org".into(),
        cursor: Some(cursor.clone()),
        replay_limit: 64,
    };
    let batch = v2::SessionDeltaBatch {
        session_id: "session-1".into(),
        stream_epoch: "epoch-1".into(),
        base_revision: 4,
        revision: 5,
        first_sequence: 10,
        last_sequence: 10,
        digest: "batch-1".into(),
        ..Default::default()
    };
    let mut response = message_frame(&batch);
    response.extend(end_frame());

    Mock::given(method("POST"))
        .and(path(
            "/proto.agent_workbench.v2.AgentWorkbenchService/StreamSessionDeltas",
        ))
        .and(header("authorization", "Bearer stream-token"))
        .and(header("content-type", "application/connect+proto"))
        .and(body_bytes(message_frame(&request)))
        .respond_with(ResponseTemplate::new(200).set_body_bytes(response))
        .mount(&server)
        .await;

    let stream = client
        .stream_agent_workbench_session_deltas_connect_native(&access, cursor, 64)
        .await
        .unwrap();
    futures::pin_mut!(stream);
    assert_eq!(stream.next().await.unwrap().unwrap(), batch);
    assert!(stream.next().await.is_none());
}

#[tokio::test]
async fn delta_stream_maps_unauthorized_to_auth_expired() {
    let server = MockServer::start().await;
    Mock::given(method("POST"))
        .and(path(
            "/proto.agent_workbench.v2.AgentWorkbenchService/StreamSessionDeltas",
        ))
        .respond_with(ResponseTemplate::new(401))
        .mount(&server)
        .await;
    let client = ApiClient::new(
        server.uri(),
        MutableTokenStore::new("expired", "stream-org"),
    );
    let access = AgentWorkbenchAccessScope::new("stream-org", "expired").unwrap();

    let error = client
        .stream_agent_workbench_session_deltas_connect_native(
            &access,
            v2::SessionCursor::default(),
            1,
        )
        .await
        .err()
        .unwrap();
    assert!(matches!(error, ApiError::AuthExpired));
}

#[test]
fn access_scope_rejects_missing_org_or_bearer_token() {
    assert!(AgentWorkbenchAccessScope::new("", "token").is_err());
    assert!(AgentWorkbenchAccessScope::new("org-a", "").is_err());
}

#[test]
fn cloned_base_url_keeps_the_same_auth_store() {
    let store = MutableTokenStore::new("token-a", "org-a");
    let client = ApiClient::new("http://proxy.test".into(), store.clone());
    let stream_client = client.clone_with_base_url("http://stream.test".into());

    assert_eq!(stream_client.base_url, "http://stream.test");
    assert_eq!(stream_client.current_org_slug(), "org-a");

    store.set("token-b", "org-b");
    assert_eq!(stream_client.current_org_slug(), "org-b");
}

async fn mount_unary<Req, Res>(
    server: &MockServer,
    procedure: &'static str,
    token: &'static str,
    request: Req,
    response: Res,
) where
    Req: Message,
    Res: Message,
{
    Mock::given(method("POST"))
        .and(path(procedure))
        .and(header("authorization", format!("Bearer {token}")))
        .and(body_bytes(request.encode_to_vec()))
        .respond_with(ResponseTemplate::new(200).set_body_bytes(response.encode_to_vec()))
        .mount(server)
        .await;
}

fn message_frame(message: &impl Message) -> Vec<u8> {
    let payload = message.encode_to_vec();
    let mut frame = Vec::with_capacity(5 + payload.len());
    frame.push(0);
    frame.extend_from_slice(&(payload.len() as u32).to_be_bytes());
    frame.extend(payload);
    frame
}

fn end_frame() -> Vec<u8> {
    vec![2, 0, 0, 0, 2, b'{', b'}']
}

fn snapshot() -> v2::SessionSnapshot {
    v2::SessionSnapshot {
        session_id: "session-1".into(),
        stream_epoch: "epoch-1".into(),
        revision: 4,
        latest_sequence: 9,
        status: v2::SessionStatus::Running as i32,
        ..Default::default()
    }
}
