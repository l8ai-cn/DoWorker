import { create, toBinary } from "@bufbuild/protobuf";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  CreateRegistrationCommandResponseSchema,
  ListExecutionClustersResponseSchema,
} from "@proto/execution_cluster/v1/execution_cluster_pb";

const listExecutionClustersConnect = vi.fn();
const createRegistrationCommandConnect = vi.fn();

vi.mock("@/lib/wasm-core", () => ({
  getExecutionClusterService: () => ({
    listExecutionClustersConnect,
    createRegistrationCommandConnect,
  }),
}));

import {
  createRegistrationCommand,
  listExecutionClusters,
} from "./executionClusterConnect";

describe("execution cluster Connect adapter", () => {
  beforeEach(() => {
    listExecutionClustersConnect.mockReset();
    createRegistrationCommandConnect.mockReset();
  });

  it("lists clusters through the wasm service and preserves the exact cluster id", async () => {
    listExecutionClustersConnect.mockResolvedValue(
      toBinary(
        ListExecutionClustersResponseSchema,
        create(ListExecutionClustersResponseSchema, {
          items: [
            {
              id: 12n,
              slug: "local",
              name: "本地集群",
              kind: "local",
              status: "ready",
            },
          ],
        }),
      ),
    );

    await expect(listExecutionClusters("dev-org")).resolves.toEqual([
      {
        id: 12,
        slug: "local",
        name: "本地集群",
        kind: "local",
        status: "ready",
        runnerCount: 0,
        onlineRunnerCount: 0,
        availableRunnerCount: 0,
        tunnelStatus: "",
      },
    ]);

    expect(listExecutionClustersConnect).toHaveBeenCalledTimes(1);
  });

  it("encodes the selected cluster id when generating a registration command", async () => {
    createRegistrationCommandConnect.mockResolvedValue(
      toBinary(
        CreateRegistrationCommandResponseSchema,
        create(CreateRegistrationCommandResponseSchema, {
          command:
            "runner register --server https://example.test --token secret",
          expiresAt: "2026-07-12T12:15:00Z",
        }),
      ),
    );

    await expect(
      createRegistrationCommand("dev-org", 12, "mac-studio"),
    ).resolves.toEqual({
      command: "runner register --server https://example.test --token secret",
      expiresAt: "2026-07-12T12:15:00Z",
    });
    expect(createRegistrationCommandConnect).toHaveBeenCalledTimes(1);
  });
});
