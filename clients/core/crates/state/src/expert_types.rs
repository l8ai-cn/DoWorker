use serde::{Deserialize, Serialize};

// Client-side wire DTO for the Expert domain. Expert has no proto/Connect
// coverage on the backend (REST+JSON via Gin), so — unlike repo/runner/pod —
// the cache type is a hand-written serde struct instead of a re-exported prost
// message. `knowledge_mounts` / `config_overrides` are backend `jsonb` columns
// with no fixed client-side shape, so they stay as opaque `serde_json::Value`
// to round-trip losslessly.
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(default)]
pub struct Expert {
    pub id: i64,
    pub slug: String,
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    pub agent_slug: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub runner_id: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub repository_id: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub branch_name: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub prompt: Option<String>,
    pub interaction_mode: String,
    pub perpetual: bool,
    pub used_env_bundles: Vec<String>,
    pub skill_slugs: Vec<String>,
    pub knowledge_mounts: serde_json::Value,
    pub config_overrides: serde_json::Value,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub agentfile_layer: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub source_pod_key: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub worker_spec_snapshot_id: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub source_market_application_id: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub source_market_release_id: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub organization_id: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub created_by_id: Option<i64>,
    pub run_count: i64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub last_run_at: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

/// Envelope for the `GET .../experts` list response (`{experts, total}`).
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct ExpertListResponse {
    #[serde(default)]
    pub experts: Vec<Expert>,
    #[serde(default)]
    pub total: i64,
}

/// Envelope for single-expert responses (`{expert}`).
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct ExpertEnvelope {
    pub expert: Expert,
}
