#!/usr/bin/env node
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname } from "node:path";

import { executeOilanPostgresReadOnly } from "../Server/lib/oilan-postgres-doops-readonly.mjs";

const { command, inputPath, outputPath } = parseArgs(process.argv);
const input = JSON.parse(await readFile(inputPath, "utf8"));

let response;
try {
  if (command !== "probe" && command !== "query") {
    throw new Error(`unsupported command: ${command}`);
  }
  const queryName = resolveQueryName(command, input);
  const evidence = await executeOilanPostgresReadOnly({ ...input, queryName });
  response = {
    operationId: input.operationId,
    status: "succeeded",
    command,
    mode: "read_only",
    evidence,
  };
} catch (error) {
  response = {
    operationId: input.operationId,
    status: "failed",
    command,
    error: { message: error instanceof Error ? error.message : String(error) },
  };
}

await mkdir(dirname(outputPath), { recursive: true });
await writeFile(outputPath, `${JSON.stringify(response, null, 2)}\n`, "utf8");
if (response.status !== "succeeded") process.exitCode = 1;

function parseArgs(argv) {
  const [command, inputFlag, inputPath, outputFlag, outputPath] = argv.slice(2);
  if (!command || inputFlag !== "--input" || !inputPath || outputFlag !== "--output" || !outputPath) {
    throw new Error("usage: oilan-postgres-doops-readonly.mjs <probe|query> --input <path> --output <path>");
  }
  return { command, inputPath, outputPath };
}

function resolveQueryName(command, input) {
  if (command === "probe") {
    if (Object.hasOwn(input, "queryName")) {
      throw new Error("probe does not accept queryName");
    }
    return "asset-probe";
  }
  if (input.queryName !== "migration-version") {
    throw new Error("queryName must be migration-version");
  }
  return input.queryName;
}
