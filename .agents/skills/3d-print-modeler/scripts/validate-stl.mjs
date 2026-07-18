#!/usr/bin/env node

import { readFile, writeFile } from "node:fs/promises";

const [inputPath, reportPath] = process.argv.slice(2);
if (!inputPath) {
  console.error("usage: validate-stl.mjs <model.stl> [report.json]");
  process.exit(2);
}

const bytes = await readFile(inputPath);
const triangles = isBinaryStl(bytes) ? parseBinary(bytes) : parseAscii(bytes);
const report = inspect(triangles);
const output = `${JSON.stringify(report, null, 2)}\n`;
if (reportPath) await writeFile(reportPath, output);
process.stdout.write(output);
process.exit(report.passed ? 0 : 1);

function isBinaryStl(buffer) {
  if (buffer.length < 84) return false;
  return 84 + buffer.readUInt32LE(80) * 50 === buffer.length;
}

function parseBinary(buffer) {
  const count = buffer.readUInt32LE(80);
  return Array.from({ length: count }, (_, triangleIndex) => {
    const start = 84 + triangleIndex * 50 + 12;
    return Array.from({ length: 3 }, (_, vertexIndex) => {
      const offset = start + vertexIndex * 12;
      return [
        buffer.readFloatLE(offset),
        buffer.readFloatLE(offset + 4),
        buffer.readFloatLE(offset + 8),
      ];
    });
  });
}

function parseAscii(buffer) {
  const values = [...buffer.toString("utf8").matchAll(
    /\bvertex\s+([^\s]+)\s+([^\s]+)\s+([^\s]+)/gi,
  )].map((match) => match.slice(1).map(Number));
  if (values.length === 0 || values.length % 3 !== 0) {
    throw new Error("STL contains no complete triangle vertices");
  }
  return Array.from(
    { length: values.length / 3 },
    (_, index) => values.slice(index * 3, index * 3 + 3),
  );
}

function inspect(triangles) {
  const bounds = [
    [Infinity, Infinity, Infinity],
    [-Infinity, -Infinity, -Infinity],
  ];
  const edges = new Map();
  let degenerateTriangles = 0;
  let signedVolume = 0;

  for (const triangle of triangles) {
    if (triangle.some((vertex) => vertex.some((value) => !Number.isFinite(value)))) {
      degenerateTriangles += 1;
      continue;
    }
    for (const vertex of triangle) {
      for (let axis = 0; axis < 3; axis += 1) {
        bounds[0][axis] = Math.min(bounds[0][axis], vertex[axis]);
        bounds[1][axis] = Math.max(bounds[1][axis], vertex[axis]);
      }
    }
    const [a, b, c] = triangle;
    const crossProduct = cross(subtract(b, a), subtract(c, a));
    if (dot(crossProduct, crossProduct) <= 1e-18) degenerateTriangles += 1;
    signedVolume += dot(a, cross(b, c)) / 6;
    addEdge(edges, a, b);
    addEdge(edges, b, c);
    addEdge(edges, c, a);
  }

  const edgeStats = [...edges.values()];
  const nonManifoldEdges = edgeStats.filter(({ count }) => count !== 2).length;
  const inconsistentEdges = edgeStats.filter(
    ({ count, orientation }) => count === 2 && orientation !== 0,
  ).length;
  const dimensions = bounds[0].map((minimum, axis) => bounds[1][axis] - minimum);
  const volumeMm3 = Math.abs(signedVolume);
  const passed = triangles.length > 0
    && degenerateTriangles === 0
    && nonManifoldEdges === 0
    && inconsistentEdges === 0
    && volumeMm3 > 0
    && dimensions.every((dimension) => dimension > 0);

  return {
    passed,
    triangleCount: triangles.length,
    dimensionsMm: dimensions.map(round),
    boundsMm: { min: bounds[0].map(round), max: bounds[1].map(round) },
    volumeMm3: round(volumeMm3),
    degenerateTriangles,
    nonManifoldEdges,
    inconsistentEdges,
  };
}

function addEdge(edges, first, second) {
  const firstKey = vertexKey(first);
  const secondKey = vertexKey(second);
  const key = [firstKey, secondKey].sort().join("|");
  const current = edges.get(key) ?? { count: 0, orientation: 0 };
  current.count += 1;
  current.orientation += firstKey < secondKey ? 1 : -1;
  edges.set(key, current);
}

function vertexKey(vertex) {
  return vertex.map((value) => value.toFixed(6)).join(",");
}

function subtract(first, second) {
  return first.map((value, index) => value - second[index]);
}

function cross(first, second) {
  return [
    first[1] * second[2] - first[2] * second[1],
    first[2] * second[0] - first[0] * second[2],
    first[0] * second[1] - first[1] * second[0],
  ];
}

function dot(first, second) {
  return first.reduce((sum, value, index) => sum + value * second[index], 0);
}

function round(value) {
  return Number(value.toFixed(6));
}
