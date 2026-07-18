import path from "node:path";

export function loadPublicWorkerDefinitions({
  definitionCatalog,
  readJson,
  root,
}) {
  return definitionCatalog.worker_types
    .map((entry) => ({
      entry,
      definition: readJson(path.join(root, entry.definition_path)),
    }))
    .filter(({ definition }) => definition.internal !== true);
}
