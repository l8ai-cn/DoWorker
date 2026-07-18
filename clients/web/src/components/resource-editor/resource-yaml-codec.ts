import {
  parseAllDocuments,
  stringify,
  visit,
} from "yaml";
import type {
  ResourceDraft,
  ResourceEditorKind,
} from "./resource-editor-types";
import { assertCanonicalResourceDraft } from "./canonical-resource-draft-validation";

const MAX_DOCUMENT_BYTES = 256 * 1024;
const MAX_LINE_BYTES = 64 * 1024;
const MAX_NODES = 10_000;
const MAX_DEPTH = 64;

export function parseResourceYaml(
  source: string,
  expectedKind: ResourceEditorKind,
): ResourceDraft {
  enforceSourceLimits(source);
  const documents = parseAllDocuments(source, {
    resolveKnownTags: false,
    schema: "core",
    strict: true,
    stringKeys: true,
    uniqueKeys: true,
    prettyErrors: false,
  });
  if (documents.length !== 1) {
    throw new Error("YAML must contain exactly one document.");
  }
  const document = documents[0];
  if (document.errors.length > 0) {
    throw new Error("YAML syntax error.");
  }
  enforceTreeLimits(document);
  const value = document.toJS({ maxAliasCount: 0 });
  return assertResourceDraft(value, expectedKind);
}

export function stringifyResourceYaml(draft: ResourceDraft): string {
  const source = stringify(draft, {
    aliasDuplicateObjects: false,
    blockQuote: "literal",
    indent: 2,
    lineWidth: 0,
    resolveKnownTags: false,
    schema: "core",
  });
  enforceSourceLimits(source);
  return source;
}

export function parseCanonicalResourceJson(
  content: Uint8Array,
  expectedKind: ResourceEditorKind,
): ResourceDraft {
  const value = JSON.parse(new TextDecoder().decode(content)) as unknown;
  return assertCanonicalResourceDraft(value, expectedKind);
}

function enforceSourceLimits(source: string): void {
  if (new TextEncoder().encode(source).byteLength > MAX_DOCUMENT_BYTES) {
    throw new Error("YAML exceeds the 256 KiB document limit.");
  }
  for (const line of source.split(/\r?\n/)) {
    if (new TextEncoder().encode(line).byteLength > MAX_LINE_BYTES) {
      throw new Error("YAML contains a line longer than 64 KiB.");
    }
  }
}

function enforceTreeLimits(document: ReturnType<typeof parseAllDocuments>[number]): void {
  let count = 0;
  let forbidden = false;
  let unsupportedNumber = false;
  let unsafeInteger = false;
  const countNode = (depth: number) => {
    count += 1;
    if (count > MAX_NODES || depth > MAX_DEPTH) {
      throw new Error("YAML structure exceeds the supported limits.");
    }
  };
  visit(document, {
    Alias(_key, _node, path) {
      countNode(path.length);
      forbidden = true;
    },
    Node(_key, node, path) {
      countNode(path.length);
      if (("anchor" in node && node.anchor) || node.tag) {
        forbidden = true;
      }
    },
    Scalar(_key, node, path) {
      countNode(path.length);
      if (node.anchor || node.tag) forbidden = true;
      if (node.format || (
        typeof node.value === "number" && !Number.isFinite(node.value)
      )) {
        unsupportedNumber = true;
      }
      if (
        typeof node.value === "number" &&
        Number.isInteger(node.value) &&
        !Number.isSafeInteger(node.value)
      ) {
        unsafeInteger = true;
      }
    },
  });
  if (forbidden) {
    throw new Error("YAML anchors, aliases, and custom tags are not supported.");
  }
  if (unsupportedNumber) {
    throw new Error("YAML contains a non-JSON number.");
  }
  if (unsafeInteger) {
    throw new Error("YAML contains an integer outside the safe integer range.");
  }
}

function assertResourceDraft(
  value: unknown,
  expectedKind: ResourceEditorKind,
): ResourceDraft {
  if (!isRecord(value) ||
      value.apiVersion !== "agentsmesh.io/v1alpha1" ||
      value.kind !== expectedKind ||
      !isRecord(value.metadata) ||
      !isRecord(value.spec)) {
    throw new Error(`YAML must be a ${expectedKind} resource.`);
  }
  return value as unknown as ResourceDraft;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
