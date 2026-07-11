"use client";

import { useParams } from "next/navigation";
import { MobilePodWorkspace } from "@/components/mobile/MobilePodWorkspace";

export default function MobileWorkerPage() {
  const params = useParams<{ podKey: string }>();
  const podKey = typeof params.podKey === "string" ? params.podKey : "";

  return <MobilePodWorkspace podKey={podKey} />;
}
