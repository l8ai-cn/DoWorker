"use client";

import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";

interface GoalLoopExecutionNumberFieldProps {
  id: string;
  label: string;
  value: number | "";
  min: number;
  max?: number;
  onChange: (value: string) => void;
}

export function GoalLoopExecutionNumberField({
  id,
  label,
  value,
  min,
  max,
  onChange,
}: GoalLoopExecutionNumberFieldProps) {
  return (
    <FormField label={label} htmlFor={id} className="flex-1">
      <Input
        id={id}
        type="number"
        min={min}
        max={max}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
    </FormField>
  );
}
