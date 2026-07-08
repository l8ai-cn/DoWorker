import { AdminGuard } from "@/components/admin/AdminGuard";

// System-admin console mounted under the wasm dashboard layout (Phase 3: merge
// clients/web-admin into clients/web). AdminGuard gates on is_system_admin.
export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return <AdminGuard>{children}</AdminGuard>;
}
