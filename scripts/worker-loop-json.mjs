import fs from "node:fs";

export function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
}

export function mapBySlug(items, name) {
  const result = new Map();
  for (const item of items) {
    if (result.has(item.slug)) {
      throw new Error(`${name} repeats Worker slug: ${item.slug}`);
    }
    result.set(item.slug, item);
  }
  return result;
}

export function assertSameSlugs(left, right) {
  const expected = left.map((item) => item.slug).sort().join(",");
  const actual = right.map((item) => item.slug).sort().join(",");
  if (expected !== actual) {
    throw new Error("Definition catalog and evidence matrix Worker slugs differ");
  }
}
