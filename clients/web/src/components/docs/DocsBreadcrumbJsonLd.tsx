interface DocsBreadcrumbJsonLdProps {
  breadcrumbs: Array<{ titleKey: string; href?: string }>;
  labels: string[];
}

export function DocsBreadcrumbJsonLd({
  breadcrumbs,
  labels,
}: DocsBreadcrumbJsonLdProps) {
  if (breadcrumbs.length <= 1) return null;

  const items = breadcrumbs.map((crumb, index) => ({
    "@type": "ListItem" as const,
    position: index + 1,
    name: labels[index],
    ...(crumb.href ? { item: `https://agentsmesh.ai${crumb.href}` } : {}),
  }));

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{
        __html: JSON.stringify({
          "@context": "https://schema.org",
          "@type": "BreadcrumbList",
          itemListElement: items,
        }),
      }}
    />
  );
}
