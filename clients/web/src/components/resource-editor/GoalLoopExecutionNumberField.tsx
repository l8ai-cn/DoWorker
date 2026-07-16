"use client";

import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import type { GoalLoopIntegerDraft } from "./resource-editor-types";

interface GoalLoopExecutionNumberFieldProps {
  id: string;
  label: string;
  value: GoalLoopIntegerDraft;
  min: number;
  max?: number;
  error?: string;
  required?: boolean;
  onChange: (value: string) => void;
}

export function GoalLoopExecutionNumberField({
  id,
  label,
  value,
  min,
  max,
  error,
  required,
  onChange,
}: GoalLoopExecutionNumberFieldProps) {
  return (
    <FormField
      label={label}
      htmlFor={id}
      className="flex-1"
      error={error}
      required={required}
    >
      <Input
        id={id}
        type="number"
        step={1}
        min={min}
        max={max}
        value={value}
        aria-invalid={Boolean(error)}
        onChange={(event) => onChange(event.target.value)}
      />
    </FormField>
  );
}
