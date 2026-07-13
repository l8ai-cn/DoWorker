import * as Blockly from "blockly";

const IDENTIFIER_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const PARAMETER_PATTERN = /{{\s*([^{}]+?)\s*}}/g;

export interface CustomBlockDefinition {
  id: string;
  name: string;
  template: string;
  parameters: string[];
}

export interface CustomBlockDefinitionInput {
  id: string;
  name: string;
  template: string;
}

export interface CustomBlockDefinitionResult {
  definition?: CustomBlockDefinition;
  errors: string[];
}

export interface CustomBlockExpansionResult {
  value?: string;
  missingParameters: string[];
}

function validIdentifier(value: string): boolean {
  return value.length >= 2 &&
    value.length <= 100 &&
    IDENTIFIER_PATTERN.test(value);
}

function extractParameters(template: string): {
  parameters: string[];
  invalid: string[];
} {
  const parameters: string[] = [];
  const invalid: string[] = [];
  for (const match of template.matchAll(PARAMETER_PATTERN)) {
    const parameter = match[1].trim();
    if (!validIdentifier(parameter)) {
      invalid.push(parameter);
    } else if (!parameters.includes(parameter)) {
      parameters.push(parameter);
    }
  }
  return { parameters, invalid };
}

export function createCustomBlockDefinition(
  input: CustomBlockDefinitionInput,
): CustomBlockDefinitionResult {
  const errors: string[] = [];
  if (!validIdentifier(input.id)) {
    errors.push("积木 ID 必须是 2-100 位小写字母、数字或连字符。");
  }
  if (input.name.trim() === "") {
    errors.push("积木名称不能为空。");
  }
  if (input.template.trim() === "") {
    errors.push("积木模板不能为空。");
  }
  const { parameters, invalid } = extractParameters(input.template);
  if (invalid.length > 0) {
    errors.push(`参数名无效：${invalid.join("、")}。`);
  }
  if (errors.length > 0) return { errors };

  return {
    definition: {
      id: input.id,
      name: input.name.trim(),
      template: input.template,
      parameters,
    },
    errors,
  };
}

export function customBlockType(id: string): string {
  return `loom_custom_${id.replaceAll("-", "_")}`;
}

export function registerCustomBlock(definition: CustomBlockDefinition): void {
  const type = customBlockType(definition.id);
  const signature = JSON.stringify(definition);
  const existing = Blockly.Blocks[type] as unknown as
    | Record<string, unknown>
    | undefined;
  if (existing) {
    if (existing.loomDefinitionSignature !== signature) {
      throw new Error(`自定义积木类型冲突：${definition.id}。`);
    }
    return;
  }
  const args = definition.parameters.map((parameter) => ({
    type: "field_input",
    name: parameter,
    text: parameter,
  }));
  const fields = args.map((_, index) => `%${index + 1}`).join(" ");
  Blockly.common.defineBlocksWithJsonArray([{
    type,
    message0: fields
      ? `${definition.name} ${fields}`
      : definition.name,
    args0: args,
    previousStatement: "LoomInstruction",
    nextStatement: "LoomInstruction",
    style: "loom_task_blocks",
  }]);
  const registered = Blockly.Blocks[type] as unknown as Record<string, unknown>;
  registered.loomDefinitionSignature = signature;
}

export function expandCustomBlockTemplate(
  definition: CustomBlockDefinition,
  values: Record<string, string>,
): CustomBlockExpansionResult {
  const missingParameters = definition.parameters.filter(
    (parameter) => typeof values[parameter] !== "string" ||
      values[parameter].trim() === "",
  );
  if (missingParameters.length > 0) return { missingParameters };
  return {
    missingParameters,
    value: definition.template.replace(
    PARAMETER_PATTERN,
      (_match, rawParameter: string) => values[rawParameter.trim()],
    ),
  };
}
