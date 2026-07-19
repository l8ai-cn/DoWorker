import { ZoomInIcon, ZoomOutIcon } from "lucide-react";
import { useEffect, useRef, useState, type ReactNode } from "react";

import type { AgentLightboxImage } from "./AgentImageLightbox";

const MIN_ZOOM = 1;
const MAX_ZOOM = 8;
const ZOOM_STEP = 0.5;

export function AgentImageZoomViewer({ image }: { image: AgentLightboxImage }) {
  const [zoom, setZoom] = useState(MIN_ZOOM);
  const [offset, setOffset] = useState({ x: 0, y: 0 });
  const dragRef = useRef<{
    pointerId: number;
    startX: number;
    startY: number;
  } | null>(null);
  const zoomed = zoom > MIN_ZOOM;
  const applyZoom = (next: number) =>
    setZoom(Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, next)));

  useEffect(() => {
    if (!zoomed) setOffset({ x: 0, y: 0 });
  }, [zoomed]);

  return (
    <>
      <div
        className="absolute inset-0 flex items-center justify-center overflow-hidden"
        onDoubleClick={() => applyZoom(zoomed ? MIN_ZOOM : 2)}
        onPointerCancel={() => {
          dragRef.current = null;
        }}
        onPointerDown={(event) => {
          if (!zoomed) return;
          dragRef.current = {
            pointerId: event.pointerId,
            startX: event.clientX - offset.x,
            startY: event.clientY - offset.y,
          };
          event.currentTarget.setPointerCapture(event.pointerId);
        }}
        onPointerMove={(event) => {
          const drag = dragRef.current;
          if (!drag) return;
          setOffset({
            x: event.clientX - drag.startX,
            y: event.clientY - drag.startY,
          });
        }}
        onPointerUp={(event) => {
          if (dragRef.current?.pointerId === event.pointerId) {
            dragRef.current = null;
          }
        }}
        style={{ cursor: zoomed ? "grab" : "zoom-in" }}
      >
        <img
          alt={image.alt}
          className="max-h-[92vh] max-w-[94vw] origin-center object-contain select-none"
          draggable={false}
          src={image.src}
          style={{
            transform: `translate(${offset.x}px, ${offset.y}px) scale(${zoom})`,
            transition: dragRef.current ? "none" : "transform 120ms ease-out",
          }}
        />
      </div>
      <div className="absolute bottom-3 left-1/2 z-[62] flex -translate-x-1/2 items-center gap-1 rounded-full bg-background/80 p-1 shadow-sm ring-1 ring-foreground/10 backdrop-blur-xs">
        <ZoomButton
          disabled={zoom <= MIN_ZOOM}
          label="Zoom out"
          onClick={() => applyZoom(zoom - ZOOM_STEP)}
        >
          <ZoomOutIcon className="size-4" />
        </ZoomButton>
        <button
          aria-label="Reset zoom"
          className="min-w-[3ch] px-1 text-center text-xs tabular-nums text-muted-foreground hover:text-foreground"
          onClick={() => applyZoom(MIN_ZOOM)}
          type="button"
        >
          {Math.round(zoom * 100)}%
        </button>
        <ZoomButton
          disabled={zoom >= MAX_ZOOM}
          label="Zoom in"
          onClick={() => applyZoom(zoom + ZOOM_STEP)}
        >
          <ZoomInIcon className="size-4" />
        </ZoomButton>
      </div>
    </>
  );
}

function ZoomButton({
  children,
  disabled,
  label,
  onClick,
}: {
  children: ReactNode;
  disabled: boolean;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      aria-label={label}
      className="inline-flex size-8 items-center justify-center rounded-full text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-40"
      disabled={disabled}
      onClick={onClick}
      type="button"
    >
      {children}
    </button>
  );
}
