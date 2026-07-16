"use client";

import { FormField } from "@/components/ui/form-field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";

export interface WorkerTemplateOption {
  value: string;
  label: string;
  selectable: boolean;
  blockingReason: string;
}

interface WorkerTemplateOptionSelectFieldProps {
  id: string;
  label: string;
  value: string;
  options: WorkerTemplateOption[];
  disabled?: boolean;
  onChange: (value: string) => void;
}

export function WorkerTemplateOptionSelectField({
  id,
  label,
  value,
  options,
  disabled,
  onChange,
}: WorkerTemplateOptionSelectFieldProps) {
  const selected = options.find((option) => option.value === value);
  const selectedLabel = selected?.label ?? (value || label);

  return (
    <FormField
      label={label}
      htmlFor={id}
      required
      className="flex-1"
      error={!selected?.selectable && selected?.blockingReason
        ? selected.blockingReason
        : undefined}
    >
      <Select value={value} onValueChange={onChange} disabled={disabled}>
        <SelectTrigger id={id} aria-label={label}>
          <span className={!selected ? "text-muted-foreground" : undefined}>
            {selectedLabel}
          </span>
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem
              key={option.value}
              value={option.value}
              disabled={!option.selectable}
              aria-disabled={!option.selectable}
            >
              <span className="flex min-w-0 flex-col">
                <span>{option.label}</span>
                {!option.selectable && option.blockingReason && (
                  <span className="text-xs text-muted-foreground">
                    {option.blockingReason}
                  </span>
                )}
              </span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </FormField>
  );
}
