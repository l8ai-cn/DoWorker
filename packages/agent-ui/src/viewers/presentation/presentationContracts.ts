import type { AgentArtifactActionCommand } from "../../contracts";

export const PRESENTATION_GRANTS = {
  exportPresentation: "presentation.export",
  regenerateSlide: "presentation.regenerate_slide",
  reorderSlide: "presentation.reorder_slide",
  replaceSlide: "presentation.replace_slide",
  selectVersion: "presentation.select_version",
} as const;

export type PresentationGrant =
  (typeof PRESENTATION_GRANTS)[keyof typeof PRESENTATION_GRANTS];

export interface PresentationSlide {
  readonly imageSrc: string;
  readonly notes?: string;
  readonly position: number;
  readonly representationId?: string;
  readonly slideId: string;
  readonly thumbnailSrc?: string;
  readonly title?: string;
}

export interface PresentationVersion {
  readonly id: string;
  readonly label: string;
  readonly revision: bigint;
}

export type PresentationActionPayload =
  | { readonly slideId: string }
  | { readonly slideId: string; readonly targetIndex: number }
  | { readonly format: "pptx"; readonly slideId: string };

export interface PresentationArtifactAction extends AgentArtifactActionCommand<
  PresentationGrant,
  PresentationActionPayload
> {}

export interface PresentationArtifactViewerProps {
  readonly actionSchemaVersion: string;
  readonly artifactId: string;
  readonly baseRevision: bigint;
  readonly grants: readonly PresentationGrant[];
  readonly initialSlideId?: string;
  readonly onAction: (action: PresentationArtifactAction) => void;
  readonly onRequestFullscreen?: () => void | Promise<void>;
  readonly onSelectVersion?: (versionId: string) => void;
  readonly representationId?: string;
  readonly selectedVersionId: string;
  readonly slides: readonly PresentationSlide[];
  readonly versions: readonly PresentationVersion[];
}
