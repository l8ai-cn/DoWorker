import type { Plugin } from "vite";

const helperErrorMarker = "Calling `require` for";
const helperOpener =
  /([A-Za-z_$][\w$]*)\s*=\s*\/\* @__PURE__ \*\/\s*\(\([A-Za-z_$][\w$]*\)\s*=>\s*typeof require[^]*$/;

interface LocatedHelper {
  start: number;
  end: number;
  varName: string;
}

function locateHelper(code: string): LocatedHelper | null {
  const markerIndex = code.indexOf(helperErrorMarker);
  if (markerIndex === -1) return null;
  const opener = code.slice(0, markerIndex).match(helperOpener);
  if (!opener || opener.index === undefined) return null;
  const closeIndex = code.indexOf("})", markerIndex);
  if (closeIndex === -1) return null;
  return {
    start: opener.index,
    end: closeIndex + 2,
    varName: opener[1],
  };
}

function externalImportName(specifier: string): string {
  return `__doWorkerExt_${specifier.replace(/[^a-zA-Z0-9]/g, "_")}`;
}

function externalImports(externals: readonly string[]): string {
  return externals
    .map(
      (specifier) =>
        `import * as ${externalImportName(specifier)} from ${JSON.stringify(specifier)};`,
    )
    .join("\n");
}

function externalTable(externals: readonly string[]): string {
  return externals
    .map((specifier) => `${JSON.stringify(specifier)}: ${externalImportName(specifier)}`)
    .join(", ");
}

function replacementHelper(helperName: string, externals: readonly string[]): string {
  return (
    `${helperName} = function(id) {\n` +
    `\tconst __doWorkerExternals = { ${externalTable(externals)} };\n` +
    `\tif (Object.prototype.hasOwnProperty.call(__doWorkerExternals, id)) return __doWorkerExternals[id];\n` +
    `\tthrow Error("Calling \`require\` for \\"" + id + "\\" in an environment that doesn't expose the \`require\` function.");\n` +
    `}`
  );
}

export function resolveExternalCjsRequire(externals: readonly string[]): Plugin {
  let patchedCount = 0;
  return {
    name: "resolve-external-cjs-require",
    enforce: "post",
    buildStart() {
      patchedCount = 0;
    },
    renderChunk(code) {
      const helper = locateHelper(code);
      if (!helper) return null;
      patchedCount++;
      const next =
        externalImports(externals) +
        "\n" +
        code.slice(0, helper.start) +
        replacementHelper(helper.varName, externals) +
        code.slice(helper.end);
      return { code: next, map: null };
    },
    generateBundle() {
      if (patchedCount !== 1) {
        throw new Error(
          "resolve-external-cjs-require: expected exactly one rolldown require helper",
        );
      }
    },
  };
}
