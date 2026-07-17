#![cfg(target_arch = "wasm32")]

use std::cell::RefCell;
use std::rc::Rc;
use std::sync::Arc;

use agentsmesh_api_client::{AgentWorkbenchAccessScope, ApiClient, WasmAbortHandle};
use agentsmesh_state::app_state::AppState;
use futures::StreamExt;
use parking_lot::RwLock;
use wasm_bindgen::prelude::*;
use wasm_bindgen_futures::spawn_local;

use crate::agent_workbench_stream_status::AgentWorkbenchStreamStatus;
use crate::service_agent_workbench::{apply_stream_batch, session_cursor};

#[wasm_bindgen]
pub struct WasmAgentWorkbenchStream {
    abort: Rc<RefCell<Option<WasmAbortHandle>>>,
    status: Rc<RefCell<AgentWorkbenchStreamStatus>>,
    on_close: js_sys::Function,
}

#[wasm_bindgen]
impl WasmAgentWorkbenchStream {
    pub fn close(&self) {
        close_stream(&self.abort, &self.status, &self.on_close);
    }

    pub fn status(&self) -> String {
        self.status.borrow().code().into()
    }

    #[wasm_bindgen(js_name = terminalError)]
    pub fn terminal_error(&self) -> Option<String> {
        self.status.borrow().error().map(String::from)
    }
}

impl Drop for WasmAgentWorkbenchStream {
    fn drop(&mut self) {
        close_stream(&self.abort, &self.status, &self.on_close);
    }
}

pub(crate) async fn open_agent_workbench_stream(
    client: Arc<ApiClient>,
    state: Arc<RwLock<AppState>>,
    access: AgentWorkbenchAccessScope,
    session_id: String,
    replay_limit: u32,
    on_commit: js_sys::Function,
    on_error: js_sys::Function,
    on_close: js_sys::Function,
) -> Result<WasmAgentWorkbenchStream, String> {
    let cursor = session_cursor(&state, &session_id)?;
    let (stream, abort_handle) = client
        .stream_agent_workbench_session_deltas_connect_wasm(&access, cursor, replay_limit)
        .await
        .map_err(|error| error.to_wire())?;
    let abort = Rc::new(RefCell::new(Some(abort_handle)));
    let status = Rc::new(RefCell::new(AgentWorkbenchStreamStatus::default()));
    let task_abort = abort.clone();
    let task_status = status.clone();
    let task_on_close = on_close.clone();

    spawn_local(async move {
        futures::pin_mut!(stream);
        while let Some(item) = stream.next().await {
            let outcome = match item {
                Ok(batch) => apply_stream_batch(&state, &batch),
                Err(error) => Err(error.to_wire()),
            };
            match outcome {
                Ok(true) => {
                    let _ = on_commit.call0(&JsValue::NULL);
                }
                Ok(false) => {}
                Err(error) => {
                    if task_status.borrow_mut().mark_failed(error.clone()) {
                        let _ = on_error.call1(&JsValue::NULL, &JsValue::from_str(&error));
                        notify_close(&task_on_close, &task_status.borrow());
                    }
                    task_abort.borrow_mut().take();
                    return;
                }
            }
        }
        task_abort.borrow_mut().take();
        if task_status.borrow_mut().mark_remote_closed() {
            notify_close(&task_on_close, &task_status.borrow());
        }
    });

    Ok(WasmAgentWorkbenchStream {
        abort,
        status,
        on_close,
    })
}

fn close_stream(
    abort: &Rc<RefCell<Option<WasmAbortHandle>>>,
    status: &Rc<RefCell<AgentWorkbenchStreamStatus>>,
    on_close: &js_sys::Function,
) {
    if !status.borrow_mut().mark_client_closed() {
        return;
    }
    if let Some(handle) = abort.borrow_mut().take() {
        handle.abort();
    }
    notify_close(on_close, &status.borrow());
}

fn notify_close(on_close: &js_sys::Function, status: &AgentWorkbenchStreamStatus) {
    let detail = js_sys::Object::new();
    let _ = js_sys::Reflect::set(
        &detail,
        &JsValue::from_str("status"),
        &JsValue::from_str(status.code()),
    );
    let error = status
        .error()
        .map(JsValue::from_str)
        .unwrap_or(JsValue::NULL);
    let _ = js_sys::Reflect::set(&detail, &JsValue::from_str("error"), &error);
    let _ = on_close.call1(&JsValue::NULL, &detail);
}
