import { Check, HelpCircle, X } from "lucide-react";
import { useState } from "react";

import { useAgentWorkspaceText } from "./AgentWorkspaceLocaleContext";
import type {
  AgentPermissionQuestion,
  AgentQuestionPermissionRequest,
} from "./agentPermissionContracts";

interface PermissionQuestionFormProps {
  disabled: boolean;
  permission: AgentQuestionPermissionRequest;
  onReject: () => void;
  onSubmit: (answers: Record<string, string[]>) => void;
}

export function PermissionQuestionForm({
  disabled,
  permission,
  onReject,
  onSubmit,
}: PermissionQuestionFormProps) {
  const [selected, setSelected] = useState<Record<string, string[]>>({});
  const [custom, setCustom] = useState<Record<string, string>>({});
  const text = useAgentWorkspaceText();
  const answers = permission.questions.reduce<Record<string, string[]>>(
    (result, question) => {
      result[question.id] = questionAnswers(
        question,
        selected[question.id] ?? [],
        custom[question.id] ?? "",
      );
      return result;
    },
    {},
  );
  const allAnswered =
    permission.questions.length > 0 &&
    permission.questions.every((question) => answers[question.id].length > 0);

  const selectOption = (question: AgentPermissionQuestion, label: string) => {
    setSelected((current) => {
      if (!question.multiple) {
        return { ...current, [question.id]: [label] };
      }
      const next = new Set(current[question.id] ?? []);
      if (next.has(label)) next.delete(label);
      else next.add(label);
      return { ...current, [question.id]: [...next] };
    });
    if (!question.multiple) {
      setCustom((current) => ({ ...current, [question.id]: "" }));
    }
  };

  const updateCustom = (question: AgentPermissionQuestion, value: string) => {
    setCustom((current) => ({ ...current, [question.id]: value }));
    if (!question.multiple && value.trim()) {
      setSelected((current) => ({ ...current, [question.id]: [] }));
    }
  };

  return (
    <section className="border-t border-border bg-muted/30 px-3 py-3">
      <div className="mb-3 flex items-center gap-2">
        <HelpCircle className="size-4 shrink-0 text-muted-foreground" />
        <h2 className="text-sm font-medium">{permission.title}</h2>
      </div>
      <form
        className="space-y-3"
        onSubmit={(event) => {
          event.preventDefault();
          if (allAnswered) onSubmit(answers);
        }}
      >
        <fieldset className="contents" disabled={disabled}>
          {permission.questions.map((question) => (
            <QuestionField
              custom={custom[question.id] ?? ""}
              key={question.id}
              onCustomChange={(value) => updateCustom(question, value)}
              onOptionChange={(label) => selectOption(question, label)}
              question={question}
              selected={selected[question.id] ?? []}
            />
          ))}
        </fieldset>
        <div className="flex justify-end gap-2">
          <button
            className="inline-flex h-11 items-center gap-1.5 rounded-md border border-border px-3 text-xs font-medium"
            disabled={disabled}
            onClick={onReject}
            type="button"
          >
            <X className="size-3.5" />
            {text.reject}
          </button>
          <button
            className="inline-flex h-11 items-center gap-1.5 rounded-md bg-primary px-3 text-xs font-medium text-primary-foreground disabled:cursor-not-allowed disabled:opacity-50"
            disabled={disabled || !allAnswered}
            type="submit"
          >
            <Check className="size-3.5" />
            {text.submitAnswers}
          </button>
        </div>
      </form>
    </section>
  );
}

function QuestionField({
  custom,
  onCustomChange,
  onOptionChange,
  question,
  selected,
}: {
  custom: string;
  onCustomChange: (value: string) => void;
  onOptionChange: (label: string) => void;
  question: AgentPermissionQuestion;
  selected: string[];
}) {
  const text = useAgentWorkspaceText();
  return (
    <fieldset className="space-y-2 rounded-md border border-border bg-background p-3">
      <legend className="px-1 text-xs font-medium text-muted-foreground">
        {question.header}
      </legend>
      <p className="text-sm font-medium">{question.prompt}</p>
      <div className="space-y-1.5">
        {question.options.map((option) => (
          <label
            className="flex cursor-pointer items-start gap-2 rounded-md border border-border px-2.5 py-2 text-sm hover:bg-muted/60"
            key={option.label}
          >
            <input
              aria-label={option.label}
              checked={selected.includes(option.label)}
              className="mt-0.5 accent-primary"
              name={question.id}
              onChange={() => onOptionChange(option.label)}
              type={question.multiple ? "checkbox" : "radio"}
              value={option.label}
            />
            <span className="min-w-0">
              <span className="block font-medium">{option.label}</span>
              <span className="block text-xs text-muted-foreground">
                {option.description}
              </span>
            </span>
          </label>
        ))}
        {question.allowCustom && (
          <label className="block space-y-1 text-xs font-medium text-muted-foreground">
            {text.customAnswer}
            <input
              aria-label={text.customAnswerFor(question.prompt)}
              autoComplete="off"
              className="h-11 w-full rounded-md border border-input bg-background px-2.5 text-sm text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring"
              onChange={(event) => onCustomChange(event.target.value)}
              type={question.secret ? "password" : "text"}
              value={custom}
            />
          </label>
        )}
      </div>
    </fieldset>
  );
}

function questionAnswers(
  question: AgentPermissionQuestion,
  selected: string[],
  custom: string,
): string[] {
  const customValue = question.allowCustom ? custom.trim() : "";
  if (!question.multiple) {
    if (customValue) return [customValue];
    return selected.slice(0, 1);
  }
  return customValue ? [...new Set([...selected, customValue])] : selected;
}
