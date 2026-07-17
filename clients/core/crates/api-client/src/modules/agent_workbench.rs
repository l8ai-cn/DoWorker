use agentsmesh_types::proto_agent_workbench_v2 as v2;
use futures::stream::Stream;

use crate::connect_call::connect_call_with_bearer;
use crate::{ApiClient, ApiError};

const GET_SNAPSHOT: &str = "/proto.agent_workbench.v2.AgentWorkbenchService/GetSessionSnapshot";
const STREAM_DELTAS: &str = "/proto.agent_workbench.v2.AgentWorkbenchService/StreamSessionDeltas";
const EXECUTE_COMMAND: &str = "/proto.agent_workbench.v2.AgentWorkbenchService/ExecuteCommand";

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct AgentWorkbenchAccessScope {
    org_slug: String,
    bearer_token: String,
}

impl AgentWorkbenchAccessScope {
    pub fn new(
        org_slug: impl Into<String>,
        bearer_token: impl Into<String>,
    ) -> Result<Self, ApiError> {
        let org_slug = org_slug.into();
        let bearer_token = bearer_token.into();
        if org_slug.trim().is_empty() {
            return Err(ApiError::Decode(
                "agent workbench access scope requires org_slug".into(),
            ));
        }
        if bearer_token.trim().is_empty() {
            return Err(ApiError::Decode(
                "agent workbench access scope requires bearer token".into(),
            ));
        }
        Ok(Self {
            org_slug,
            bearer_token,
        })
    }
}

impl ApiClient {
    pub async fn get_agent_workbench_session_snapshot_connect(
        &self,
        access: &AgentWorkbenchAccessScope,
        session_id: &str,
    ) -> Result<v2::SessionSnapshot, ApiError> {
        let request = v2::GetSessionSnapshotRequest {
            org_slug: access.org_slug.clone(),
            session_id: session_id.into(),
        };
        connect_call_with_bearer(self, GET_SNAPSHOT, &request, &access.bearer_token).await
    }

    pub async fn execute_agent_workbench_command_connect(
        &self,
        access: &AgentWorkbenchAccessScope,
        command: v2::CommandEnvelope,
    ) -> Result<v2::CommandReceipt, ApiError> {
        let request = v2::ExecuteCommandRequest {
            org_slug: access.org_slug.clone(),
            command: Some(command),
        };
        connect_call_with_bearer(self, EXECUTE_COMMAND, &request, &access.bearer_token).await
    }

    #[cfg(not(target_arch = "wasm32"))]
    pub async fn stream_agent_workbench_session_deltas_connect_native(
        &self,
        access: &AgentWorkbenchAccessScope,
        cursor: v2::SessionCursor,
        replay_limit: u32,
    ) -> Result<impl Stream<Item = Result<v2::SessionDeltaBatch, ApiError>>, ApiError> {
        let request = agent_workbench_delta_request(access, cursor, replay_limit);
        self.connect_server_stream_native_with_bearer(STREAM_DELTAS, &request, &access.bearer_token)
            .await
    }

    #[cfg(target_arch = "wasm32")]
    pub async fn stream_agent_workbench_session_deltas_connect_wasm(
        &self,
        access: &AgentWorkbenchAccessScope,
        cursor: v2::SessionCursor,
        replay_limit: u32,
    ) -> Result<
        (
            impl Stream<Item = Result<v2::SessionDeltaBatch, ApiError>>,
            crate::WasmAbortHandle,
        ),
        ApiError,
    > {
        let request = agent_workbench_delta_request(access, cursor, replay_limit);
        self.connect_server_stream_wasm_with_bearer(STREAM_DELTAS, &request, &access.bearer_token)
            .await
    }
}

fn agent_workbench_delta_request(
    access: &AgentWorkbenchAccessScope,
    cursor: v2::SessionCursor,
    replay_limit: u32,
) -> v2::StreamSessionDeltasRequest {
    v2::StreamSessionDeltasRequest {
        org_slug: access.org_slug.clone(),
        cursor: Some(cursor),
        replay_limit,
    }
}
