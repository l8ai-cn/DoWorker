import {
  ImageLightboxProvider as AgentImageLightboxProvider,
  type ImageLightboxProviderProps,
} from "@do-worker/agent-ui";

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

export { useLightbox, ZoomableImage } from "@do-worker/agent-ui";
export type { ZoomableImageProps } from "@do-worker/agent-ui";
