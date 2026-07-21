import {
  ImageLightboxProvider as AgentImageLightboxProvider,
  type ImageLightboxProviderProps,
} from "@agent-cloud/agent-ui";

import { getEmbedRoot } from "@/lib/host";

export function ImageLightboxProvider({
  children,
}: Pick<ImageLightboxProviderProps, "children">) {
  return (
    <AgentImageLightboxProvider portalContainer={getEmbedRoot()}>
      {children}
    </AgentImageLightboxProvider>
  );
}

export { useLightbox, ZoomableImage } from "@agent-cloud/agent-ui";
export type { ZoomableImageProps } from "@agent-cloud/agent-ui";
