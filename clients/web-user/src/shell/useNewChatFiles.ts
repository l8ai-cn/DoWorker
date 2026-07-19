import { type DragEvent, useRef, useState } from "react";
import { validateAttachments } from "@/lib/attachments";

export function useNewChatFiles(initialFiles: File[] | undefined) {
  const [files, setFiles] = useState<File[]>(() => initialFiles ?? []);
  const [isDragActive, setIsDragActive] = useState(false);
  const [attachmentError, setAttachmentError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const addFiles = (incoming: File[]) => {
    const { accepted, errors } = validateAttachments(incoming);
    if (accepted.length > 0) setFiles((current) => [...current, ...accepted]);
    setAttachmentError(errors.length > 0 ? errors.join("\n") : null);
  };
  const removeFile = (index: number) => {
    setFiles((current) => current.filter((_, fileIndex) => fileIndex !== index));
    setAttachmentError(null);
  };
  const handleDrop = (event: DragEvent<HTMLFormElement>) => {
    event.preventDefault();
    event.stopPropagation();
    setIsDragActive(false);
    const dropped = Array.from(event.dataTransfer.files);
    if (dropped.length > 0) addFiles(dropped);
  };
  const handleDragOver = (event: DragEvent<HTMLFormElement>) => {
    event.preventDefault();
    event.stopPropagation();
    setIsDragActive(true);
  };
  const handleDragEnter = (event: DragEvent<HTMLFormElement>) => {
    event.preventDefault();
    event.stopPropagation();
    setIsDragActive(true);
  };
  const handleDragLeave = (event: DragEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (event.currentTarget.contains(event.relatedTarget as Node)) return;
    setIsDragActive(false);
  };

  return {
    files,
    attachmentError,
    fileInputRef,
    isDragActive,
    addFiles,
    removeFile,
    handleDrop,
    handleDragOver,
    handleDragEnter,
    handleDragLeave,
  };
}
