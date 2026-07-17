import { CapabilitySpectrum } from "@/components/landing/expert-home/CapabilitySpectrum";
import { ExpertGovernance } from "@/components/landing/expert-home/ExpertGovernance";
import { ExpertOperatingModel } from "@/components/landing/expert-home/ExpertOperatingModel";
import { MarketingPageHero } from "@/components/landing/expert-pages/MarketingPageHero";
import { MarketingPageShell } from "@/components/landing/expert-pages/MarketingPageShell";

export default function ProductPage() {
  return (
    <MarketingPageShell>
      <MarketingPageHero page="product" />
      <ExpertOperatingModel showIntro={false} />
      <CapabilitySpectrum showIntro />
      <ExpertGovernance />
    </MarketingPageShell>
  );
}
