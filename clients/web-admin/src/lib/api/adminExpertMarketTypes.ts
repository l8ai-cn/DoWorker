export type ExpertMarketReleaseStatus =
  | "pending"
  | "published"
  | "rejected"
  | "withdrawn";

export interface ExpertMarketRelease {
  id: number;
  application_id: number;
  application_slug: string;
  source_expert_id: number;
  publisher_organization_id: number;
  publisher_user_id: number;
  version: number;
  status: ExpertMarketReleaseStatus;
  name: string;
  summary: string;
  description: string;
  category: string;
  icon: string;
  tags: string[];
  outcomes: string[];
  featured: boolean;
  expert_snapshot_json: string;
  worker_spec_snapshot_json: string;
  skill_dependencies_json: string;
  reviewer_user_id?: number;
  rejection_reason?: string;
  submitted_at?: string;
  reviewed_at?: string;
  published_at?: string;
  rejected_at?: string;
  withdrawn_at?: string;
  created_at: string;
}

export interface ExpertMarketReleaseListParams {
  status: ExpertMarketReleaseStatus;
  limit?: number;
  offset?: number;
}

export interface ExpertMarketReleaseList {
  items: ExpertMarketRelease[];
  total: number;
  limit: number;
  offset: number;
}
