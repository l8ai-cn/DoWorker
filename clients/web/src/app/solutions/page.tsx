import { SolutionDomains } from "@/components/landing/expert-home/SolutionDomains";
import { MarketingPageHero } from "@/components/landing/expert-pages/MarketingPageHero";
import { MarketingPageShell } from "@/components/landing/expert-pages/MarketingPageShell";

export default function SolutionsPage() {
  return (
    <MarketingPageShell>
      <MarketingPageHero page="solutions" />
      <SolutionDomains showIntro={false} />
    </MarketingPageShell>
  );
}
