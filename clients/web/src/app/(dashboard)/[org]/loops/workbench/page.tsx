"use client";

import dynamic from "next/dynamic";
import { useParams } from "next/navigation";

const LoopWorkbench = dynamic(
  () => import("@/components/loop-builder/loop-workbench").then((module) => module.LoopWorkbench),
  { ssr: false },
);

export default function LoopPage() {
  const params = useParams();
  return <LoopWorkbench orgSlug={params.org as string} />;
}
