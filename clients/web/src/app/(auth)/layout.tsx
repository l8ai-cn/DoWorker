import { WasmProvider } from "@/providers/WasmProvider";

// Auth route group — login / register / OAuth callback / verify-email / etc.
// These routes need wasm because the form submit / OAuth handlers go through
// AuthManager. Marketing routes outside this group (and outside (dashboard))
// should NOT load wasm; see app/layout.tsx (root) for the partition.
export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <WasmProvider>{children}</WasmProvider>;
}
