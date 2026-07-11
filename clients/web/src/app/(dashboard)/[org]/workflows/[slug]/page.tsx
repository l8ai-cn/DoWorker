"use client";

import { useParams } from "next/navigation";
import { WorkflowDetailPane } from "@/components/workflows/WorkflowDetailPane";

export default function WorkflowDetailPage() {
  const params = useParams();
  const slug = params.slug as string;
  const orgSlug = params.org as string;

  return <WorkflowDetailPane slug={slug} orgSlug={orgSlug} />;
}
