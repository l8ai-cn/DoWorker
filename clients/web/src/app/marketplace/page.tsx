import { MarketplaceApplicationBrowser } from "@/components/marketplace/MarketplaceApplicationBrowser";
import { MarketplaceFooter } from "@/components/marketplace/MarketplaceFooter";
import { MarketplaceHeader } from "@/components/marketplace/MarketplaceHeader";
import { fetchPublicMarketApplications } from "@/lib/public-market-api";

export const dynamic = "force-dynamic";

export default async function MarketplacePage() {
  const result = await loadApplications();

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "CollectionPage",
    name: "Do Worker 专家应用市场",
    description: "可直接启用的 AI 专家应用目录。",
    url: "https://agentsmesh.ai/marketplace",
  };

  return (
    <div className="min-h-screen bg-surface text-foreground">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <MarketplaceHeader />
      <MarketplaceApplicationBrowser
        applications={result.applications}
        loadError={result.error}
      />
      <MarketplaceFooter />
    </div>
  );
}

async function loadApplications(): Promise<{
  applications: Awaited<ReturnType<typeof fetchPublicMarketApplications>>["items"];
  error?: string;
}> {
  try {
    const data = await fetchPublicMarketApplications();
    return { applications: data.items };
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown marketplace error";
    return { applications: [], error: message };
  }
}
