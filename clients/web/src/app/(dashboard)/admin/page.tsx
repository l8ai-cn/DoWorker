import { redirect } from "next/navigation";

// Admin root lands on the first migrated page. More pages join as Phase 3
// progresses; this redirect becomes a dashboard once there are several.
export default function AdminIndexPage() {
  redirect("/admin/audit-logs");
}
