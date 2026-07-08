// Admin-scoped call-site shapes (snake_case). Kept in clients/web so admin
// pages don't depend on the web-admin package during the merge. Only the
// audit-log surface is migrated in Phase 3 Step 1; add more as pages move.

export interface AdminPaginated<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface AuditLog {
  id: number;
  admin_user_id: number;
  action: string;
  target_type: string;
  target_id: number;
  old_data: string | null;
  new_data: string | null;
  ip_address: string | null;
  user_agent: string | null;
  created_at: string;
  admin_user?: {
    id: number;
    email: string;
    username: string;
    name: string | null;
    avatar_url: string | null;
  };
}

export interface AuditLogListParams {
  admin_user_id?: number;
  action?: string;
  target_type?: string;
  target_id?: number;
  start_time?: string;
  end_time?: string;
  page?: number;
  page_size?: number;
}
