export function mapRuntimeCatalogEvidence(runtimeCatalog, lockProbes, slug) {
  const image = runtimeCatalog.images.find(
    (candidate) => candidate.worker_type_slugs.includes(slug),
  );
  if (!image) return { status: "blocked_no_published_digest" };

  const probe = lockProbes.get(slug);
  if (!probe) {
    throw new Error(`Missing runtime lock probe for ${slug}`);
  }
  if (probe.status === "available" && image.enabled) {
    return { status: "locked_available", reference: image.reference };
  }
  if (probe.status === "available") {
    return { status: "disabled_published_runtime", reference: image.reference };
  }
  return {
    status: "invalid_published_digest",
    reference: image.reference,
    probe_status: probe.status,
  };
}

export function hasAvailableRuntime(catalogEvidence) {
  return catalogEvidence.status === "locked_available";
}
