import type {
  ListingSummary,
  TaxonomyTag,
} from "./marketplace-types";

type FilterableTaxonomyKind =
  | "scene"
  | "industry"
  | "audience"
  | "capability"
  | "integration"
  | "readiness";

export const taxonomyFilterGroups: Array<{
  key: FilterableTaxonomyKind;
  label: string;
}> = [
  { key: "scene", label: "业务场景" },
  { key: "industry", label: "所属行业" },
  { key: "audience", label: "适用角色" },
  { key: "capability", label: "核心能力" },
  { key: "integration", label: "系统连接" },
  { key: "readiness", label: "启用条件" },
];

export function collectTaxonomyTags(
  listings: ListingSummary[],
  kind: FilterableTaxonomyKind,
  selected?: TaxonomyTag,
): TaxonomyTag[] {
  const tags = new Map<string, TaxonomyTag>();
  if (selected) {
    tags.set(selected.slug, selected);
  }
  listings.forEach((listing) => {
    listing.tags
      .filter((tag) => tag.kind === kind)
      .forEach((tag) => tags.set(tag.slug, tag));
  });
  return [...tags.values()];
}
