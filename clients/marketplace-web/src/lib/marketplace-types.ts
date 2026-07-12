export type ResourceType = "application" | "skill" | "mcp_connector" | "resource";

export interface Market {
  marketplace_id: string;
  slug: string;
  name: string;
  summary: string;
  status: string;
  default_locale: string;
}

export interface Publisher {
  slug: string;
  display_name: string;
  verified: boolean;
}

export interface Space {
  slug: string;
  name: string;
}

export interface QuotaSummary {
  mode: string;
  estimated_credits_micro: string;
}

export interface ListingSummary {
  listing_id: string;
  listing_version_id: string;
  slug: string;
  resource_type: ResourceType;
  display_name: string;
  tagline: string;
  publisher: Publisher;
  spaces: Space[];
  quota?: QuotaSummary;
  published_at: string;
}

export interface ListingDetail extends ListingSummary {
  description: string;
  outcomes: string[];
  use_cases: string[];
  target_audience: string[];
  requirements: string[];
  permissions: string[];
  version: string;
  release_notes: string;
  examples?: Array<{ input: string; output: string }>;
  documentation_url?: string;
  support_url?: string;
}

export interface MarketplaceErrorEnvelope {
  error?: {
    code?: string;
    message?: string;
    detail?: string;
  };
}
