#[cfg(test)]
mod bootstrap_tests {
    use wiremock::matchers::{method, path};
    use wiremock::{Mock, MockServer, ResponseTemplate};

    use crate::bootstrap::{BootstrapCleanupReason, BootstrapResult};
    use crate::manager::{now_unix_secs, AuthManager};
    use crate::state::{session_storage_key, PersistedSession, SCHEMA_VERSION};
    use crate::storage::PersistentStorage;
    use crate::test_support::InMemoryStorage;

    fn me_json() -> serde_json::Value {
        serde_json::json!({
            "user": {
                "id": 1, "email": "dev@test.com",
                "username": "dev", "name": "Dev User", "avatar_url": null
            }
        })
    }

    fn orgs_json() -> serde_json::Value {
        serde_json::json!({
            "organizations": [
                {"id": 10, "name": "Org A", "slug": "org-a"},
                {"id": 11, "name": "Org B", "slug": "org-b"}
            ]
        })
    }

    fn install_session(storage: &InMemoryStorage, base_url: &str, session: PersistedSession) {
        let key = session_storage_key(base_url);
        let json = serde_json::to_string(&session).unwrap();
        storage.set(&key, &json);
    }

    fn fresh_session(base_url: &str) -> PersistedSession {
        PersistedSession {
            access_token: "live-tok".into(),
            refresh_token: "live-ref".into(),
            expires_at: now_unix_secs() + 3600,
            base_url: base_url.into(),
            current_org_slug: None,
            schema_version: SCHEMA_VERSION,
        }
    }

    #[tokio::test]
    async fn happy_path_authenticates_and_picks_first_org() {
        let server = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/users/me"))
            .respond_with(ResponseTemplate::new(200).set_body_json(me_json()))
            .mount(&server).await;
        Mock::given(method("GET")).and(path("/api/v1/users/me/organizations"))
            .respond_with(ResponseTemplate::new(200).set_body_json(orgs_json()))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        install_session(&storage, &server.uri(), fresh_session(&server.uri()));
        let manager = AuthManager::new(server.uri(), storage);

        match manager.bootstrap().await {
            BootstrapResult::Authenticated { user, current_org } => {
                assert_eq!(user.email, "dev@test.com");
                assert_eq!(current_org.unwrap().slug, "org-a");
            }
            other => panic!("expected Authenticated, got {other:?}"),
        }
    }

    #[tokio::test]
    async fn happy_path_honors_persisted_org_slug() {
        let server = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/users/me"))
            .respond_with(ResponseTemplate::new(200).set_body_json(me_json()))
            .mount(&server).await;
        Mock::given(method("GET")).and(path("/api/v1/users/me/organizations"))
            .respond_with(ResponseTemplate::new(200).set_body_json(orgs_json()))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        let mut session = fresh_session(&server.uri());
        session.current_org_slug = Some("org-b".into());
        install_session(&storage, &server.uri(), session);
        let manager = AuthManager::new(server.uri(), storage);

        match manager.bootstrap().await {
            BootstrapResult::Authenticated { current_org, .. } => {
                assert_eq!(current_org.unwrap().slug, "org-b");
            }
            other => panic!("expected Authenticated, got {other:?}"),
        }
    }

    #[tokio::test]
    async fn empty_storage_is_anonymous() {
        let storage = InMemoryStorage::new();
        let manager = AuthManager::new("http://localhost".into(), storage);
        assert!(matches!(manager.bootstrap().await, BootstrapResult::Anonymous));
    }

    #[tokio::test]
    async fn base_url_mismatch_cleans() {
        let storage = InMemoryStorage::new();
        let mut session = fresh_session("http://manager-base");
        session.base_url = "http://other-server".into();
        // 写到 manager-base 的 namespace 下，但 session.base_url 是 other-server
        install_session(&storage, "http://manager-base", session);
        let storage_arc = storage.clone();
        let manager = AuthManager::new("http://manager-base".into(), storage_arc);

        match manager.bootstrap().await {
            BootstrapResult::AnonymousAfterCleanup { reason } => {
                assert_eq!(reason, BootstrapCleanupReason::BaseUrlMismatch);
            }
            other => panic!("expected cleanup, got {other:?}"),
        }
        // session key 已被清
        assert!(!storage
            .snapshot()
            .keys()
            .any(|k| k == &session_storage_key("http://manager-base")));
    }

    #[tokio::test]
    async fn storage_corrupt_cleans() {
        let storage = InMemoryStorage::new();
        let key = session_storage_key("http://localhost");
        storage.set(&key, "not-valid-json{{{");
        let manager = AuthManager::new("http://localhost".into(), storage.clone());

        match manager.bootstrap().await {
            BootstrapResult::AnonymousAfterCleanup { reason } => {
                assert_eq!(reason, BootstrapCleanupReason::StorageCorrupt);
            }
            other => panic!("expected cleanup, got {other:?}"),
        }
        assert!(storage.get(&key).is_none());
    }

    #[tokio::test]
    async fn legacy_key_alone_is_purged() {
        let storage = InMemoryStorage::new();
        storage.set("agentsmesh-auth", r#"{"token":"old"}"#);
        let manager = AuthManager::new("http://localhost".into(), storage.clone());

        match manager.bootstrap().await {
            BootstrapResult::AnonymousAfterCleanup { reason } => {
                assert_eq!(reason, BootstrapCleanupReason::LegacyDataPurged);
            }
            other => panic!("expected cleanup, got {other:?}"),
        }
        assert!(storage.get("agentsmesh-auth").is_none());
    }

    #[tokio::test]
    async fn expired_token_refreshes_and_continues() {
        let server = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/auth/refresh"))
            .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!({
                "token": "new-access",
                "refresh_token": "new-refresh",
                "expires_in": 7200
            })))
            .mount(&server).await;
        Mock::given(method("GET")).and(path("/api/v1/users/me"))
            .respond_with(ResponseTemplate::new(200).set_body_json(me_json()))
            .mount(&server).await;
        Mock::given(method("GET")).and(path("/api/v1/users/me/organizations"))
            .respond_with(ResponseTemplate::new(200).set_body_json(orgs_json()))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        let mut session = fresh_session(&server.uri());
        session.expires_at = now_unix_secs() - 100; // 已过期
        install_session(&storage, &server.uri(), session);
        let manager = AuthManager::new(server.uri(), storage);

        match manager.bootstrap().await {
            BootstrapResult::Authenticated { .. } => {}
            other => panic!("expected Authenticated, got {other:?}"),
        }
    }

    #[tokio::test]
    async fn expired_token_refresh_failure_cleans() {
        let server = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/auth/refresh"))
            .respond_with(ResponseTemplate::new(401).set_body_json(serde_json::json!({
                "message": "refresh expired"
            })))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        let mut session = fresh_session(&server.uri());
        session.expires_at = now_unix_secs() - 100;
        install_session(&storage, &server.uri(), session);
        let manager = AuthManager::new(server.uri(), storage.clone());

        match manager.bootstrap().await {
            BootstrapResult::AnonymousAfterCleanup { reason } => {
                assert_eq!(reason, BootstrapCleanupReason::TokenExpiredAndRefreshFailed);
            }
            other => panic!("expected cleanup, got {other:?}"),
        }
        assert!(storage.get(&session_storage_key(&server.uri())).is_none());
    }

    #[tokio::test]
    async fn identity_401_cleans() {
        let server = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/users/me"))
            .respond_with(ResponseTemplate::new(401).set_body_json(serde_json::json!({
                "message": "invalid token"
            })))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        install_session(&storage, &server.uri(), fresh_session(&server.uri()));
        let manager = AuthManager::new(server.uri(), storage.clone());

        match manager.bootstrap().await {
            BootstrapResult::AnonymousAfterCleanup { reason } => {
                assert_eq!(reason, BootstrapCleanupReason::UnauthorizedFromIdentityCall);
            }
            other => panic!("expected cleanup, got {other:?}"),
        }
        assert!(storage.get(&session_storage_key(&server.uri())).is_none());
    }

    #[tokio::test]
    async fn identity_5xx_keeps_session_returns_anonymous() {
        let server = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/users/me"))
            .respond_with(ResponseTemplate::new(503).set_body_json(serde_json::json!({
                "message": "service unavailable"
            })))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        install_session(&storage, &server.uri(), fresh_session(&server.uri()));
        let manager = AuthManager::new(server.uri(), storage.clone());

        match manager.bootstrap().await {
            BootstrapResult::Anonymous => {}
            other => panic!("expected transient anonymous, got {other:?}"),
        }
        // session 应保留以便重试
        assert!(storage.get(&session_storage_key(&server.uri())).is_some());
    }

    #[tokio::test]
    async fn organizations_failure_falls_back_to_no_current_org() {
        let server = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/users/me"))
            .respond_with(ResponseTemplate::new(200).set_body_json(me_json()))
            .mount(&server).await;
        Mock::given(method("GET")).and(path("/api/v1/users/me/organizations"))
            .respond_with(ResponseTemplate::new(500).set_body_json(serde_json::json!({
                "message": "internal"
            })))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        install_session(&storage, &server.uri(), fresh_session(&server.uri()));
        let manager = AuthManager::new(server.uri(), storage);

        match manager.bootstrap().await {
            BootstrapResult::Authenticated { current_org, .. } => {
                assert!(current_org.is_none());
            }
            other => panic!("expected Authenticated with no org, got {other:?}"),
        }
    }

    #[tokio::test]
    async fn organizations_401_cleans_session() {
        // R3 regression guard: a 401 from /users/me/organizations indicates
        // the token was just revoked (after /users/me succeeded). It MUST
        // run cleanup — silently falling back to empty orgs would leave a
        // stale session on disk + dashboard with empty data, so next
        // business request 401's into a refresh loop.
        let server = MockServer::start().await;
        Mock::given(method("GET")).and(path("/api/v1/users/me"))
            .respond_with(ResponseTemplate::new(200).set_body_json(me_json()))
            .mount(&server).await;
        Mock::given(method("GET")).and(path("/api/v1/users/me/organizations"))
            .respond_with(ResponseTemplate::new(401).set_body_json(serde_json::json!({
                "message": "token revoked"
            })))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        install_session(&storage, &server.uri(), fresh_session(&server.uri()));
        let manager = AuthManager::new(server.uri(), storage.clone());

        match manager.bootstrap().await {
            BootstrapResult::AnonymousAfterCleanup { reason } => {
                assert_eq!(reason, BootstrapCleanupReason::UnauthorizedFromIdentityCall);
            }
            other => panic!("expected cleanup, got {other:?}"),
        }
        // session file must be wiped on 401-during-orgs cleanup
        assert!(storage.get(&session_storage_key(&server.uri())).is_none());
    }

    #[tokio::test]
    async fn isolated_namespaces_for_different_base_urls() {
        let storage = InMemoryStorage::new();
        // 两个 base_url 各自独立 session，互不覆盖
        install_session(&storage, "http://server-a", {
            let mut s = fresh_session("http://server-a");
            s.access_token = "tok-a".into();
            s
        });
        install_session(&storage, "http://server-b", {
            let mut s = fresh_session("http://server-b");
            s.access_token = "tok-b".into();
            s
        });

        let dump = storage.snapshot();
        assert_eq!(dump.len(), 2);
        let key_a = session_storage_key("http://server-a");
        let key_b = session_storage_key("http://server-b");
        assert!(dump.contains_key(&key_a));
        assert!(dump.contains_key(&key_b));
        assert_ne!(key_a, key_b);
    }

    #[tokio::test]
    async fn persisted_session_does_not_contain_user_object() {
        // I2 不变量校验：不可以把 user 完整对象写到 disk
        let storage = InMemoryStorage::new();
        install_session(&storage, "http://localhost", fresh_session("http://localhost"));

        let key = session_storage_key("http://localhost");
        let raw = storage.get(&key).unwrap();
        assert!(!raw.contains("\"email\""), "session blob leaked email field");
        assert!(!raw.contains("\"username\""), "session blob leaked username field");
        assert!(!raw.contains("\"name\""), "session blob leaked name field");
        assert!(!raw.contains("\"avatar_url\""), "session blob leaked avatar_url field");
    }
}
