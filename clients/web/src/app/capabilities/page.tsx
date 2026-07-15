import { CapabilitySpectrum } from "@/components/landing/expert-home/CapabilitySpectrum";
import { ExpertGovernance } from "@/components/landing/expert-home/ExpertGovernance";
import { MarketingPageHero } from "@/components/landing/expert-pages/MarketingPageHero";
import { MarketingPageShell } from "@/components/landing/expert-pages/MarketingPageShell";

export default function CapabilitiesPage() {
  return (
    <MarketingPageShell>
      <MarketingPageHero page="capabilities" />
      <CapabilitySpectrum showIntro={false} />
      <ExpertGovernance />
    </MarketingPageShell>
  );
}
