import { Footer, Navbar } from "@/components/landing";
import { MarketplaceSkillBrowser } from "@/components/marketplace/MarketplaceSkillBrowser";
import { fetchPublicMarketSkills } from "@/lib/public-market-api";

export const dynamic = "force-dynamic";

export default async function MarketplacePage() {
  const result = await loadSkills();

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "CollectionPage",
    name: "Do Worker Skill Marketplace",
    description:
      "Public catalog of reusable Skills for AI workers and agent workflows.",
    url: "https://agentsmesh.ai/marketplace",
  };

  return (
    <div className="azure-theme min-h-screen bg-background">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <Navbar />
      <MarketplaceSkillBrowser skills={result.skills} loadError={result.error} />
      <Footer />
    </div>
  );
}

async function loadSkills(): Promise<{ skills: Awaited<ReturnType<typeof fetchPublicMarketSkills>>["items"]; error?: string }> {
  try {
    const data = await fetchPublicMarketSkills();
    return { skills: data.items };
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown marketplace error";
    return { skills: [], error: message };
  }
}
