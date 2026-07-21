/// <reference types="vite/client" />

interface ImportMetaEnv {
  /**
   * Base URL of an OTLP/HTTP collector for browser tracing, e.g.
   * `http://localhost:4318`. When unset, browser tracing is disabled.
   */
  readonly VITE_OTEL_EXPORTER_OTLP_ENDPOINT?: string;
  /** Service name reported for browser spans. Defaults to `omni-web`. */
  readonly VITE_OTEL_SERVICE_NAME?: string;
  readonly VITE_PREVIEW_PUBLIC_ORIGIN?: string;
  readonly VITE_DO_WORKER_API_URL?: string;
  readonly VITE_AGENTCLOUD_API_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
