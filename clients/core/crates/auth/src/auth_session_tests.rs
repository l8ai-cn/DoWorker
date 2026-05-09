#[cfg(test)]
mod auth_session_tests {
    use agentsmesh_types::RegisterRequest;
    use wiremock::matchers::{method, path};
    use wiremock::{Mock, MockServer, ResponseTemplate};

    use crate::manager::AuthManager;
    use crate::state::session_storage_key;
    use crate::storage::PersistentStorage;
    use crate::test_support::InMemoryStorage;

    fn session_json() -> serde_json::Value {
        serde_json::json!({
            "token": "access-tok",
            "refresh_token": "refresh-tok",
            "user": {
                "id": 1, "email": "dev@test.com",
                "username": "dev", "name": "Dev User", "avatar_url": null
            },
            "expires_in": 3600
        })
    }

    #[tokio::test]
    async fn login_success() {
        let server = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/auth/login"))
            .respond_with(ResponseTemplate::new(200).set_body_json(session_json()))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        let manager = AuthManager::new(server.uri(), storage.clone());

        let session = manager.login("dev@test.com", "pass").await.unwrap();
        assert_eq!(session.token, "access-tok");
        assert_eq!(session.user.email, "dev@test.com");
        assert!(manager.is_authenticated());
        let key = session_storage_key(&server.uri());
        assert!(storage.get(&key).is_some());
        // legacy key 不应被写
        assert!(storage.get("agentsmesh-auth").is_none());
    }

    #[tokio::test]
    async fn login_failure_401() {
        let server = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/auth/login"))
            .respond_with(
                ResponseTemplate::new(401)
                    .set_body_json(serde_json::json!({"message": "invalid credentials"})),
            )
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        let manager = AuthManager::new(server.uri(), storage);

        let err = manager.login("bad@test.com", "wrong").await.unwrap_err();
        match err {
            crate::AuthError::Server { status, message, .. } => {
                assert_eq!(status, 401);
                assert!(message.contains("invalid credentials"));
            }
            _ => panic!("expected Server error, got: {err:?}"),
        }
    }

    #[tokio::test]
    async fn register_success() {
        let server = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/auth/register"))
            .respond_with(ResponseTemplate::new(200).set_body_json(session_json()))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        let manager = AuthManager::new(server.uri(), storage);

        let req = RegisterRequest {
            name: "Dev".into(), email: "dev@test.com".into(),
            username: "dev".into(), password: "pass123".into(),
        };
        let session = manager.register(&req).await.unwrap();
        assert_eq!(session.user.username, "dev");
        assert!(manager.is_authenticated());
    }

    #[tokio::test]
    async fn register_failure() {
        let server = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/auth/register"))
            .respond_with(
                ResponseTemplate::new(409)
                    .set_body_json(serde_json::json!({"message": "email taken"})),
            )
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        let manager = AuthManager::new(server.uri(), storage);

        let req = RegisterRequest {
            name: "Dev".into(), email: "dup@test.com".into(),
            username: "dev".into(), password: "pass123".into(),
        };
        let err = manager.register(&req).await.unwrap_err();
        assert!(matches!(err, crate::AuthError::Server { status: 409, .. }));
    }

    #[tokio::test]
    async fn logout_success() {
        let server = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/auth/login"))
            .respond_with(ResponseTemplate::new(200).set_body_json(session_json()))
            .mount(&server).await;
        Mock::given(method("POST")).and(path("/api/v1/auth/logout"))
            .respond_with(ResponseTemplate::new(200))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        let manager = AuthManager::new(server.uri(), storage.clone());

        manager.login("dev@test.com", "pass").await.unwrap();
        assert!(manager.is_authenticated());

        manager.logout().await.unwrap();
        assert!(!manager.is_authenticated());
        assert!(manager.current_user().is_none());
        let key = session_storage_key(&server.uri());
        assert!(storage.get(&key).is_none());
    }

    #[tokio::test]
    async fn logout_without_auth() {
        let storage = InMemoryStorage::new();
        let manager = AuthManager::new("http://localhost".into(), storage);
        let err = manager.logout().await.unwrap_err();
        assert!(matches!(err, crate::AuthError::NotAuthenticated));
    }

    #[tokio::test]
    async fn logout_server_5xx_still_clears_local() {
        // Plan I3 invariant: server-side logout failure must NOT leave
        // the renderer with `is_authenticated() == true`. We login, then
        // mount a 500 on /auth/logout, and assert local state is dropped
        // regardless.
        let server = MockServer::start().await;
        Mock::given(method("POST")).and(path("/api/v1/auth/login"))
            .respond_with(ResponseTemplate::new(200).set_body_json(session_json()))
            .mount(&server).await;
        Mock::given(method("POST")).and(path("/api/v1/auth/logout"))
            .respond_with(ResponseTemplate::new(500))
            .mount(&server).await;

        let storage = InMemoryStorage::new();
        let manager = AuthManager::new(server.uri(), storage.clone());

        manager.login("dev@test.com", "pass").await.unwrap();
        assert!(manager.is_authenticated());

        // Returns Ok despite 5xx — local state is the contract, not the wire.
        manager.logout().await.unwrap();
        assert!(!manager.is_authenticated());
        assert!(manager.current_user().is_none());
        let key = session_storage_key(&server.uri());
        assert!(storage.get(&key).is_none());
    }

    // restore_session_* tests removed — bootstrap_tests.rs covers the
    // same paths (empty / corrupt / base_url mismatch / legacy purge)
    // through the new bootstrap protocol, which is now the only public
    // hydrate entry point.
}
