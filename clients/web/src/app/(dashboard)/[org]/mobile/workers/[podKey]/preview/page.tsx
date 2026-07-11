"use client";

import { useParams } from "next/navigation";
import { MobileWorkerPreview } from "@/components/mobile-worker/MobileWorkerPreview";

export default function MobileWorkerPreviewPage() {
  const params = useParams<{ org: string; podKey: string }>();
  const orgSlug = typeof params.org === "string" ? params.org : "";
  const podKey = typeof params.podKey === "string" ? params.podKey : "";

  return <MobileWorkerPreview orgSlug={orgSlug} podKey={podKey} />;
}
