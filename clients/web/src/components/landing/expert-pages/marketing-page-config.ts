export type MarketingPageId = "product" | "solutions";

interface MarketingPageDefinition {
  index: string;
  eyebrowKey: string;
  titleKey: string;
  descriptionKey: string;
  nextHref: string;
  nextLabelKey: string;
}

export const marketingPageConfig: Record<MarketingPageId, MarketingPageDefinition> = {
  product: {
    index: "01",
    eyebrowKey: "landing.nav.product",
    titleKey: "landing.workforce.expertHome.operating.title",
    descriptionKey: "landing.workforce.expertHome.operating.description",
    nextHref: "/solutions",
    nextLabelKey: "landing.nav.solutions",
  },
  solutions: {
    index: "02",
    eyebrowKey: "landing.nav.solutions",
    titleKey: "landing.workforce.expertHome.solutions.title",
    descriptionKey: "landing.workforce.expertHome.solutions.description",
    nextHref: "/product",
    nextLabelKey: "landing.nav.product",
  },
};
