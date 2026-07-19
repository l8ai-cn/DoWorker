import crypto from "node:crypto";
import test from "node:test";
import assert from "node:assert/strict";
import { digestCanonical } from "./pattern-worker-preflight-artifact.mjs";

test("digestCanonical matches Go encoding/json HTML escaping", () => {
  const canonical = '{"script":"echo \\u003cx\\u003e\\u0026\\u2028"}';
  const expected = `sha256:${crypto.createHash("sha256").update(canonical).digest("hex")}`;

  assert.equal(digestCanonical({ script: "echo <x>&\u2028" }), expected);
});
