#[cfg(test)]
mod api_ticket_tests {
    use std::sync::{Arc, Mutex};

    use serde_json::json;
    use wiremock::matchers::{method, path, query_param};
    use wiremock::{Mock, MockServer, ResponseTemplate};

    use crate::{ApiClient, AuthTokenStore};

    struct Tok(Mutex<Option<String>>);
    impl Tok {
        fn org(s: &str) -> Arc<Self> { Arc::new(Self(Mutex::new(Some(s.into())))) }
    }
    impl AuthTokenStore for Tok {
        fn get_token(&self) -> Option<String> { Some("tok".into()) }
        fn get_refresh_token(&self) -> Option<String> { None }
        fn set_tokens(&self, _t: String, _r: String, _e: Option<i64>) {}
        fn clear_tokens(&self) {}
        fn get_current_org_slug(&self) -> Option<String> { self.0.lock().unwrap().clone() }
    }

    fn ok(b: serde_json::Value) -> ResponseTemplate {
        ResponseTemplate::new(200).set_body_json(b)
    }

    fn ticket_json(slug: &str, title: &str) -> serde_json::Value {
        json!({"slug": slug, "title": title, "status": "open", "priority": "medium"})
    }

    #[tokio::test]
    async fn get_ticket() {
        let s = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/orgs/acme/tickets/TK-1"))
            .respond_with(ok(ticket_json("TK-1", "Bug")))
            .expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let r = c.get_ticket("TK-1").await.unwrap();
        assert_eq!(r.slug, "TK-1");
    }

    #[tokio::test]
    async fn create_ticket() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/acme/tickets"))
            .respond_with(ok(ticket_json("TK-2", "New")))
            .expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::CreateTicketRequest {
            title: "New".into(), content: None, priority: None,
            severity: None, estimate: None, repository_id: None,
            assignee_ids: None, labels: None, parent_slug: None,
        };
        let r = c.create_ticket(&data).await.unwrap();
        assert_eq!(r.title, "New");
    }

    #[tokio::test]
    async fn update_ticket() {
        let s = MockServer::start().await;
        Mock::given(method("PUT")).and(path("/api/v1/orgs/acme/tickets/TK-1"))
            .respond_with(ok(ticket_json("TK-1", "Upd")))
            .expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::UpdateTicketRequest {
            title: Some("Upd".into()), content: None, priority: None,
            severity: None, estimate: None, repository_id: None,
        };
        let _ = c.update_ticket("TK-1", &data).await.unwrap();
    }

    #[tokio::test]
    async fn delete_ticket() {
        let s = MockServer::start().await;
        Mock::given(method("DELETE")).and(path("/api/v1/orgs/acme/tickets/TK-1"))
            .respond_with(ok(json!({}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.delete_ticket("TK-1").await.unwrap();
    }

    #[tokio::test]
    async fn update_ticket_status() {
        let s = MockServer::start().await;
        Mock::given(method("PATCH")).and(path("/api/v1/orgs/acme/tickets/TK-1/status"))
            .respond_with(ok(ticket_json("TK-1", "Bug")))
            .expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::UpdateTicketStatusRequest { status: agentsmesh_types::TicketStatus::Done };
        let _ = c.update_ticket_status("TK-1", &data).await.unwrap();
    }

    #[tokio::test]
    async fn get_active_tickets() {
        let s = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/orgs/acme/tickets/active"))
            .and(query_param("limit", "5"))
            .respond_with(ok(json!({"tickets":[]}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.get_active_tickets(Some(5)).await.unwrap();
    }

    #[tokio::test]
    async fn get_sub_tickets() {
        let s = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/orgs/acme/tickets/TK-1/sub-tickets"))
            .respond_with(ok(json!({"tickets":[]}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.get_sub_tickets("TK-1").await.unwrap();
    }

    #[tokio::test]
    async fn list_labels() {
        let s = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/orgs/acme/labels"))
            .and(query_param("repository_id", "10"))
            .respond_with(ok(json!({"labels":[]}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.list_labels(Some(10)).await.unwrap();
    }

    #[tokio::test]
    async fn create_label() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/acme/labels"))
            .respond_with(ok(json!({"id":1,"name":"bug","color":"#f00"})))
            .expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::CreateLabelRequest {
            name: "bug".into(), color: "#f00".into(),
        };
        let r = c.create_label(&data).await.unwrap();
        assert_eq!(r.name, "bug");
    }

    #[tokio::test]
    async fn update_label() {
        let s = MockServer::start().await;
        Mock::given(method("PUT")).and(path("/api/v1/orgs/acme/labels/1"))
            .respond_with(ok(json!({"id":1,"name":"feat","color":"#0f0"})))
            .expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::UpdateLabelRequest {
            name: Some("feat".into()), color: Some("#0f0".into()),
        };
        let _ = c.update_label(1, &data).await.unwrap();
    }

    #[tokio::test]
    async fn delete_label() {
        let s = MockServer::start().await;
        Mock::given(method("DELETE")).and(path("/api/v1/orgs/acme/labels/1"))
            .respond_with(ok(json!({}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.delete_label(1).await.unwrap();
    }

    #[tokio::test]
    async fn add_ticket_assignee() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/acme/tickets/TK-1/assignees"))
            .respond_with(ok(json!({}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::AddAssigneeRequest { user_id: 42 };
        let _ = c.add_ticket_assignee("TK-1", &data).await.unwrap();
    }

    #[tokio::test]
    async fn remove_ticket_assignee() {
        let s = MockServer::start().await;
        Mock::given(method("DELETE"))
            .and(path("/api/v1/orgs/acme/tickets/TK-1/assignees/42"))
            .respond_with(ok(json!({}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.remove_ticket_assignee("TK-1", 42).await.unwrap();
    }

    #[tokio::test]
    async fn get_ticket_pods() {
        let s = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/orgs/acme/tickets/TK-1/pods"))
            .and(query_param("active", "true"))
            .respond_with(ok(json!({"pods":[]}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.get_ticket_pods("TK-1", Some(true)).await.unwrap();
    }

    #[tokio::test]
    async fn batch_get_ticket_pods() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/acme/tickets/batch-pods"))
            .respond_with(ok(json!({}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::BatchPodRequest {
            ticket_slugs: vec!["TK-1".into()],
        };
        let _ = c.batch_get_ticket_pods(&data).await.unwrap();
    }

    #[tokio::test]
    async fn add_ticket_label() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/acme/tickets/TK-1/labels"))
            .respond_with(ok(json!({}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::AddTicketLabelRequest { label_id: 3 };
        let _ = c.add_ticket_label("TK-1", &data).await.unwrap();
    }

    #[tokio::test]
    async fn remove_ticket_label() {
        let s = MockServer::start().await;
        Mock::given(method("DELETE")).and(path("/api/v1/orgs/acme/tickets/TK-1/labels/3"))
            .respond_with(ok(json!({}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.remove_ticket_label("TK-1", 3).await.unwrap();
    }

    #[tokio::test]
    async fn create_ticket_relation() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/acme/tickets/TK-1/relations"))
            .respond_with(ok(json!({"id":1,"source_slug":"TK-1","target_slug":"TK-2"})))
            .expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::CreateTicketRelationRequest {
            target_slug: "TK-2".into(), relation_type: "blocks".into(),
        };
        let _ = c.create_ticket_relation("TK-1", &data).await.unwrap();
    }

    #[tokio::test]
    async fn delete_ticket_relation() {
        let s = MockServer::start().await;
        Mock::given(method("DELETE"))
            .and(path("/api/v1/orgs/acme/tickets/TK-1/relations/7"))
            .respond_with(ok(json!({}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.delete_ticket_relation("TK-1", 7).await.unwrap();
    }

    #[tokio::test]
    async fn link_ticket_commit() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/acme/tickets/TK-1/commits"))
            .respond_with(ok(json!({"id":1,"commit_sha":"abc123"})))
            .expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::LinkTicketCommitRequest {
            commit_sha: "abc123".into(), commit_message: Some("fix".into()),
            commit_url: None, author_name: None, author_email: None, committed_at: None,
        };
        let _ = c.link_ticket_commit("TK-1", &data).await.unwrap();
    }

    #[tokio::test]
    async fn unlink_ticket_commit() {
        let s = MockServer::start().await;
        Mock::given(method("DELETE")).and(path("/api/v1/orgs/acme/tickets/TK-1/commits/5"))
            .respond_with(ok(json!({}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.unlink_ticket_commit("TK-1", 5).await.unwrap();
    }

    #[tokio::test]
    async fn list_ticket_merge_requests() {
        let s = MockServer::start().await;
        Mock::given(method("GET"))
            .and(path("/api/v1/orgs/acme/tickets/TK-1/merge-requests"))
            .respond_with(ok(json!({"merge_requests":[]}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.list_ticket_merge_requests("TK-1").await.unwrap();
    }

    #[tokio::test]
    async fn create_ticket_comment() {
        let s = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/orgs/acme/tickets/TK-1/comments"))
            .respond_with(ok(json!({"id":1,"content":"nice"})))
            .expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::CreateTicketCommentRequest {
            content: "nice".into(), parent_id: None, mentions: None,
        };
        let _ = c.create_ticket_comment("TK-1", &data).await.unwrap();
    }

    #[tokio::test]
    async fn update_ticket_comment() {
        let s = MockServer::start().await;
        Mock::given(method("PUT")).and(path("/api/v1/orgs/acme/tickets/TK-1/comments/1"))
            .respond_with(ok(json!({"id":1,"content":"upd"})))
            .expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let data = agentsmesh_types::UpdateTicketCommentRequest { content: "upd".into() };
        let _ = c.update_ticket_comment("TK-1", 1, &data).await.unwrap();
    }

    #[tokio::test]
    async fn delete_ticket_comment() {
        let s = MockServer::start().await;
        Mock::given(method("DELETE"))
            .and(path("/api/v1/orgs/acme/tickets/TK-1/comments/1"))
            .respond_with(ok(json!({}))).expect(1).mount(&s).await;
        let c = ApiClient::new(s.uri(), Tok::org("acme"));
        let _ = c.delete_ticket_comment("TK-1", 1).await.unwrap();
    }
}
