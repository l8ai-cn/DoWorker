use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApiKey {
    pub id: i64,
    pub name: String,
    pub description: Option<String>,
    pub key_prefix: Option<String>,
    pub scopes: Option<Vec<String>>,
    pub expires_at: Option<String>,
    pub last_used_at: Option<String>,
    pub is_revoked: Option<bool>,
    pub created_at: Option<String>,
    pub updated_at: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateApiKeyRequest {
    pub name: String,
    pub description: Option<String>,
    pub scopes: Option<Vec<String>>,
    pub expires_in: Option<i64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateApiKeyRequest {
    pub name: Option<String>,
    pub description: Option<String>,
    pub scopes: Option<Vec<String>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApiKeyListResponse {
    pub api_keys: Vec<ApiKey>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateApiKeyResponse {
    pub api_key: ApiKey,
    pub raw_key: String,
}

#[cfg(test)]
mod tests {
    use super::*;

    const BACKEND_CREATE_RESPONSE: &str = r#"{
        "api_key": {
            "id": 42,
            "name": "ci-bot",
            "description": "CI integration",
            "key_prefix": "amk_abcd1234",
            "scopes": ["pods:read", "pods:write"],
            "expires_at": null,
            "last_used_at": null,
            "is_revoked": false,
            "created_at": "2026-05-09T00:00:00Z",
            "updated_at": "2026-05-09T00:00:00Z"
        },
        "raw_key": "amk_abcd1234567890abcdef..."
    }"#;

    #[test]
    fn create_response_decodes_backend_payload() {
        let resp: CreateApiKeyResponse = serde_json::from_str(BACKEND_CREATE_RESPONSE).unwrap();
        assert_eq!(resp.api_key.id, 42);
        assert_eq!(resp.api_key.name, "ci-bot");
        assert_eq!(resp.raw_key, "amk_abcd1234567890abcdef...");
    }

    #[test]
    fn create_response_wasm_relay_preserves_raw_key() {
        let typed: CreateApiKeyResponse = serde_json::from_str(BACKEND_CREATE_RESPONSE).unwrap();
        let relayed = serde_json::to_string(&typed).unwrap();
        let parsed: serde_json::Value = serde_json::from_str(&relayed).unwrap();

        assert!(!parsed["raw_key"].is_null(), "raw_key dropped by relay");
        assert!(!parsed["api_key"]["id"].is_null(), "api_key.id dropped");
        assert_eq!(parsed["raw_key"], "amk_abcd1234567890abcdef...");
    }
}
