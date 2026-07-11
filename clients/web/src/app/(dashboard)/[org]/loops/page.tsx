"use client";

import { useParams } from "next/navigation";
import { GoalLoopPage } from "@/components/goal-loops/GoalLoopPage";

export default function LoopsPage() {
  const params = useParams();
  return <GoalLoopPage orgSlug={params.org as string} />;
}
