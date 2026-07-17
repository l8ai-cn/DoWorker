#![cfg(target_arch = "wasm32")]

use std::cell::RefCell;
use std::rc::Rc;

use bytes::Bytes;
use futures::channel::mpsc;
use futures::stream::Stream;
use wasm_bindgen::{JsCast, JsValue};
use wasm_bindgen_futures::{spawn_local, JsFuture};
use web_sys::ReadableStreamDefaultReader;

use crate::error::ApiError;

pub(crate) fn read_chunks(
    reader: ReadableStreamDefaultReader,
) -> impl Stream<Item = Result<Bytes, ApiError>> {
    let (sender, receiver) = mpsc::unbounded();
    let reader = Rc::new(RefCell::new(reader));
    spawn_local(async move {
        loop {
            let value = match JsFuture::from(reader.borrow().read()).await {
                Ok(value) => value,
                Err(error) => {
                    let _ = sender.unbounded_send(Err(js_error_value("reader.read", error)));
                    return;
                }
            };
            let done = match stream_field(&value, "done").and_then(|value| {
                value
                    .as_bool()
                    .ok_or_else(|| ApiError::Decode("wasm reader.done is not boolean".into()))
            }) {
                Ok(done) => done,
                Err(error) => {
                    let _ = sender.unbounded_send(Err(error));
                    return;
                }
            };
            if done {
                return;
            }
            let chunk = match stream_field(&value, "value") {
                Ok(chunk) => chunk,
                Err(error) => {
                    let _ = sender.unbounded_send(Err(error));
                    return;
                }
            };
            let array: js_sys::Uint8Array = match chunk.dyn_into() {
                Ok(array) => array,
                Err(_) => {
                    let _ = sender.unbounded_send(Err(ApiError::Decode(
                        "wasm reader.value is not Uint8Array".into(),
                    )));
                    return;
                }
            };
            let mut bytes = vec![0; array.length() as usize];
            array.copy_to(&mut bytes);
            if sender.unbounded_send(Ok(Bytes::from(bytes))).is_err() {
                return;
            }
        }
    });
    receiver
}

fn stream_field(value: &JsValue, name: &str) -> Result<JsValue, ApiError> {
    js_sys::Reflect::get(value, &JsValue::from_str(name))
        .map_err(|error| js_error_value(&format!("reader.read {name}"), error))
}

pub(crate) fn js_error(context: &'static str) -> impl FnOnce(JsValue) -> ApiError + 'static {
    move |error| js_error_value(context, error)
}

fn js_error_value(context: &str, error: JsValue) -> ApiError {
    ApiError::Decode(format!("wasm {context}: {error:?}"))
}
