import type { Metadata } from "next";
import DashboardShell from "./DashboardShell";
import { WasmProvider } from "@/providers/WasmProvider";
import { PostHogIdentify } from "@/providers/PostHogProvider";

export const metadata: Metadata = {
  robots: { index: false, follow: false },
};

// Dashboard route group — wasm-bound. WasmProvider is mounted here (not in
// the root layout) so marketing pages stay free of the 21MB wasm load.
// PostHogIdentify is also scoped here because it depends on auth hooks.
export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <WasmProvider>
      <PostHogIdentify />
      <DashboardShell>{children}</DashboardShell>
    </WasmProvider>
  );
}
