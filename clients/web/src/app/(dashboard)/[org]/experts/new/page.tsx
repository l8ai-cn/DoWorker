"use client";

import { useParams, useRouter } from "next/navigation";
import { ResourceEditorShell } from "@/components/resource-editor/ResourceEditorShell";
import { appliedResource } from "@/components/resource-editor/resource-apply-result";

export default function NewExpertPage() {
  const params = useParams();
  const router = useRouter();
  const orgSlug = String(params.org ?? "");
  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto w-full max-w-5xl px-6 py-6">
        <ResourceEditorShell
          orgSlug={orgSlug}
          kind="Expert"
          onApplied={(result) => {
            const name = appliedResource(result)?.identity?.target?.name;
            if (name) router.push(`/${orgSlug}/experts/${name}`);
          }}
        />
      </div>
    </div>
  );
}
