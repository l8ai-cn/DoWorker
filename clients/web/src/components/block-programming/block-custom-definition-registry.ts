export interface BlockCustomDefinitionIdentity {
  slug: string;
}

export interface BlockCustomDefinitionRegistration<TDefinition> {
  definitions: readonly TDefinition[];
  registered: boolean;
}

export function hasBlockCustomDefinition<
  TDefinition extends BlockCustomDefinitionIdentity,
>(
  definitions: readonly TDefinition[],
  slug: string,
): boolean {
  return definitions.some((definition) => definition.slug === slug);
}

export function registerBlockCustomDefinition<
  TDefinition extends BlockCustomDefinitionIdentity,
>(
  definitions: readonly TDefinition[],
  definition: TDefinition,
): BlockCustomDefinitionRegistration<TDefinition> {
  if (hasBlockCustomDefinition(definitions, definition.slug)) {
    return { definitions, registered: false };
  }
  return { definitions: [...definitions, definition], registered: true };
}
