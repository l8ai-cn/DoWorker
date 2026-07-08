"use client";

import { useEffect } from "react";
import { useParams } from "next/navigation";
import { ExpertDetailPane } from "@/components/experts/ExpertDetailPane";
import { useExpertStore } from "@/stores/expert";

export default function ExpertDetailPage() {
  const params = useParams();
  const slug = params.slug as string;
  const orgSlug = params.org as string;
  const fetchExpert = useExpertStore((s) => s.fetchExpert);
  const clearCurrentExpert = useExpertStore((s) => s.clearCurrentExpert);

  useEffect(() => {
    fetchExpert(slug);
    return () => clearCurrentExpert();
  }, [slug, fetchExpert, clearCurrentExpert]);

  return <ExpertDetailPane slug={slug} orgSlug={orgSlug} />;
}
