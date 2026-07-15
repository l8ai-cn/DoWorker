import { ExpertOperatingModel } from "@/components/landing/expert-home/ExpertOperatingModel";
import { MarketingPageHero } from "@/components/landing/expert-pages/MarketingPageHero";
import { MarketingPageShell } from "@/components/landing/expert-pages/MarketingPageShell";

export default function HowItWorksPage() {
  return (
    <MarketingPageShell>
      <MarketingPageHero page="how-it-works" />
      <ExpertOperatingModel showIntro={false} />
    </MarketingPageShell>
  );
}
