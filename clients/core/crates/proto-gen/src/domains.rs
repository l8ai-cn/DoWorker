//! Client-facing proto domains (mirrors types/BUILD.bazel; excludes runner/v1).

pub struct Domain {
    pub name: &'static str,
    pub version: &'static str,
    pub srcs: &'static [&'static str],
    pub deps: &'static [&'static str],
}

macro_rules! domain {
    ($name:literal, [$($source:literal),*], [$($dependency:literal),*]) => {
        Domain {
            name: $name,
            version: "v1",
            srcs: &[$($source),*],
            deps: &[$($dependency),*],
        }
    };
}

macro_rules! versioned_domain {
    ($name:literal, $version:literal, [$($source:literal),*], [$($dependency:literal),*]) => {
        Domain {
            name: $name,
            version: $version,
            srcs: &[$($source),*],
            deps: &[$($dependency),*],
        }
    };
}

pub const DOMAINS: &[Domain] = &[
    versioned_domain!(
        "agent_workbench",
        "v2",
        [
            "content.proto",
            "configuration.proto",
            "session_state.proto",
            "artifact.proto",
            "tool.proto",
            "command.proto",
            "session.proto",
            "runner_ingress.proto",
            "service.proto"
        ],
        []
    ),
    domain!("agent", ["agent.proto"], []),
    domain!("ai_resource", ["ai_resource.proto", "types.proto"], []),
    domain!("execution_cluster", ["execution_cluster.proto"], []),
    domain!("apikey", ["api_key.proto"], []),
    domain!("app_state", ["app_state.proto"], []),
    domain!("auth", ["auth.proto"], []),
    domain!("org", ["org.proto"], []),
    domain!("auth_state", ["auth_state.proto"], ["auth", "org"]),
    domain!("autopilot", ["autopilot.proto"], []),
    domain!("billing", ["billing.proto", "billing_admin.proto"], []),
    domain!("binding", ["binding.proto"], []),
    domain!("blockstore", ["blockstore.proto"], []),
    domain!(
        "blockstore_state",
        ["blockstore_state.proto"],
        ["blockstore"]
    ),
    domain!("channel", ["channel.proto"], []),
    domain!(
        "pod",
        [
            "agentpod_settings.proto",
            "worker_creation.proto",
            "worker_skill_publish.proto",
            "pod.proto"
        ],
        []
    ),
    domain!(
        "channel_state",
        ["channel_state.proto", "mutations.proto"],
        ["pod"]
    ),
    domain!("env_bundle", ["env_bundle.proto"], []),
    domain!(
        "extension",
        ["market.proto", "repo_mcp.proto", "repo_skill.proto"],
        []
    ),
    domain!("events", ["event_data.proto", "events.proto"], []),
    domain!("file", ["file.proto"], []),
    domain!("grant", ["grant.proto"], []),
    domain!("invitation", ["invitation.proto"], []),
    domain!("knowledgebase", ["knowledgebase.proto"], []),
    domain!("license", ["license.proto"], []),
    domain!("goalloop", ["goalloop.proto"], []),
    domain!("workflow", ["workflow.proto"], []),
    domain!("workflow_state", ["workflow_state.proto"], ["workflow"]),
    domain!("mesh", ["mesh.proto"], []),
    domain!("mesh_state", ["mesh_state.proto"], ["mesh"]),
    domain!("notification", ["notification.proto"], []),
    domain!(
        "orchestration_resource",
        ["orchestration_resource.proto"],
        []
    ),
    domain!("pod_state", ["pod_state.proto"], ["pod"]),
    domain!(
        "promocode",
        ["promocode.proto", "promocode_admin.proto"],
        []
    ),
    domain!("repository", ["repository.proto"], []),
    domain!("repo_state", ["repo_state.proto"], ["repository"]),
    domain!("runner_api", ["runner.proto"], []),
    domain!("runner_state", ["runner_state.proto"], ["runner_api"]),
    domain!("autopilot_state", ["autopilot_state.proto"], []),
    domain!("acp_state", ["acp_state.proto"], []),
    domain!("sso", ["sso.proto", "sso_admin.proto"], []),
    domain!(
        "support_ticket",
        ["support_ticket.proto", "support_ticket_admin.proto"],
        []
    ),
    domain!("ticket", ["ticket.proto"], []),
    domain!("ticket_state", ["ticket_state.proto"], ["ticket"]),
    domain!("ticket_relations", ["ticket_relations.proto"], []),
    domain!("token_usage", ["token_usage.proto"], []),
    domain!("user", ["user.proto"], []),
    domain!("user_credential", ["user_credential.proto"], []),
];
