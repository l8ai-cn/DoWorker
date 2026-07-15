import { CapabilitySpectrum } from "./CapabilitySpectrum";
import { ExpertGovernance } from "./ExpertGovernance";
import { ExpertHero } from "./ExpertHero";
import { ExpertMarketplace } from "./ExpertMarketplace";
import { ExpertOperatingModel } from "./ExpertOperatingModel";
import { SolutionDomains } from "./SolutionDomains";

export function ExpertHome() {
  return (
    <>
      <ExpertHero />
      <SolutionDomains />
      <ExpertOperatingModel />
      <CapabilitySpectrum />
      <ExpertMarketplace />
      <ExpertGovernance />
    </>
  );
}
