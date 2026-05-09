use agentsmesh_types::{
    ActorType, ApplyOpsRequest, ApplyOpsResult, Block, BlockOp, BlockRef, ChildrenResult,
    MeshChannelInfo, MeshEdge, MeshNode, MeshRunnerInfo, MeshTopology, NotificationPreference,
    NotificationPreferenceListResponse, OpEnvelope, OpKind, PodStatus, RunnerStatus, SearchHit,
    SemanticSearchRequest, SetNotificationPreferenceRequest, Workspace,
};

use super::{PodStatusDto, RunnerStatusDto};

// ── Blockstore ────────────────────────────────────────────

#[derive(Clone, Copy, Debug, uniffi::Enum)]
pub enum ActorTypeDto {
    User,
    Agent,
    System,
}

impl From<ActorType> for ActorTypeDto {
    fn from(a: ActorType) -> Self {
        match a {
            ActorType::User => Self::User,
            ActorType::Agent => Self::Agent,
            ActorType::System => Self::System,
        }
    }
}

impl From<ActorTypeDto> for ActorType {
    fn from(a: ActorTypeDto) -> Self {
        match a {
            ActorTypeDto::User => Self::User,
            ActorTypeDto::Agent => Self::Agent,
            ActorTypeDto::System => Self::System,
        }
    }
}

#[derive(Clone, Copy, Debug, uniffi::Enum)]
pub enum OpKindDto {
    CreateBlock,
    UpdateBlock,
    DeleteBlock,
    AddRef,
    RemoveRef,
    UpdateRef,
}

impl From<OpKind> for OpKindDto {
    fn from(o: OpKind) -> Self {
        match o {
            OpKind::CreateBlock => Self::CreateBlock,
            OpKind::UpdateBlock => Self::UpdateBlock,
            OpKind::DeleteBlock => Self::DeleteBlock,
            OpKind::AddRef => Self::AddRef,
            OpKind::RemoveRef => Self::RemoveRef,
            OpKind::UpdateRef => Self::UpdateRef,
        }
    }
}

impl From<OpKindDto> for OpKind {
    fn from(o: OpKindDto) -> Self {
        match o {
            OpKindDto::CreateBlock => Self::CreateBlock,
            OpKindDto::UpdateBlock => Self::UpdateBlock,
            OpKindDto::DeleteBlock => Self::DeleteBlock,
            OpKindDto::AddRef => Self::AddRef,
            OpKindDto::RemoveRef => Self::RemoveRef,
            OpKindDto::UpdateRef => Self::UpdateRef,
        }
    }
}

fn value_to_string(v: serde_json::Value) -> String {
    serde_json::to_string(&v).unwrap_or_else(|_| "{}".into())
}

fn string_to_value(s: String) -> serde_json::Value {
    serde_json::from_str(&s).unwrap_or(serde_json::Value::Null)
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct BlockDto {
    pub id: String,
    pub workspace_id: String,
    pub block_type: String,
    pub data_json: String,
    pub text: Option<String>,
    pub meta_json: String,
    pub created_by: i64,
    pub created_at: String,
    pub updated_at: String,
    pub deleted_at: Option<String>,
}

impl From<Block> for BlockDto {
    fn from(b: Block) -> Self {
        Self {
            id: b.id,
            workspace_id: b.workspace_id,
            block_type: b.block_type,
            data_json: value_to_string(b.data),
            text: b.text,
            meta_json: value_to_string(b.meta),
            created_by: b.created_by,
            created_at: b.created_at,
            updated_at: b.updated_at,
            deleted_at: b.deleted_at,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct BlockRefDto {
    pub id: i64,
    pub workspace_id: String,
    pub from_id: String,
    pub to_id: String,
    pub rel: String,
    pub order_key: Option<String>,
    pub anchor: Option<String>,
    pub meta_json: String,
    pub created_by: i64,
    pub created_at: String,
    pub updated_at: String,
}

impl From<BlockRef> for BlockRefDto {
    fn from(r: BlockRef) -> Self {
        Self {
            id: r.id,
            workspace_id: r.workspace_id,
            from_id: r.from_id,
            to_id: r.to_id,
            rel: r.rel,
            order_key: r.order_key,
            anchor: r.anchor,
            meta_json: value_to_string(r.meta),
            created_by: r.created_by,
            created_at: r.created_at,
            updated_at: r.updated_at,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct BlockOpDto {
    pub id: i64,
    pub workspace_id: String,
    pub idempotency_key: Option<String>,
    pub actor_type: ActorTypeDto,
    pub actor_id: i64,
    pub op: OpKindDto,
    pub target_block: Option<String>,
    pub target_ref: Option<i64>,
    pub payload_json: String,
    pub forward_json: String,
    pub inverse_json: String,
    pub parent_op_id: Option<i64>,
    pub applied_at: String,
}

impl From<BlockOp> for BlockOpDto {
    fn from(o: BlockOp) -> Self {
        Self {
            id: o.id,
            workspace_id: o.workspace_id,
            idempotency_key: o.idempotency_key,
            actor_type: o.actor_type.into(),
            actor_id: o.actor_id,
            op: o.op.into(),
            target_block: o.target_block,
            target_ref: o.target_ref,
            payload_json: value_to_string(o.payload),
            forward_json: value_to_string(o.forward),
            inverse_json: value_to_string(o.inverse),
            parent_op_id: o.parent_op_id,
            applied_at: o.applied_at,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct WorkspaceDto {
    pub id: String,
    pub organization_id: i64,
    pub slug: String,
    pub name: String,
    pub root_block_id: Option<String>,
    pub created_at: String,
}

impl From<Workspace> for WorkspaceDto {
    fn from(w: Workspace) -> Self {
        Self {
            id: w.id,
            organization_id: w.organization_id,
            slug: w.slug,
            name: w.name,
            root_block_id: w.root_block_id,
            created_at: w.created_at,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct OpEnvelopeDto {
    pub op: OpKindDto,
    pub payload_json: String,
}

impl From<OpEnvelopeDto> for OpEnvelope {
    fn from(d: OpEnvelopeDto) -> Self {
        Self {
            op: d.op.into(),
            payload: string_to_value(d.payload_json),
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct ApplyOpsRequestDto {
    pub workspace_id: String,
    pub ops: Vec<OpEnvelopeDto>,
    pub idempotency_key: Option<String>,
    pub parent_op_id: Option<i64>,
}

impl From<ApplyOpsRequestDto> for ApplyOpsRequest {
    fn from(d: ApplyOpsRequestDto) -> Self {
        Self {
            workspace_id: d.workspace_id,
            ops: d.ops.into_iter().map(Into::into).collect(),
            idempotency_key: d.idempotency_key,
            parent_op_id: d.parent_op_id,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct ApplyOpsResultDto {
    pub op_ids: Vec<i64>,
    pub was_replay: bool,
    pub parent_op_id: Option<i64>,
}

impl From<ApplyOpsResult> for ApplyOpsResultDto {
    fn from(r: ApplyOpsResult) -> Self {
        Self {
            op_ids: r.op_ids,
            was_replay: r.was_replay,
            parent_op_id: r.parent_op_id,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct ChildrenResultDto {
    pub blocks: Vec<BlockDto>,
    pub refs: Vec<BlockRefDto>,
}

impl From<ChildrenResult> for ChildrenResultDto {
    fn from(r: ChildrenResult) -> Self {
        Self {
            blocks: r.blocks.into_iter().map(BlockDto::from).collect(),
            refs: r.refs.into_iter().map(BlockRefDto::from).collect(),
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct SearchHitDto {
    pub block_id: String,
    pub block_type: String,
    pub snippet: String,
    pub score: f64,
}

impl From<SearchHit> for SearchHitDto {
    fn from(h: SearchHit) -> Self {
        Self {
            block_id: h.block_id,
            block_type: h.block_type,
            snippet: h.snippet,
            score: h.score,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct SemanticSearchRequestDto {
    pub query: String,
    pub top_k: Option<u32>,
    pub min_score: Option<f64>,
    pub block_type: Option<String>,
}

impl From<SemanticSearchRequestDto> for SemanticSearchRequest {
    fn from(d: SemanticSearchRequestDto) -> Self {
        Self {
            query: d.query,
            top_k: d.top_k,
            min_score: d.min_score,
            block_type: d.block_type,
        }
    }
}

// ── Mesh ──────────────────────────────────────────────────

#[derive(Clone, Debug, uniffi::Record)]
pub struct MeshNodeDto {
    pub pod_key: String,
    pub alias: Option<String>,
    pub status: PodStatusDto,
    pub agent_status: Option<String>,
    pub agent_slug: String,
    pub runner_id: Option<i64>,
    pub model: Option<String>,
    pub title: Option<String>,
    pub ticket_id: Option<i64>,
    pub ticket_slug: Option<String>,
    pub ticket_title: Option<String>,
    pub repository_id: Option<i64>,
    pub created_by_id: Option<i64>,
    pub runner_node_id: Option<String>,
    pub runner_status: Option<String>,
    pub started_at: Option<String>,
}

impl From<MeshNode> for MeshNodeDto {
    fn from(n: MeshNode) -> Self {
        let _: PodStatus = n.status;
        Self {
            pod_key: n.pod_key,
            alias: n.alias,
            status: n.status.into(),
            agent_status: n.agent_status,
            agent_slug: n.agent_slug,
            runner_id: n.runner_id,
            model: n.model,
            title: n.title,
            ticket_id: n.ticket_id,
            ticket_slug: n.ticket_slug,
            ticket_title: n.ticket_title,
            repository_id: n.repository_id,
            created_by_id: n.created_by_id,
            runner_node_id: n.runner_node_id,
            runner_status: n.runner_status,
            started_at: n.started_at,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct MeshEdgeDto {
    pub id: Option<i64>,
    pub source: String,
    pub target: String,
    pub binding_status: Option<String>,
    pub status: Option<String>,
    pub granted_scopes: Option<Vec<String>>,
    pub pending_scopes: Option<Vec<String>>,
}

impl From<MeshEdge> for MeshEdgeDto {
    fn from(e: MeshEdge) -> Self {
        Self {
            id: e.id,
            source: e.source,
            target: e.target,
            binding_status: e.binding_status,
            status: e.status,
            granted_scopes: e.granted_scopes,
            pending_scopes: e.pending_scopes,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct MeshChannelInfoDto {
    pub id: i64,
    pub name: String,
    pub description: Option<String>,
    pub pod_keys: Vec<String>,
    pub message_count: Option<i64>,
    pub is_archived: Option<bool>,
}

impl From<MeshChannelInfo> for MeshChannelInfoDto {
    fn from(c: MeshChannelInfo) -> Self {
        Self {
            id: c.id,
            name: c.name,
            description: c.description,
            pod_keys: c.pod_keys,
            message_count: c.message_count,
            is_archived: c.is_archived,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct MeshRunnerInfoDto {
    pub id: i64,
    pub name: String,
    pub status: RunnerStatusDto,
    pub node_id: Option<String>,
    pub max_concurrent_pods: Option<i32>,
    pub current_pods: Option<i32>,
    pub pod_keys: Vec<String>,
}

impl From<MeshRunnerInfo> for MeshRunnerInfoDto {
    fn from(r: MeshRunnerInfo) -> Self {
        let _: RunnerStatus = r.status;
        Self {
            id: r.id,
            name: r.name,
            status: r.status.into(),
            node_id: r.node_id,
            max_concurrent_pods: r.max_concurrent_pods,
            current_pods: r.current_pods,
            pod_keys: r.pod_keys,
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct MeshTopologyDto {
    pub nodes: Vec<MeshNodeDto>,
    pub edges: Vec<MeshEdgeDto>,
    pub channels: Vec<MeshChannelInfoDto>,
    pub runners: Vec<MeshRunnerInfoDto>,
}

impl From<MeshTopology> for MeshTopologyDto {
    fn from(t: MeshTopology) -> Self {
        Self {
            nodes: t.nodes.into_iter().map(MeshNodeDto::from).collect(),
            edges: t.edges.into_iter().map(MeshEdgeDto::from).collect(),
            channels: t
                .channels
                .into_iter()
                .map(MeshChannelInfoDto::from)
                .collect(),
            runners: t.runners.into_iter().map(MeshRunnerInfoDto::from).collect(),
        }
    }
}

// ── Notification ──────────────────────────────────────────

#[derive(Clone, Debug, uniffi::Record)]
pub struct NotificationPreferenceDto {
    pub source: Option<String>,
    pub entity_id: Option<String>,
    pub is_muted: Option<bool>,
    pub channels: Option<Vec<String>>,
}

impl From<NotificationPreference> for NotificationPreferenceDto {
    fn from(p: NotificationPreference) -> Self {
        Self {
            source: p.source,
            entity_id: p.entity_id,
            is_muted: p.is_muted,
            channels: p.channels.map(|m| m.into_iter().filter(|(_, v)| *v).map(|(k, _)| k).collect()),
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct NotificationPreferenceListResponseDto {
    pub preferences: Vec<NotificationPreferenceDto>,
}

impl From<NotificationPreferenceListResponse> for NotificationPreferenceListResponseDto {
    fn from(r: NotificationPreferenceListResponse) -> Self {
        Self {
            preferences: r
                .preferences
                .into_iter()
                .map(NotificationPreferenceDto::from)
                .collect(),
        }
    }
}

#[derive(Clone, Debug, uniffi::Record)]
pub struct SetNotificationPreferenceRequestDto {
    pub source: String,
    pub entity_id: Option<String>,
    pub is_muted: Option<bool>,
    pub channels: Option<Vec<String>>,
}

impl From<SetNotificationPreferenceRequestDto> for SetNotificationPreferenceRequest {
    fn from(d: SetNotificationPreferenceRequestDto) -> Self {
        Self {
            source: d.source,
            entity_id: d.entity_id,
            is_muted: d.is_muted,
            channels: d.channels.map(|v| v.into_iter().map(|k| (k, true)).collect()),
        }
    }
}
