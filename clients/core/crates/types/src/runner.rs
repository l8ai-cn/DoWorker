use serde::{Deserialize, Serialize};

// Interactive registration (Tailscale-style device authorization). The
// browser polls /runners/grpc/auth-status while the runner waits for
// authorization. These two structs are converted to proto.runner_api.v1.*
// on the api-client boundary. They're STRIPPED projections of the proto
// messages (omit org_slug + sensitive mTLS cert fields), kept as Rust types so:
//   1. The wasm-bindgen surface doesn't leak certs into renderer memory.
//   2. Hosts can consume typed structs without re-implementing proto codegen
//      for two single-use messages.

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuthorizeRunnerRequest {
    pub auth_key: String,
    pub node_id: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RunnerAuthStatus {
    pub status: String,
    pub runner_id: Option<i64>,
    pub organization_slug: Option<String>,
}
