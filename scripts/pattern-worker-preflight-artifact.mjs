import crypto from "node:crypto";
import {
  openAICompatibleProviders,
  requiredPatternSkills,
} from "./pattern-worker-preflight-config.mjs";

const lovartFields = ["LOVART_ACCESS_KEY", "LOVART_SECRET_KEY"];

export function summarizeWorkerSpec(
  artifacts,
  organizationId,
  actorUserId,
  providerModel,
  lovartCredential,
  orgSlug,
) {
  for (const artifact of artifacts ?? []) {
    const document = artifact.artifact_json;
    const digest = digestCanonical(document);
    const checks = {
      digestVerified: digest === artifact.artifact_digest,
      organizationMatched: Number(document?.organization_id) === Number(organizationId),
      namespaceMatched: document?.namespace === orgSlug,
      workerTypeMatched: document?.worker?.worker_type === "pattern-designer",
      skills: summarizeSkills(document),
      primaryModel: summarizePrimaryModel(document, providerModel, orgSlug),
      lovartSecrets: summarizeLovartSecretRefs(
        document,
        organizationId,
        actorUserId,
        lovartCredential,
        orgSlug,
      ),
    };
    if (Object.values(checks).every(checkReady)) {
      return {
        ready: true,
        snapshotId: artifact.snapshot_id,
        artifactDigest: artifact.artifact_digest,
        ...checks,
      };
    }
  }
  return {
    ready: false,
    artifactCount: artifacts?.length ?? 0,
    reason: "immutable Pattern WorkerSpec dependency artifact is missing required bindings",
  };
}

function checkReady(value) {
  return typeof value === "boolean" ? value : Boolean(value?.ready);
}

function summarizeSkills(document) {
  const skills = Array.isArray(document?.skills) ? document.skills : [];
  const slugs = new Set(skills.map((skill) => skill.slug));
  const ready = requiredPatternSkills.every((slug) => {
    const skill = skills.find((entry) => entry.slug === slug);
    return skill && Number(skill.version) > 0 && skill.content_digest &&
      skill.storage_key && Number(skill.package_size) > 0;
  });
  return { ready, slugs: [...slugs].sort() };
}

function summarizePrimaryModel(document, expected, orgSlug) {
  const model = document?.models?.primary;
  const ready = Boolean(
    model &&
    expected &&
    openAICompatibleProviders.includes(model.provider_key) &&
    model.pin?.reference?.kind === "ModelBinding" &&
    model.pin?.reference?.namespace === orgSlug &&
    model.protocol_adapter === expected.protocol_adapter &&
    model.provider_key === expected.provider_key &&
    sameNumber(model.pin?.domain_id, expected.resource_id) &&
    sameNumber(model.connection_id, expected.connection_id) &&
    sameNumber(model.resource_revision, expected.resource_revision) &&
    sameNumber(model.connection_revision, expected.connection_revision) &&
    model.model_id === expected.model_id &&
    model.modalities?.includes("chat") &&
    model.capabilities?.includes("text-generation"),
  );
  return {
    ready,
    providerKey: model?.provider_key,
    protocolAdapter: model?.protocol_adapter,
    modelResourceId: model?.pin?.domain_id,
    connectionId: model?.connection_id,
    resourceRevision: model?.resource_revision,
    connectionRevision: model?.connection_revision,
    modalities: model?.modalities ?? [],
    capabilities: model?.capabilities ?? [],
  };
}

function summarizeLovartSecretRefs(document, organizationId, actorUserId, expected, orgSlug) {
  const refs = Array.isArray(document?.secret_refs) ? document.secret_refs : [];
  const matched = lovartFields.map((field) => refs.find((ref) =>
    ref.field === field &&
    ref.bundle_key === field &&
    ref.pin?.reference?.kind === "EnvironmentBundle" &&
    ref.pin?.reference?.namespace === orgSlug &&
    ref.pin?.reference?.name === "lovart" &&
    sameNumber(ref.pin?.domain_id, expected?.id) &&
    ref.owner_scope === expected?.owner_scope &&
    sameNumber(ref.owner_id, expected?.owner_id) &&
    secretOwnerMatches(ref, organizationId, actorUserId)));
  return {
    ready: matched.every(Boolean),
    fields: matched.filter(Boolean).map((ref) => ref.field).sort(),
    ownerScopes: [...new Set(matched.filter(Boolean).map((ref) => ref.owner_scope))].sort(),
    bundleIds: matched.filter(Boolean).map((ref) => ref.pin?.domain_id).sort(),
  };
}

function secretOwnerMatches(ref, organizationId, actorUserId) {
  if (ref.owner_scope === "org") return Number(ref.owner_id) === Number(organizationId);
  if (ref.owner_scope === "user") return Number(ref.owner_id) === Number(actorUserId);
  return false;
}

function sameNumber(left, right) {
  return Number(left) > 0 && Number(left) === Number(right);
}

export function digestCanonical(value) {
  return `sha256:${crypto.createHash("sha256").update(stableJSON(value)).digest("hex")}`;
}

function stableJSON(value) {
  if (Array.isArray(value)) return `[${value.map(stableJSON).join(",")}]`;
  if (value && typeof value === "object") {
    return `{${Object.keys(value).sort().map((key) =>
      `${goJSONString(key)}:${stableJSON(value[key])}`).join(",")}}`;
  }
  return typeof value === "string" ? goJSONString(value) : JSON.stringify(value);
}

function goJSONString(value) {
  return JSON.stringify(value)
    .replaceAll("<", "\\u003c")
    .replaceAll(">", "\\u003e")
    .replaceAll("&", "\\u0026")
    .replaceAll("\u2028", "\\u2028")
    .replaceAll("\u2029", "\\u2029");
}
