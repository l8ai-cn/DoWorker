#!/usr/bin/env node

import assert from "node:assert/strict";
import { mkdtemp, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";

const vertices = [
  [0, 0, 0],
  [10, 0, 0],
  [0, 10, 0],
  [0, 0, 10],
];
const outwardFaces = [
  [0, 2, 1],
  [0, 1, 3],
  [1, 2, 3],
  [2, 0, 3],
];
const directory = await mkdtemp(join(tmpdir(), "validate-stl-"));

const validReport = await runValidation("valid", outwardFaces);
assert.equal(validReport.status, 0);
assert.equal(validReport.report.passed, true);
assert.equal(validReport.report.nonManifoldEdges, 0);
assert.equal(validReport.report.inconsistentEdges, 0);

const flippedFaces = outwardFaces.map((face) => [...face]);
flippedFaces[0].reverse();
const flippedReport = await runValidation("flipped", flippedFaces);
assert.equal(flippedReport.status, 1);
assert.equal(flippedReport.report.passed, false);
assert.ok(flippedReport.report.inconsistentEdges > 0);

process.stdout.write("validate-stl regression tests passed\n");

async function runValidation(name, faces) {
  const inputPath = join(directory, `${name}.stl`);
  const reportPath = join(directory, `${name}.json`);
  const facets = faces.map((face) => {
    const points = face.map((index) => vertices[index]);
    return [
      "facet normal 0 0 0",
      "  outer loop",
      ...points.map((point) => `    vertex ${point.join(" ")}`),
      "  endloop",
      "endfacet",
    ].join("\n");
  });
  const stl = `solid ${name}\n${facets.join("\n")}\nendsolid ${name}\n`;
  await writeFile(inputPath, stl);
  const result = spawnSync(
    process.execPath,
    [new URL("validate-stl.mjs", import.meta.url).pathname, inputPath, reportPath],
    { encoding: "utf8" },
  );
  return {
    status: result.status,
    report: JSON.parse(result.stdout),
  };
}
