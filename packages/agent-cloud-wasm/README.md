# Agent Cloud WASM

Cargo + `wasm-pack` build of `clients/core/crates/wasm`, published as npm package `agent-cloud-wasm`.

## Build

```bash
export PATH="$HOME/.local/bin:$HOME/.cargo/bin:$PATH"
bash scripts/build-wasm.sh
```

Requires `protoc` (≥29), `wasm-pack`, `rustc` with `wasm32-unknown-unknown`, and `cargo`.
