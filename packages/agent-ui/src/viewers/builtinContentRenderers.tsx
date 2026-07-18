import type { AgentContentRendererRegistration } from "../react/contentRendererTypes";
import { ContentRendererRegistry } from "../registry/ContentRendererRegistry";
import { ImageEditArtifactViewer } from "./image/ImageEditArtifactViewer";
import { PresentationManifestArtifactViewer } from "./presentation/PresentationManifestArtifactViewer";
import { PlainVideoArtifactViewer } from "./video/PlainVideoArtifactViewer";
import { VideoManifestArtifactViewer } from "./video/VideoManifestArtifactViewer";

const VIDEO_MEDIA_TYPES = [
  "video/mp4",
  "video/quicktime",
  "video/webm",
  "video/x-m4v",
] as const;
const VIDEO_ROLES = [
  "artifact",
  "original",
  "playable",
  "preview",
  "primary",
] as const;

export function createBuiltinContentRenderers(): ContentRendererRegistry<AgentContentRendererRegistration> {
  const registry =
    new ContentRendererRegistry<AgentContentRendererRegistration>();
  registerManifest(registry, "image_edit", ImageEditArtifactViewer);
  registerManifest(registry, "video", VideoManifestArtifactViewer);
  registerManifest(
    registry,
    "presentation",
    PresentationManifestArtifactViewer,
  );
  for (const mediaType of VIDEO_MEDIA_TYPES) {
    for (const role of VIDEO_ROLES) {
      registry.register(
        {
          blockKind: "artifact",
          mediaType,
          role,
          schemaVersion: "1",
        },
        { viewer: PlainVideoArtifactViewer },
        `builtin.video.${mediaType}.${role}`,
      );
    }
  }
  return registry;
}

function registerManifest(
  registry: ContentRendererRegistry<AgentContentRendererRegistration>,
  manifestType: string,
  viewer: AgentContentRendererRegistration["viewer"],
) {
  registry.register(
    {
      blockKind: "artifact",
      manifestType,
      schemaVersion: "1",
    },
    { viewer },
    `builtin.${manifestType}`,
  );
}
