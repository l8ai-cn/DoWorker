import { useCallback, useEffect, useState } from "react";
import { X } from "lucide-react";

interface Props {
  src: string;
  alt?: string;
  className?: string;
  thumbClassName?: string;
  caption?: string;
  children?: React.ReactNode;
}

/**
 * Wrap any image thumbnail — click to open a full-screen lightbox.
 * Use `children` to pass a custom trigger; otherwise renders an <img>.
 */
export function Lightbox({ src, alt = "", className, thumbClassName, caption, children }: Props) {
  const [open, setOpen] = useState(false);

  const close = useCallback(() => setOpen(false), []);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") close();
    };
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", onKey);
    return () => {
      document.body.style.overflow = prev;
      window.removeEventListener("keydown", onKey);
    };
  }, [open, close]);

  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        aria-label={alt ? `放大查看：${alt}` : "放大查看图片"}
        className={"group block cursor-zoom-in " + (className ?? "")}
      >
        {children ?? (
          <img
            src={src}
            alt={alt}
            loading="lazy"
            className={thumbClassName ?? "w-full object-cover transition-opacity group-hover:opacity-90"}
          />
        )}
      </button>

      {open && (
        <div
          role="dialog"
          aria-modal="true"
          onClick={close}
          className="fixed inset-0 z-[100] flex items-center justify-center bg-black/85 p-4 backdrop-blur-sm"
        >
          <button
            type="button"
            onClick={close}
            aria-label="关闭"
            className="absolute right-4 top-4 flex h-9 w-9 items-center justify-center rounded-full bg-white/10 text-white hover:bg-white/20"
          >
            <X className="h-4 w-4" />
          </button>
          <figure
            onClick={(e) => e.stopPropagation()}
            className="flex max-h-full max-w-full flex-col items-center gap-2"
          >
            <img
              src={src}
              alt={alt}
              className="max-h-[90vh] max-w-[95vw] cursor-zoom-out rounded-lg object-contain shadow-2xl"
              onClick={close}
            />
            {(caption || alt) && (
              <figcaption className="text-center text-[12px] text-white/80">
                {caption ?? alt}
              </figcaption>
            )}
          </figure>
        </div>
      )}
    </>
  );
}
