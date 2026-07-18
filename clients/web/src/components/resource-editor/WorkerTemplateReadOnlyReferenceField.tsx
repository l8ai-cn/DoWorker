"use client";

import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import type { ResourceReference } from "./resource-editor-types";

interface WorkerTemplateReadOnlyReferenceFieldProps {
  id: string;
  label: string;
  revisionLabel: string;
  value: ResourceReference;
}

export function WorkerTemplateReadOnlyReferenceField({
  id,
  label,
  revisionLabel,
  value,
}: WorkerTemplateReadOnlyReferenceFieldProps) {
  return (
    <FormField label={label} htmlFor={id}>
      <div className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_7rem]">
        <Input
          id={id}
          value={value.name}
          readOnly
          aria-readonly
          className="bg-muted/40 text-muted-foreground"
        />
        <Input
          id={`${id}-revision`}
          type="number"
          aria-label={`${label} ${revisionLabel}`}
          value={value.revision ?? ""}
          readOnly
          aria-readonly
          className="bg-muted/40 text-muted-foreground"
        />
      </div>
    </FormField>
  );
}
