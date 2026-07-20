import type { JSONMap } from "@/lib/viewModels/blockstore";
import {
  customBlockDefinitionTypeKey,
  isValidCustomBlockDefinition,
  type LoopCustomBlockDefinition,
} from "./loop-custom-block-types";

export const LOOP_CUSTOM_BLOCK_RECORD_KEY = "loop_custom_definition";

export function customBlockDefinitionRecord(
  definition: LoopCustomBlockDefinition,
): JSONMap {
  return {
    type_key: customBlockDefinitionTypeKey(definition.slug),
    revision: definition.version,
    label: definition.label,
    description: "Loop custom block definition",
    default_view: "document",
    supported_views: ["document"],
    required_data_key: [],
    allowed_children: [],
    [LOOP_CUSTOM_BLOCK_RECORD_KEY]: {
      schema_version: 1,
      definition,
    },
  };
}

export function customBlockDefinitionFromRecord(
  data: JSONMap,
): LoopCustomBlockDefinition | undefined {
  const record = objectField(data[LOOP_CUSTOM_BLOCK_RECORD_KEY]);
  if (!record || record.schema_version !== 1) return undefined;
  const definition = decodeDefinition(record.definition);
  if (!definition || !isValidCustomBlockDefinition(definition)) return undefined;
  if (data.type_key !== customBlockDefinitionTypeKey(definition.slug)) return undefined;
  if (data.revision !== definition.version) return undefined;
  return definition;
}

function decodeDefinition(value: unknown): LoopCustomBlockDefinition | undefined {
  const raw = objectField(value);
  const expansion = raw ? objectField(raw.expansion) : undefined;
  if (!raw || !expansion) return undefined;
  if (
    typeof raw.slug !== "string" ||
    typeof raw.version !== "number" ||
    typeof raw.label !== "string" ||
    !Array.isArray(raw.parameters) ||
    raw.parameters.some((parameter) => typeof parameter !== "string") ||
    typeof expansion.agentLocalId !== "string" ||
    typeof expansion.verifierLocalId !== "string" ||
    typeof expansion.promptTemplate !== "string" ||
    typeof expansion.commandTemplate !== "string" ||
    typeof expansion.acceptTemplate !== "string"
  ) return undefined;
  return {
    slug: raw.slug,
    version: raw.version,
    label: raw.label,
    parameters: raw.parameters,
    expansion: {
      agentLocalId: expansion.agentLocalId,
      verifierLocalId: expansion.verifierLocalId,
      promptTemplate: expansion.promptTemplate,
      commandTemplate: expansion.commandTemplate,
      acceptTemplate: expansion.acceptTemplate,
    },
  };
}

function objectField(value: unknown): Record<string, unknown> | undefined {
  return value && typeof value === "object" && !Array.isArray(value)
    ? value as Record<string, unknown>
    : undefined;
}
