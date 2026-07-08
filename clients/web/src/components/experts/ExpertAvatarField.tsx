"use client";

import { useCallback, useRef, useState, type ChangeEvent } from "react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { ImagePlus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { ExpertAvatarDraft } from "./expertFormModel";

// Mirror the backend allow-list + 2 MB cap in
// backend/internal/api/rest/v1/expert_handler_git.go so the user gets a fast,
// local error before an upload is attempted. The backend re-validates.
const ALLOWED_TYPES = ["image/png", "image/jpeg", "image/webp", "image/gif"];
const MAX_AVATAR_BYTES = 2 * 1024 * 1024;

interface Props {
  value: ExpertAvatarDraft | null;
  onChange: (avatar: ExpertAvatarDraft | null) => void;
}

/** Strip the `data:<mime>;base64,` prefix, leaving raw base64 for the API. */
function rawBase64(dataUrl: string): string {
  const comma = dataUrl.indexOf(",");
  return comma >= 0 ? dataUrl.slice(comma + 1) : dataUrl;
}

export function ExpertAvatarField({ value, onChange }: Props) {
  const t = useTranslations("experts.create");
  const inputRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState<string | null>(null);

  const handleFile = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      // Reset so re-selecting the same file re-triggers onChange.
      e.target.value = "";
      if (!file) return;
      if (!ALLOWED_TYPES.includes(file.type)) {
        setError(t("avatarErrorType"));
        return;
      }
      if (file.size > MAX_AVATAR_BYTES) {
        setError(t("avatarErrorSize"));
        return;
      }
      setError(null);
      const reader = new FileReader();
      reader.onload = () => {
        const dataUrl = typeof reader.result === "string" ? reader.result : "";
        if (!dataUrl) {
          setError(t("avatarErrorRead"));
          return;
        }
        onChange({
          filename: file.name,
          contentBase64: rawBase64(dataUrl),
          previewUrl: dataUrl,
        });
      };
      reader.onerror = () => setError(t("avatarErrorRead"));
      reader.readAsDataURL(file);
    },
    [onChange, t],
  );

  const clear = useCallback(() => {
    setError(null);
    onChange(null);
  }, [onChange]);

  return (
    <div className="space-y-1.5">
      <div className="flex items-center gap-3">
        <div className="flex h-16 w-16 shrink-0 items-center justify-center overflow-hidden rounded-md bg-surface-raised ring-1 ring-border/35">
          {value ? (
            <Image
              src={value.previewUrl}
              alt={t("avatarLabel")}
              width={64}
              height={64}
              unoptimized
              className="h-full w-full object-cover"
            />
          ) : (
            <ImagePlus className="h-6 w-6 text-muted-foreground" />
          )}
        </div>
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center gap-2">
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => inputRef.current?.click()}
            >
              {value ? t("avatarChange") : t("avatarUpload")}
            </Button>
            {value && (
              <Button type="button" size="sm" variant="ghost" className="gap-1" onClick={clear}>
                <X className="h-3.5 w-3.5" />
                {t("avatarRemove")}
              </Button>
            )}
          </div>
          <p className="text-xs text-muted-foreground">{t("avatarHint")}</p>
        </div>
      </div>
      <input
        ref={inputRef}
        type="file"
        accept={ALLOWED_TYPES.join(",")}
        className="hidden"
        onChange={handleFile}
      />
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  );
}
