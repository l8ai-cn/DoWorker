import type {
  AgentPermissionQuestion,
  AgentPermissionRequest,
} from "@do-worker/agent-ui";

import type { AcpPermissionRequest } from "@/stores/acpSessionTypes";

export function projectWebAcpPermission(
  permission: AcpPermissionRequest,
): AgentPermissionRequest {
  const payload = parseObject(permission.argumentsJson);
  const questions = questionPayload(permission.toolName, payload);
  if (questions.length > 0) {
    return {
      id: permission.requestId,
      kind: "question",
      title: permission.description || "Agent needs input",
      questions,
    };
  }
  return {
    id: permission.requestId,
    kind: "approval",
    title: permission.toolName || "Agent approval",
    description: permission.description || permission.argumentsJson,
  };
}

function questionPayload(
  toolName: string,
  payload: Record<string, unknown>,
): AgentPermissionQuestion[] {
  if (toolName === "mcpElicitation") {
    return schemaQuestions(asRecord(payload.requestedSchema));
  }
  if (toolName !== "requestUserInput" && toolName !== "AskUserQuestion") {
    return [];
  }
  if (!Array.isArray(payload.questions)) return [];
  return payload.questions.flatMap((entry, index) => {
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
  const properties = asRecord(schema.properties);
  return Object.entries(properties).flatMap(([id, value]) => {
    if (!isRecord(value)) return [];
    const items = asRecord(value.items);
    const rawOptions = Array.isArray(value.enum)
      ? value.enum
      : Array.isArray(items.enum)
        ? items.enum
        : [];
    const options = rawOptions
      .filter((entry): entry is string => typeof entry === "string")
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

function parseObject(value: string): Record<string, unknown> {
  try {
    const parsed: unknown = JSON.parse(value);
    return asRecord(parsed);
  } catch {
    return {};
  }
}

function asRecord(value: unknown): Record<string, unknown> {
  return isRecord(value) ? value : {};
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
