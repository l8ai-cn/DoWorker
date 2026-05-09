import { WasmProvider } from "@/providers/WasmProvider";

// Popout terminal — opens a pod terminal in a separate browser window. Needs
// wasm because TerminalPane consumes the relay backend and uses auth hooks
// from the same wasm-bound stores as the dashboard. URL stays /popout/...
// regardless of this layout.
export default function PopoutLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <WasmProvider>{children}</WasmProvider>;
}
