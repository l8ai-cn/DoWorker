use agentsmesh_relay::RelayConnectionPool;
use agentsmesh_transport::runtime::{PlatformRuntime, Runtime};
use futures::stream::StreamExt;
use wasm_bindgen::prelude::*;

use crate::js_bridge::{
    make_acp_callback, make_disconnect_callback, make_output_callback, make_status_callback,
};

#[wasm_bindgen]
pub struct WasmRelayManager {
    pool: RelayConnectionPool<PlatformRuntime>,
}

#[wasm_bindgen]
impl WasmRelayManager {
    #[wasm_bindgen(constructor)]
    pub fn new() -> Self {
        let (pool, mut rx) = RelayConnectionPool::with_runtime(PlatformRuntime);
        let pool_ref = pool.clone();
        PlatformRuntime.spawn(Box::pin(async move {
            while let Some((pod_key, sub_id)) = rx.next().await {
                pool_ref.unsubscribe(&pod_key, &sub_id).await;
            }
        }));
        Self { pool }
    }

    pub async fn subscribe(
        &self,
        pod_key: String,
        subscription_id: String,
        relay_url: String,
        token: String,
        callback: js_sys::Function,
    ) -> Result<(), String> {
        let output_cb = make_output_callback(callback);
        self.pool
            .subscribe(&pod_key, &subscription_id, &relay_url, &token, output_cb)
            .await;
        Ok(())
    }

    pub async fn unsubscribe(&self, pod_key: String, subscription_id: String) {
        self.pool.unsubscribe(&pod_key, &subscription_id).await;
    }

    pub async fn send(&self, pod_key: String, data: String) {
        self.pool.send(&pod_key, &data).await;
    }

    pub async fn send_resize(&self, pod_key: String, cols: u16, rows: u16) {
        self.pool.send_resize(&pod_key, cols, rows).await;
    }

    pub async fn force_resize(&self, pod_key: String, cols: u16, rows: u16) {
        self.pool.force_resize(&pod_key, cols, rows).await;
    }

    pub async fn acquire_control(
        &self,
        pod_key: String,
        client_label: String,
    ) -> Result<(), String> {
        self.pool
            .acquire_control(&pod_key, &client_label)
            .await
            .map_err(agentsmesh_services::wire)
    }

    pub async fn renew_control(&self, pod_key: String, lease_id: String) -> Result<(), String> {
        self.pool
            .renew_control(&pod_key, &lease_id)
            .await
            .map_err(agentsmesh_services::wire)
    }

    pub async fn release_control(&self, pod_key: String, lease_id: String) -> Result<(), String> {
        self.pool
            .release_control(&pod_key, &lease_id)
            .await
            .map_err(agentsmesh_services::wire)
    }

    pub async fn send_acp_command(&self, pod_key: String, command: String) -> Result<(), String> {
        let val: serde_json::Value =
            serde_json::from_str(&command).map_err(agentsmesh_services::wire)?;
        self.pool
            .send_acp_command(&pod_key, &val)
            .await
            .map_err(agentsmesh_services::wire)
    }

    pub async fn on_status_change(&self, pod_key: String, callback: js_sys::Function) {
        let cb = make_status_callback(callback);
        self.pool.on_status_change(&pod_key, cb).await;
    }

    pub async fn set_status_listener(
        &self,
        pod_key: String,
        listener_id: String,
        callback: js_sys::Function,
    ) {
        let cb = make_status_callback(callback);
        self.pool
            .set_status_listener(&pod_key, &listener_id, cb)
            .await;
    }

    pub fn remove_status_listener(&self, pod_key: String, listener_id: String) {
        self.pool.remove_status_listener(&pod_key, &listener_id);
    }

    pub async fn on_acp_message(&self, pod_key: String, callback: js_sys::Function) {
        let cb = make_acp_callback(callback);
        self.pool.on_acp_message(&pod_key, cb).await;
    }

    pub async fn set_acp_listener(
        &self,
        pod_key: String,
        listener_id: String,
        callback: js_sys::Function,
    ) {
        let cb = make_acp_callback(callback);
        self.pool.set_acp_listener(&pod_key, &listener_id, cb).await;
    }

    pub fn remove_acp_listener(&self, pod_key: String, listener_id: String) {
        self.pool.remove_acp_listener(&pod_key, &listener_id);
    }

    /// Register the single pod-disconnected sink — `(podKey: string) => void`.
    /// The relay adapter clears its register-once guard so the next subscribe
    /// re-registers status/ACP listeners.
    pub fn on_pod_disconnected(&self, callback: js_sys::Function) {
        self.pool
            .set_on_pod_disconnected(make_disconnect_callback(callback));
    }

    pub async fn get_status(&self, pod_key: String) -> String {
        self.pool.get_status(&pod_key).await.to_string()
    }

    pub async fn is_runner_disconnected(&self, pod_key: String) -> bool {
        self.pool.is_runner_disconnected(&pod_key).await
    }

    pub async fn get_pod_size(&self, pod_key: String) -> JsValue {
        match self.pool.get_pod_size(&pod_key).await {
            Some((cols, rows)) => {
                let obj = js_sys::Object::new();
                let _ = js_sys::Reflect::set(&obj, &"cols".into(), &cols.into());
                let _ = js_sys::Reflect::set(&obj, &"rows".into(), &rows.into());
                obj.into()
            }
            None => JsValue::NULL,
        }
    }

    pub async fn get_control_lease(&self, pod_key: String) -> JsValue {
        let lease = self.pool.get_control_lease(&pod_key).await;
        let obj = js_sys::Object::new();
        let _ = js_sys::Reflect::set(&obj, &"status".into(), &lease.state.as_str().into());
        if let Some(lease_id) = lease.lease_id {
            let _ = js_sys::Reflect::set(&obj, &"leaseId".into(), &lease_id.into());
        }
        if let Some(expires_at) = lease.expires_at {
            let _ = js_sys::Reflect::set(&obj, &"expiresAt".into(), &(expires_at as f64).into());
        }
        obj.into()
    }

    pub async fn disconnect(&self, pod_key: String) {
        self.pool.disconnect(&pod_key).await;
    }

    pub async fn disconnect_all(&self) {
        self.pool.disconnect_all().await;
    }
}

impl Default for WasmRelayManager {
    fn default() -> Self {
        Self::new()
    }
}
