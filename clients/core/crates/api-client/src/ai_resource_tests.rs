use std::sync::Arc;

use agentcloud_types::proto_ai_resource_v1 as ai;
use prost::Message;
use wiremock::matchers::{body_bytes, method, path};
use wiremock::{Mock, MockServer, ResponseTemplate};

use crate::{ApiClient, AuthTokenStore};

struct TokenStore;

impl AuthTokenStore for TokenStore {
    fn get_token(&self) -> Option<String> {
        Some("token".into())
    }
    fn get_refresh_token(&self) -> Option<String> {
        None
    }
    fn set_tokens(&self, _: String, _: String, _: Option<i64>) {}
    fn clear_tokens(&self) {}
    fn get_current_org_slug(&self) -> Option<String> {
        Some("acme".into())
    }
}

#[tokio::test]
async fn create_connection_uses_binary_request_and_decodes_safe_metadata() {
    let server = MockServer::start().await;
    let request = ai::CreateOrganizationConnectionRequest {
        org_slug: "acme".into(),
        identifier: "openai-main".into(),
        provider_key: "openai".into(),
        name: "OpenAI".into(),
        base_url: "https://api.openai.com".into(),
        credentials: [("api_key".into(), "secret-value".into())].into(),
    };
    let response = ai::ProviderConnection {
        id: 7,
        configured_fields: vec!["api_key".into()],
        ..Default::default()
    };
    Mock::given(method("POST"))
        .and(path(
            "/proto.ai_resource.v1.AIResourceService/CreateOrganizationConnection",
        ))
        .and(body_bytes(request.encode_to_vec()))
        .respond_with(ResponseTemplate::new(200).set_body_bytes(response.encode_to_vec()))
        .mount(&server)
        .await;

    let client = ApiClient::new(server.uri(), Arc::new(TokenStore));
    let decoded = client
        .create_organization_connection_connect(&request)
        .await
        .unwrap();

    assert_eq!(decoded.id, 7);
    assert_eq!(decoded.configured_fields, ["api_key"]);
    assert!(!format!("{decoded:?}").contains("secret-value"));
}
