// DTO layer: UniFFI `Record` / `Enum` definitions that map from internal
// `agentsmesh-types` structs to FFI-friendly shapes (no HashMap, no
// chrono, no lifetimes). Populated per service migration in Phase 1.
//
// Re-exports are kept broad so the generated Swift surface contains every
// named record; unused warnings are suppressed because uniffi consumes the
// types via `#[uniffi::export]` in service files, not via direct imports.
#![allow(unused_imports)]

mod automation;
mod blocks_mesh;
mod channel;
mod message;
mod misc;
mod pod;
mod ticket;
mod user;
mod workspace;

pub use automation::{
    AutopilotControllerDto, AutopilotIterationDto, AutopilotIterationListResponseDto,
    AutopilotListResponseDto, AutopilotStatusDto, CreateAutopilotRequestDto, CreateLoopRequestDto,
    LoopDataDto, LoopListResponseDto, LoopRunDataDto, LoopRunListResponseDto, LoopRunStatusDto,
    UpdateLoopRequestDto,
};
pub(crate) use automation::approve_autopilot_req;
pub use blocks_mesh::{
    ActorTypeDto, ApplyOpsRequestDto, ApplyOpsResultDto, BlockDto, BlockOpDto, BlockRefDto,
    ChildrenResultDto, MeshChannelInfoDto, MeshEdgeDto, MeshNodeDto, MeshRunnerInfoDto,
    MeshTopologyDto, NotificationPreferenceDto, NotificationPreferenceListResponseDto, OpEnvelopeDto,
    OpKindDto, SearchHitDto, SemanticSearchRequestDto, SetNotificationPreferenceRequestDto,
    WorkspaceDto,
};
pub use misc::{
    InvitationDto, InvitationListResponseDto, PresignRequestDto, PresignResponseDto,
    ResourceGrantDto, ResourceGrantListResponseDto, ResourceGrantResponseDto,
    ResourceGrantUserBriefDto,
};
pub(crate) use misc::{create_invitation_req, create_resource_grant_req};

pub use channel::{
    ChannelDto, ChannelListResponseDto, ChannelMemberDto, ChannelMemberListResponseDto,
    ChannelMessageDto, ChannelMessageListResponseDto, ChannelUnreadResponseDto,
    CreateChannelRequestDto, MessagePreviewDto, SenderAgentInfoDto, SenderPodInfoDto,
    UpdateChannelRequestDto,
};
pub(crate) use channel::{
    edit_message_req, invite_channel_members_req, join_channel_pod_req, mute_channel_req,
    send_message_req,
};
pub use message::{
    DeadLetterEntryDto, DeadLetterListResponseDto, DirectMessageDto, DirectMessageListResponseDto,
    ReplayDeadLetterResponseDto, SendDirectMessageRequestDto, UnreadCountResponseDto,
};
pub(crate) use message::mark_messages_read_req;
pub use ticket::{
    BoardColumnDto, BoardResponseDto, CreateLabelRequestDto, CreateTicketCommentRequestDto,
    CreateTicketRelationRequestDto, CreateTicketRequestDto, LabelDto, LabelListResponseDto,
    LinkTicketCommitRequestDto, TicketCommentDto, TicketCommentListResponseDto, TicketCommitDto,
    TicketCommitListResponseDto, TicketDto, TicketListResponseDto, TicketPriorityDto,
    TicketRelationDto, TicketRelationListResponseDto, TicketStatusDto, UpdateLabelRequestDto,
    UpdateTicketCommentRequestDto, UpdateTicketRequestDto,
};
pub(crate) use ticket::{add_assignee_req, add_ticket_label_req, update_ticket_status_req};

pub use pod::{
    CreatePodRequestDto, CreatePodResponseDto, PodAgentInfoDto, PodConnectionInfoDto,
    PodCreatedByInfoDto, PodDto, PodListResponseDto, PodLoopInfoDto, PodRepositoryInfoDto,
    PodRunnerInfoDto, PodStatusDto, PodTicketInfoDto,
};
pub(crate) use pod::update_pod_alias_req;
pub use user::{
    AuthSessionDto, AuthTokensDto, BootstrapCleanupReasonDto, BootstrapResultDto,
    OrganizationDto, SSOConfigDto, UserDto, UserIdentityDto,
};
pub use workspace::{
    AuthorizeRunnerRequestDto, BranchDto, CreateRepositoryRequestDto, CreateRunnerTokenRequestDto,
    GrpcRegistrationTokenDto, MergeRequestListResponseDto, RepositoryDto, RepositoryListResponseDto,
    RepositoryMergeRequestDto, RunnerAuthStatusDto, RunnerDto, RunnerListResponseDto, RunnerLogDto,
    RunnerLogListResponseDto, RunnerStatusDto, RunnerTokenListResponseDto, UpdateRepositoryRequestDto,
    UpdateRunnerRequestDto, UpgradeRunnerRequestDto, WebhookSecretDto, WebhookStatusDto,
};
