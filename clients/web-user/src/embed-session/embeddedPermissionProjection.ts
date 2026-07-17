import type {
  AgentPermissionQuestion,
  AgentPermissionRequest,
} from "@do-worker/agent-ui";

import type { RenderItem } from "@/lib/renderItems";

type ElicitationItem = Extract<RenderItem, { kind: "elicitation" }>;

export function projectEmbeddedPermission(
  item: ElicitationItem,
): AgentPermissionRequest {
  const questions = structuredQuestions(item);
  if (questions.length > 0) {
    return {
      id: item.elicitationId,
      kind: "question",
      title: item.message || "Agent needs input",
      questions,
    };
  }
  return {
    id: item.elicitationId,
    kind: "approval",
    title: item.message || "Agent approval required",
    description:
      item.contentPreview || item.policyName || "Review the requested agent action.",
  };
}

function structuredQuestions(item: ElicitationItem): AgentPermissionQuestion[] {
  const direct = questionEntries(item.askUserQuestion);
  if (direct.length > 0) return direct;
  return schemaQuestions(item.requestedSchema);
}

function questionEntries(value: Record<string, unknown> | null | undefined) {
  if (!value || !Array.isArray(value.questions)) return [];
  return value.questions.flatMap((entry, index) => {
    if (!isRecord(entry) || typeof entry.question !== "string") return [];
    const prompt = entry.question.trim();
    if (!prompt) return [];
    const options = Array.isArray(entry.options)
      ? entry.options.flatMap((option) => {
          if (!isRecord(option) || typeof option.label !== "string") return [];
          const label = option.label.trim();
          if (!label) return [];
          return [{
            label,
            description:
              typeof option.description === "string" ? option.description : "",
          }];
        })
      : [];
    return [{
      id:
        typeof entry.id === "string" && entry.id.trim()
          ? entry.id
          : `question-${index + 1}`,
      prompt,
      header:
        typeof entry.header === "string" && entry.header.trim()
          ? entry.header
          : `Question ${index + 1}`,
      options,
      multiple: entry.multiSelect === true,
      allowCustom: entry.isOther !== false,
      secret: entry.isSecret === true,
    }];
  });
}

function schemaQuestions(schema: Record<string, unknown>): AgentPermissionQuestion[] {
  if (!isRecord(schema.properties)) return [];
  return Object.entries(schema.properties).flatMap(([id, value]) => {
    if (!isRecord(value)) return [];
    const arrayItems = isRecord(value.items) ? value.items : null;
    const rawOptions = Array.isArray(value.enum)
      ? value.enum
      : Array.isArray(arrayItems?.enum)
        ? arrayItems.enum
        : [];
    const options = rawOptions
      .filter((option): option is string => typeof option === "string")
      .map((label) => ({ label, description: "" }));
    const title = typeof value.title === "string" ? value.title : id;
    return [{
      id,
      prompt:
        typeof value.description === "string" && value.description.trim()
          ? value.description
          : title,
      header: title,
      options,
      multiple: value.type === "array",
      allowCustom: options.length === 0,
      secret: value.format === "password",
    }];
  });
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
