/// Backend-agnostic persistence used by `AuthManager` to read / write the
/// per-base_url session blob. Three implementations exist: WebLocalStorage
/// (wasm), FileStorage (node-bridge), KeychainStorage via StorageBridge
/// (iOS). Auth never enumerates keys — it reads/writes/removes a single
/// known key per `base_url`, so a `clear()` method on this trait would be
/// dead weight.
pub trait PersistentStorage: Send + Sync {
    fn get(&self, key: &str) -> Option<String>;
    fn set(&self, key: &str, value: &str);
    fn remove(&self, key: &str);
}
