import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  GetMobileAccessDescriptorRequestSchema,
  GetPodConnectionRequestSchema,
  MobileAccessDescriptorSchema,
  PodConnectionInfoSchema,
} from "@do-worker/proto/pod/v1/pod_pb";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getMobilePodConnection, getMobileWorkerDescriptor } from "./mobile-pod-api";
import { readOrgSlug } from "./auth-store";
import { getMobilePodService } from "./mobile-wasm";

vi.mock("./auth-store", () => ({ readOrgSlug: vi.fn() }));
vi.mock("./mobile-wasm", () => ({ getMobilePodService: vi.fn() }));

const podService = {
  get_mobile_access_descriptor_connect: vi.fn(),
  get_pod_connection_connect: vi.fn(),
};

describe("mobile Pod API", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(readOrgSlug).mockReturnValue("dev-org");
    vi.mocked(getMobilePodService).mockResolvedValue(podService as never);
  });

  it("uses the authenticated organization to resolve a Worker descriptor", async () => {
    podService.get_mobile_access_descriptor_connect.mockResolvedValue(
      toBinary(
        MobileAccessDescriptorSchema,
        create(MobileAccessDescriptorSchema, {
          canonicalUrl: "https://mobile.example/workers/pod-1",
          podKey: "pod-1",
          interactionMode: "acp",
          consoleAvailable: true,
        }),
      ),
    );

    await expect(getMobileWorkerDescriptor("pod-1")).resolves.toMatchObject({
      canonicalUrl: "https://mobile.example/workers/pod-1",
      interactionMode: "acp",
      consoleAvailable: true,
    });

    const request = fromBinary(
      GetMobileAccessDescriptorRequestSchema,
      podService.get_mobile_access_descriptor_connect.mock.calls[0][0],
    );
    expect(request.orgSlug).toBe("dev-org");
    expect(request.podKey).toBe("pod-1");
  });

  it("mints a direct Pod Relay connection without requiring an Agent Session", async () => {
    podService.get_pod_connection_connect.mockResolvedValue(
      toBinary(
        PodConnectionInfoSchema,
        create(PodConnectionInfoSchema, {
          relayUrl: "wss://relay.example",
          token: "short-lived-token",
          podKey: "pod-1",
        }),
      ),
    );

    await expect(getMobilePodConnection("pod-1")).resolves.toEqual({
      relayUrl: "wss://relay.example",
      token: "short-lived-token",
      podKey: "pod-1",
    });

    const request = fromBinary(
      GetPodConnectionRequestSchema,
      podService.get_pod_connection_connect.mock.calls[0][0],
    );
    expect(request.orgSlug).toBe("dev-org");
    expect(request.podKey).toBe("pod-1");
  });
});
