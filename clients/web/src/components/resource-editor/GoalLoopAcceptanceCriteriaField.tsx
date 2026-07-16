"use client";

import { Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Textarea } from "@/components/ui/textarea";

interface GoalLoopAcceptanceCriteriaFieldProps {
  value: string[];
  onChange: (value: string[]) => void;
}

export function GoalLoopAcceptanceCriteriaField({
  value,
  onChange,
}: GoalLoopAcceptanceCriteriaFieldProps) {
  const t = useTranslations("resourceEditor");
  const addLabel = t("collections.addAcceptanceCriteria");
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <h4 className="text-sm font-medium">
          {t("fields.acceptanceCriteria")}
        </h4>
        <Button
          type="button"
          variant="outline"
          size="icon"
          title={addLabel}
          aria-label={addLabel}
          onClick={() => onChange([...value, ""])}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      {value.length === 0 && (
        <p className="text-sm text-muted-foreground">
          {t("collections.none")}
        </p>
      )}
      {value.map((criterion, index) => {
        const label = t("collections.acceptanceCriterion", {
          index: index + 1,
        });
        const removeLabel = t("collections.removeAcceptanceCriterion", {
          index: index + 1,
        });
        return (
          <div
            key={`acceptance-criterion-${index}`}
            className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_2.5rem]"
          >
            <FormField
              label={label}
              htmlFor={`goal-loop-acceptance-${index}`}
            >
              <Textarea
                id={`goal-loop-acceptance-${index}`}
                value={criterion}
                onChange={(event) => onChange(value.map((item, itemIndex) =>
                  itemIndex === index ? event.target.value : item))}
              />
            </FormField>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="self-start sm:mt-7"
              title={removeLabel}
              aria-label={removeLabel}
              onClick={() => onChange(value.filter((_, item) => item !== index))}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        );
      })}
    </div>
  );
}
