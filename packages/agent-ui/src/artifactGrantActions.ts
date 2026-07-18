import type { AgentArtifactItem } from "./agentArtifactContracts";

export function artifactActionAllowed(
  item: AgentArtifactItem,
  actionType: string,
  representationId?: string,
): boolean {
  return item.grants.some((grant) => {
    if (!grant.actions.includes(actionType)) return false;
    if (grant.expiresAt) {
      const expiresAt = Date.parse(grant.expiresAt);
      if (!Number.isFinite(expiresAt) || expiresAt <= Date.now()) return false;
    }
    if (
      grant.representationIds.length > 0 &&
      (!representationId || !grant.representationIds.includes(representationId))
    ) {
      return false;
    }
    if (
      grant.minimumRevision !== undefined &&
      item.revision < grant.minimumRevision
    ) {
      return false;
    }
    return (
      grant.maximumRevision === undefined ||
      item.revision <= grant.maximumRevision
    );
  });
}
