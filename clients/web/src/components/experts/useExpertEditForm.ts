"use client";

import { useState, useEffect, useCallback } from "react";
import { useExpertStore } from "@/stores/expert";
import { readCurrentOrg } from "@/stores/auth";
import { listAgents, type AgentData } from "@/lib/api/facade/agentConnect";
import type { Expert } from "@/lib/api/expertApi";
import {
  EMPTY_EXPERT_FORM,
  ExpertConfigJsonError,
  buildExpertConfig,
  expertToForm,
  isValidConfigOverrides,
  slugifyExpert,
  type ExpertFormState,
} from "./expertFormModel";

export type { ExpertFormState } from "./expertFormModel";

const CONFIG_JSON_ERROR = "config_overrides_invalid_json";

export function useExpertEditForm(open: boolean, expert: Expert | null) {
  const isEdit = expert != null;
  const createExpert = useExpertStore((s) => s.createExpert);
  const updateExpert = useExpertStore((s) => s.updateExpert);

  const [form, setForm] = useState<ExpertFormState>(EMPTY_EXPERT_FORM);
  const [slugTouched, setSlugTouched] = useState(false);
  const [agents, setAgents] = useState<AgentData[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setForm(expert ? expertToForm(expert) : EMPTY_EXPERT_FORM);
    setSlugTouched(false);
    setError(null);
  }, [open, expert]);

  useEffect(() => {
    if (!open || isEdit) return;
    let cancelled = false;
    (async () => {
      try {
        const res = await listAgents(readCurrentOrg()?.slug ?? "");
        if (cancelled) return;
        const seen = new Set<string>();
        const list: AgentData[] = [];
        for (const a of [...res.builtin_agents, ...res.custom_agents, ...res.agents]) {
          if (seen.has(a.slug)) continue;
          seen.add(a.slug);
          list.push(a);
        }
        setAgents(list);
        setForm((f) => (f.agentSlug ? f : { ...f, agentSlug: list[0]?.slug ?? "" }));
      } catch {
        /* leave agent select empty; submit stays gated on agentSlug */
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [open, isEdit]);

  const patch = useCallback((next: Partial<ExpertFormState>) => {
    setForm((f) => {
      const merged = { ...f, ...next };
      if (next.name !== undefined && !slugTouched && !isEdit) {
        merged.slug = slugifyExpert(next.name);
      }
      return merged;
    });
  }, [slugTouched, isEdit]);

  const setSlug = useCallback((slug: string) => {
    setSlugTouched(true);
    setForm((f) => ({ ...f, slug }));
  }, []);

  const configValid = isValidConfigOverrides(form.configOverrides);
  const baseValid = isEdit
    ? form.name.trim().length > 0
    : form.name.trim().length > 0 && form.slug.trim().length > 0 && form.agentSlug.trim().length > 0;
  const canSubmit = baseValid && configValid;

  const submit = useCallback(async (): Promise<Expert | null> => {
    if (!canSubmit) return null;
    setSubmitting(true);
    setError(null);
    try {
      const config = buildExpertConfig(form);
      if (isEdit && expert) {
        return await updateExpert(expert.slug, config);
      }
      return await createExpert({
        ...config,
        slug: form.slug.trim(),
        agent_slug: form.agentSlug,
      });
    } catch (e) {
      if (e instanceof ExpertConfigJsonError) setError(CONFIG_JSON_ERROR);
      else setError(e instanceof Error ? e.message : String(e));
      return null;
    } finally {
      setSubmitting(false);
    }
  }, [canSubmit, form, isEdit, expert, updateExpert, createExpert]);

  return { form, patch, setSlug, agents, isEdit, canSubmit, submitting, error, submit };
}
