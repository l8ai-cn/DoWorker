"use client";

import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { ExternalLink, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";

interface MediaLightboxProps {
  src: string;
  alt?: string;
  open: boolean;
  onClose: () => void;
}

// Fullscreen image lightbox. Esc or backdrop click closes; a toolbar offers
// opening the original in a new tab.
export function MediaLightbox({ src, alt, open, onClose }: MediaLightboxProps) {
  const t = useTranslations("media");

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKey);
      document.body.style.overflow = "";
    };
  }, [open, onClose]);

  if (!open) return null;

  return createPortal(
    <div
      role="dialog"
      aria-modal="true"
      aria-label={t("preview")}
      className="fixed inset-0 z-[60] flex items-center justify-center bg-black/80 p-4 animate-in fade-in duration-150"
      onClick={onClose}
    >
      <div className="absolute right-4 top-4 flex items-center gap-2">
        <a
          href={src}
          target="_blank"
          rel="noopener noreferrer"
          onClick={(e) => e.stopPropagation()}
          aria-label={t("openInNewTab")}
          className="flex h-9 w-9 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
        >
          <ExternalLink className="h-4 w-4" />
        </a>
        <button
          type="button"
          onClick={onClose}
          aria-label={t("close")}
          className="flex h-9 w-9 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={src}
        alt={alt ?? ""}
        onClick={(e) => e.stopPropagation()}
        className="max-h-[92vh] max-w-[94vw] rounded-md object-contain shadow-2xl"
      />
    </div>,
    document.body,
  );
}

interface LightboxImageProps {
  src: string;
  alt?: string;
  className?: string;
  imgClassName?: string;
  onError?: () => void;
}

// Inline thumbnail that opens the lightbox on click.
export function LightboxImage({ src, alt, className, imgClassName, onError }: LightboxImageProps) {
  const t = useTranslations("media");
  const [open, setOpen] = useState(false);

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        aria-label={t("preview")}
        className={cn(
          "inline-block cursor-zoom-in overflow-hidden rounded-md border border-border align-top",
          className,
        )}
      >
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={src}
          alt={alt ?? ""}
          className={cn("block max-w-full", imgClassName)}
          onError={onError}
        />
      </button>
      <MediaLightbox src={src} alt={alt} open={open} onClose={() => setOpen(false)} />
    </>
  );
}
