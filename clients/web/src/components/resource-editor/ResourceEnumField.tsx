"use client";

import { FormField } from "@/components/ui/form-field";

interface ResourceEnumFieldProps {
  id: string;
  label: string;
  value: string;
  options: Array<{ value: string; label: string }>;
  onChange: (value: string) => void;
}

export function ResourceEnumField({
  id,
  label,
  value,
  options,
  onChange,
}: ResourceEnumFieldProps) {
  return (
    <FormField label={label} htmlFor={id}>
      <select
        id={id}
        className="h-9 w-full rounded-md bg-surface-raised px-3 text-sm ring-1 ring-border/35 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/35"
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </FormField>
  );
}
