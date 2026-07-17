import type { ReactNode } from "react";
import type { LoopProgram } from "@proto/goalloop/v1/goalloop_pb";
import type { LoopAIProjectionMessages } from "./loop-workbench-messages";

interface LoopProgramExplanationProps {
  program?: LoopProgram;
  valid: boolean;
  messages: LoopAIProjectionMessages;
}

export function LoopProgramExplanation({
  program,
  valid,
  messages,
}: LoopProgramExplanationProps) {
  if (!valid || !program) {
    return (
      <p className="rounded-md border border-warning/30 bg-warning/10 px-4 py-3 text-sm">
        {messages.unavailable}
      </p>
    );
  }

  const limits = program.limits;
  const repeat = program.repeat;
  const until = repeat?.until;
  const untilValue = until?.localId && until.field
    ? `${until.localId}.${until.field}`
    : messages.empty;

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-semibold">{messages.title}</h3>
      <div className="divide-y divide-border overflow-hidden rounded-md border border-border">
        <DetailSection title={messages.overview}>
          <DetailRow label={messages.schemaVersion} value={String(program.schemaVersion)} />
          <DetailRow label={messages.loopName} value={text(program.loop?.localId, messages)} />
        </DetailSection>
        <DetailSection title={messages.limits}>
          <DetailRow label={messages.iterations} value={number(limits?.iterations, messages.iterationsValue, messages)} />
          <DetailRow label={messages.tokens} value={number(limits?.tokens, messages.tokensValue, messages)} />
          <DetailRow label={messages.timeout} value={number(limits?.timeoutMinutes, messages.minutesValue, messages)} />
          <DetailRow label={messages.noProgress} value={number(limits?.noProgress, messages.timesValue, messages)} />
          <DetailRow label={messages.sameError} value={number(limits?.sameError, messages.timesValue, messages)} />
        </DetailSection>
        <DetailSection title={messages.repeat}>
          <DetailRow label={messages.repeatName} value={text(repeat?.identity?.localId, messages)} />
          <DetailRow label={messages.repeatMax} value={number(repeat?.max, messages.timesValue, messages)} />
          <DetailRow label={messages.until} value={untilValue} />
        </DetailSection>
        <DetailSection title={messages.agent}>
          <DetailRow label={messages.agentName} value={text(repeat?.agent?.identity?.localId, messages)} />
          <DetailRow label={messages.agentPrompt} value={text(repeat?.agent?.prompt, messages)} wide />
        </DetailSection>
        <DetailSection title={messages.verifier}>
          <DetailRow label={messages.verifierName} value={text(repeat?.verifier?.identity?.localId, messages)} />
          <DetailRow label={messages.verifierCommand} value={text(repeat?.verifier?.command, messages)} wide />
          <DetailRow label={messages.verifierAccept} value={text(repeat?.verifier?.accept, messages)} wide />
        </DetailSection>
        <DetailSection title={messages.failure}>
          <DetailRow label={messages.failurePolicy} value={messages.failurePolicyLabel(program.failurePolicy)} />
        </DetailSection>
      </div>
    </div>
  );
}

function number(
  value: bigint | undefined,
  format: (value: string) => string,
  messages: LoopAIProjectionMessages,
) {
  return value === undefined ? messages.empty : format(String(value));
}

function text(value: string | undefined, messages: LoopAIProjectionMessages) {
  return value?.trim() || messages.empty;
}

function DetailSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="grid gap-3 bg-background px-4 py-3 md:grid-cols-[9rem_minmax(0,1fr)]">
      <h4 className="text-xs font-semibold text-muted-foreground">{title}</h4>
      <dl className="grid min-w-0 gap-2 sm:grid-cols-2">{children}</dl>
    </section>
  );
}

function DetailRow({
  label,
  value,
  wide = false,
}: {
  label: string;
  value: string;
  wide?: boolean;
}) {
  return (
    <div className={wide ? "min-w-0 sm:col-span-2" : "min-w-0"}>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-0.5 whitespace-pre-wrap break-words text-sm">{value}</dd>
    </div>
  );
}
