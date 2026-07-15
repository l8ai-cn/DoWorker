export type MarketingPageId = "solutions" | "how-it-works" | "capabilities";

interface MarketingPageDefinition {
  index: string;
  eyebrowKey: string;
  titleKey: string;
  descriptionKey: string;
  nextHref: string;
  nextLabelKey: string;
}

export const marketingPageConfig: Record<MarketingPageId, MarketingPageDefinition> = {
  solutions: {
    index: "01",
    eyebrowKey: "landing.nav.scenarios",
    titleKey: "landing.workforce.expertHome.solutions.title",
    descriptionKey: "landing.workforce.expertHome.solutions.description",
    nextHref: "/how-it-works",
    nextLabelKey: "landing.nav.workflow",
  },
  "how-it-works": {
    index: "02",
    eyebrowKey: "landing.nav.workflow",
    titleKey: "landing.workforce.expertHome.operating.title",
    descriptionKey: "landing.workforce.expertHome.operating.description",
    nextHref: "/capabilities",
    nextLabelKey: "landing.nav.capabilities",
  },
  capabilities: {
    index: "03",
    eyebrowKey: "landing.nav.capabilities",
    titleKey: "landing.workforce.expertHome.capabilities.title",
    descriptionKey: "landing.workforce.expertHome.capabilities.description",
    nextHref: "/marketplace",
    nextLabelKey: "landing.nav.marketplace",
  },
};
