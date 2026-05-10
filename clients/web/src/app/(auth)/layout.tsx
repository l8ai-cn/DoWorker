import { AuthBootstrap } from "@/components/auth/AuthBootstrap";

// Auth route group — login / register / OAuth callback / verify-email /
// onboarding / runners-authorize / invite. Some pages here are anonymous
// (login, register), others require an already-authenticated user
// (onboarding, runners/authorize, invite acceptance flow). Wrapping the
// whole group in AuthBootstrap means a deep-link / new-window entry
// rehydrates the session from localStorage before page-level guards run;
// for anonymous pages the bootstrap result is `Anonymous` and behaves as
// before. Marketing routes outside this group should NOT load wasm; see
// app/layout.tsx (root) for the partition.
export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <AuthBootstrap>{children}</AuthBootstrap>;
}
