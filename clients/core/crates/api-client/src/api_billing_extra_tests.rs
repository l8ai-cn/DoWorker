#[cfg(test)]
mod api_billing_extra_tests {
    use std::sync::{Arc, Mutex};
    use serde_json::json;
    use wiremock::matchers::{method, path};
    use wiremock::{Mock, MockServer, ResponseTemplate};
    use crate::{ApiClient, AuthTokenStore};
    use agentsmesh_types::*;

    struct MT { org: Mutex<Option<String>> }
    impl MT { fn org(s: &str) -> Arc<Self> { Arc::new(Self { org: Mutex::new(Some(s.into())) }) } }
    impl AuthTokenStore for MT {
        fn get_token(&self) -> Option<String> { Some("t".into()) }
        fn get_refresh_token(&self) -> Option<String> { None }
        fn set_tokens(&self, _: String, _: String, _: Option<i64>) {}
        fn clear_tokens(&self) {}
        fn get_current_org_slug(&self) -> Option<String> { self.org.lock().unwrap().clone() }
    }
    fn ok(b: serde_json::Value) -> ResponseTemplate { ResponseTemplate::new(200).set_body_json(b) }

    #[tokio::test]
    async fn create_billing_checkout() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/a/billing/checkout"))
            .respond_with(ok(json!({"order_no":"o","status":"pending"}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), MT::org("a"));
        let _ = c.create_billing_checkout(&CheckoutRequest { plan_name: "pro".into(), order_type: None, billing_cycle: None, seats: None, provider: None, success_url: None, cancel_url: None }).await;
    }
    #[tokio::test]
    async fn get_billing_checkout_status() {
        let s = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/orgs/a/billing/checkout/o1"))
            .respond_with(ok(json!({"order_no":"o1","status":"completed"}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), MT::org("a"));
        let _ = c.get_billing_checkout_status("o1").await;
    }
    #[tokio::test]
    async fn request_cancel_subscription() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/a/billing/subscription/cancel"))
            .respond_with(ok(json!({}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), MT::org("a"));
        let _ = c.request_cancel_subscription(&CancelSubscriptionRequest { immediate: Some(false) }).await;
    }
    #[tokio::test]
    async fn reactivate_subscription() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/a/billing/subscription/reactivate"))
            .respond_with(ok(json!({"plan_name":"pro"}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), MT::org("a"));
        let _ = c.reactivate_subscription().await;
    }
    #[tokio::test]
    async fn upgrade_subscription() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/a/billing/subscription/upgrade"))
            .respond_with(ok(json!({"plan_name":"ent"}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), MT::org("a"));
        let _ = c.upgrade_subscription(&UpgradeSubscriptionRequest { plan_name: "ent".into() }).await;
    }
    #[tokio::test]
    async fn change_billing_cycle() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/a/billing/subscription/change-cycle"))
            .respond_with(ok(json!({"plan_name":"pro"}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), MT::org("a"));
        let _ = c.change_billing_cycle(&ChangeBillingCycleRequest { billing_cycle: "yearly".into() }).await;
    }
    #[tokio::test]
    async fn update_auto_renew() {
        let s = MockServer::start().await;
        Mock::given(method("PUT")).and(path("/api/v1/orgs/a/billing/subscription/auto-renew"))
            .respond_with(ok(json!({"plan_name":"pro"}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), MT::org("a"));
        let _ = c.update_auto_renew(&UpdateAutoRenewRequest { auto_renew: true }).await;
    }
    #[tokio::test]
    async fn get_customer_portal() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/a/billing/customer-portal"))
            .respond_with(ok(json!({"url":"x"}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), MT::org("a"));
        let _ = c.get_customer_portal(&CustomerPortalRequest { return_url: Some("x".into()) }).await;
    }
    #[tokio::test]
    async fn get_billing_deployment_info() {
        let s = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/orgs/a/billing/deployment"))
            .respond_with(ok(json!({"deployment_type":"cloud"}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), MT::org("a"));
        let _ = c.get_billing_deployment_info().await;
    }
}
