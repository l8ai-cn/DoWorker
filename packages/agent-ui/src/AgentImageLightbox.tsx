import { XIcon } from "lucide-react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { createPortal } from "react-dom";

import { AgentImageZoomViewer } from "./AgentImageZoomViewer";

export interface AgentLightboxImage {
  alt: string;
  src: string;
}

interface LightboxContextValue {
  open: (image: AgentLightboxImage) => void;
}

export interface ImageLightboxProviderProps {
  children: ReactNode;
  portalContainer?: HTMLElement | null;
}

export interface ZoomableImageProps extends React.ComponentProps<"img"> {
  alt: string;
  src?: string;
}

const LightboxContext = createContext<LightboxContextValue | null>(null);

export function useLightbox(): LightboxContextValue {
  const value = useContext(LightboxContext);
  if (!value) {
    throw new Error("useLightbox must be used within ImageLightboxProvider");
  }
  return value;
}

export function ZoomableImage({
  alt,
  className,
  src,
  ...imgProps
}: ZoomableImageProps) {
  const { open } = useLightbox();
  return (
    <button
      aria-label={alt ? `Zoom image: ${alt}` : "Zoom image"}
      className="m-0 inline-flex max-w-full cursor-zoom-in appearance-none border-0 bg-transparent p-0 leading-none"
      onClick={() => {
        if (src) open({ alt, src });
      }}
      type="button"
    >
      <img {...imgProps} alt={alt} className={className} src={src} />
    </button>
  );
}

export function ImageLightboxProvider({
  children,
  portalContainer,
}: ImageLightboxProviderProps) {
  const [image, setImage] = useState<AgentLightboxImage | null>(null);
  const open = useCallback((next: AgentLightboxImage) => setImage(next), []);
  const value = useMemo(() => ({ open }), [open]);
  const target =
    portalContainer ??
    (typeof document === "undefined" ? null : document.body);

  useEffect(() => {
    if (!image) return;
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setImage(null);
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [image]);

  return (
    <LightboxContext.Provider value={value}>
      {children}
      {target && image
        ? createPortal(
            <ImageLightboxDialog
              image={image}
              onClose={() => setImage(null)}
            />,
            target,
          )
        : null}
    </LightboxContext.Provider>
  );
}

function ImageLightboxDialog({
  image,
  onClose,
}: {
  image: AgentLightboxImage;
  onClose: () => void;
}) {
  const closeButtonRef = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    closeButtonRef.current?.focus();
  }, []);

  return (
    <div className="fixed inset-0 z-[60] bg-black/80" role="presentation">
      <section
        aria-label={image.alt || "Image preview"}
        aria-modal="true"
        className="fixed inset-0 z-[61] outline-none"
        role="dialog"
      >
        <AgentImageZoomViewer image={image} />
        <button
          aria-label="Close"
          className="absolute right-3 top-3 inline-flex size-8 items-center justify-center rounded-md bg-background/80 text-foreground shadow-sm ring-1 ring-foreground/10 hover:bg-background"
          onClick={onClose}
          ref={closeButtonRef}
          type="button"
        >
          <XIcon className="size-4" />
        </button>
      </section>
    </div>
  );
}
