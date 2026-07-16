import { describe, expect, it } from "vitest";

import {
  EMBED_READY_MESSAGE,
  readAllowedEmbedOpenProof,
  isEmbedReadyMessage,
} from "./embedParentHandshake";

describe("embed parent handshake", () => {
  it("accepts an open message only from an allowed parent origin", () => {
    const accepted = readAllowedEmbedOpenProof(
      {
        origin: "https://portal.example",
        source: window.parent,
        data: {
          type: "agentsmesh.embed.open",
          version: 1,
          redemptionProof: "parent-proof",
        },
      },
      window.parent,
      ["https://portal.example"],
    );
    const rejected = readAllowedEmbedOpenProof(
      {
        origin: "https://attacker.example",
        source: window.parent,
        data: {
          type: "agentsmesh.embed.open",
          version: 1,
          redemptionProof: "parent-proof",
        },
      },
      window.parent,
      ["https://portal.example"],
    );

    expect(accepted).toBe("parent-proof");
    expect(rejected).toBeNull();
  });

  it("recognizes only a versioned ready notification", () => {
    expect(isEmbedReadyMessage({ type: EMBED_READY_MESSAGE, version: 1 })).toBe(true);
    expect(isEmbedReadyMessage({ type: EMBED_READY_MESSAGE, version: 2 })).toBe(false);
    expect(isEmbedReadyMessage({ type: "agentsmesh.embed.open", version: 1 })).toBe(false);
  });
});
