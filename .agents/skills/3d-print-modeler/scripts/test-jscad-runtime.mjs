#!/usr/bin/env node

import assert from "node:assert/strict";
import { cp, mkdtemp, mkdir, readFile, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";

const skillDirectory = dirname(dirname(fileURLToPath(import.meta.url)));
const workspace = await mkdtemp(join(tmpdir(), "jscad-runtime-"));
const deliverables = join(workspace, "deliverables");
const runtime = join(deliverables, ".jscad-runtime");
await mkdir(deliverables);
await cp(join(skillDirectory, "assets/jscad-runtime"), runtime, {
  recursive: true,
});

run("npm", [
  "ci",
  "--prefix",
  runtime,
  "--ignore-scripts",
  "--no-audit",
  "--no-fund",
]);

const modelPath = join(deliverables, "model.jscad.mjs");
const stlPath = join(deliverables, "model.stl");
await writeFile(modelPath, `
import { createRequire } from "node:module";
import { writeFile } from "node:fs/promises";
const require = createRequire(import.meta.url);
const { primitives } = require("./.jscad-runtime/node_modules/@jscad/modeling");
const { serialize } = require("./.jscad-runtime/node_modules/@jscad/stl-serializer");
const data = serialize({ binary: true }, primitives.cuboid({ size: [10, 20, 30] }));
await writeFile(
  new URL("model.stl", import.meta.url),
  Buffer.concat(data.map((part) => Buffer.from(part))),
);
`);
run(process.execPath, [modelPath]);
run(process.execPath, [
  join(skillDirectory, "scripts/validate-stl.mjs"),
  stlPath,
  join(deliverables, "validation.json"),
]);

const report = JSON.parse(
  await readFile(join(deliverables, "validation.json"), "utf8"),
);
assert.equal(report.passed, true);
assert.deepEqual(report.dimensionsMm, [10, 20, 30]);
process.stdout.write("locked JSCAD runtime test passed\n");

function run(command, args) {
  const result = spawnSync(command, args, { encoding: "utf8" });
  assert.equal(result.status, 0, result.stderr || result.stdout);
}
