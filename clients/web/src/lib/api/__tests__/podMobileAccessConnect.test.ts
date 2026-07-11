import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  GetMobileAccessDescriptorRequestSchema,
  MobileAccessDescriptorSchema,
  PodSchema,
  UpdatePodPreviewConfigRequestSchema,
  UpdatePodPreviewConfigResponseSchema,
} from "@proto/pod/v1/pod_pb";

vi.mock("@/lib/wasm-core", () => ({
  getPodService: vi.fn(),
}));

import { getPodService } from "@/lib/wasm-core";
import {
  getMobileAccessDescriptor,
  updatePodPreviewConfig,
} from "../connect/podConnect";

const service = {
  get_mobile_access_descriptor_connect: vi.fn(),
  update_pod_preview_config_connect: vi.fn(),
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getPodService).mockReturnValue(
    service as unknown as ReturnType<typeof getPodService>,
  );
});

describe("mobile access Connect boundary", () => {
  it("maps the token-free canonical descriptor", async () => {
    service.get_mobile_access_descriptor_connect.mockResolvedValue(
      toBinary(
        MobileAccessDescriptorSchema,
        create(MobileAccessDescriptorSchema, {
          canonicalUrl: "https://app.example/acme/mobile/workers/pod-1",
          podKey: "pod-1",
          status: "running",
          interactionMode: "acp",
          consoleAvailable: true,
          previewAvailable: true,
          relayAvailable: true,
          previewPath: "/app",
        }),
      ),
    );

    const descriptor = await getMobileAccessDescriptor("acme", "pod-1");
    expect(descriptor).toEqual({
      canonical_url: "https://app.example/acme/mobile/workers/pod-1",
      pod_key: "pod-1",
      status: "running",
      interaction_mode: "acp",
      console_available: true,
      preview_available: true,
      relay_available: true,
      preview_path: "/app",
    });
    expect(descriptor.canonical_url).not.toContain("token");
    const request = fromBinary(
      GetMobileAccessDescriptorRequestSchema,
      service.get_mobile_access_descriptor_connect.mock.calls[0][0],
    );
    expect(request).toMatchObject({ orgSlug: "acme", podKey: "pod-1" });
  });

  it("round-trips preview config through the typed Pod response", async () => {
    service.update_pod_preview_config_connect.mockResolvedValue(
      toBinary(
        UpdatePodPreviewConfigResponseSchema,
        create(UpdatePodPreviewConfigResponseSchema, {
          pod: create(PodSchema, {
            id: BigInt(1),
            podKey: "pod-1",
            status: "running",
            interactionMode: "pty",
            previewPort: 4321,
            previewPath: "/next/api",
          }),
        }),
      ),
    );

    const pod = await updatePodPreviewConfig("acme", "pod-1", 4321, "/next//api/");
    expect(pod).toMatchObject({
      pod_key: "pod-1",
      preview_port: 4321,
      preview_path: "/next/api",
    });
    const request = fromBinary(
      UpdatePodPreviewConfigRequestSchema,
      service.update_pod_preview_config_connect.mock.calls[0][0],
    );
    expect(request).toMatchObject({
      orgSlug: "acme",
      podKey: "pod-1",
      previewPort: 4321,
      previewPath: "/next//api/",
    });
  });
});
