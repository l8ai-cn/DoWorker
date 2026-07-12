import { useRouter } from "@tanstack/react-router";
import { InteractionModeSelector } from "@/components/interaction-mode-selector";
import {
  NewTaskGoalPanel,
  buildCodexObjective,
  parseTokenBudget,
} from "@/components/new-task-goal-panel";
import { MobileFrame } from "@/components/mobile-frame";
import { useAvailableAgents } from "@/hooks/useAvailableAgents";
import { useIsAuthed } from "@/hooks/useIsAuthed";
import { useLiveExperts } from "@/hooks/useLiveExperts";
import { useLiveProjects } from "@/hooks/useLiveProjects";
import { setCodexGoal } from "@/lib/codex-goal-api";
import { dispatchExpertBySlug } from "@/lib/expert-run-dispatch";
import { messageWithExpertContext, taskTitleForSubmit } from "@/lib/expert-agent-slugs";
import { resolveMobileWorkerSelection } from "@/lib/mobile-worker-selection";
import { projectNameFromId } from "@/lib/project-label";
import { assignSessionProject, createSession } from "@/lib/sessions-api";
import { ExpertPicker } from "./expert-picker";
import { NewTaskHeader } from "./new-task-header";
import { ProjectPicker } from "./project-picker";
import { TaskPrompt } from "./task-prompt";
import { TaskShortcuts } from "./task-shortcuts";
import {
  availableAcpExperts,
  findSelectedExpert,
  type NewTaskSearch,
  useNewTaskState,
} from "./use-new-task-state";
import { WorkerPicker } from "./worker-picker";

export function NewTaskPage({ search }: { search: NewTaskSearch }) {
  const router = useRouter();
  const authed = useIsAuthed();
  const availableAgents = useAvailableAgents();
  const liveExperts = useLiveExperts();
  const liveProjects = useLiveProjects();
  const projectNames = authed ? liveProjects.names : ["默认项目"];
  const form = useNewTaskState({
    search,
    engines: availableAgents.agents,
    projectNames,
    initialProjectName: search.project
      ? projectNameFromId(search.project)
      : (projectNames[0] ?? "默认项目"),
  });
  const worker = resolveMobileWorkerSelection(
    availableAgents.agents,
    form.engineID,
    authed,
    availableAgents.loading,
    authed ? availableAgents.error : null,
  );
  const experts = availableAcpExperts(liveExperts.items, availableAgents.agents);
  const expert = findSelectedExpert(experts, form.expertSlug);
  const expertWorker = expert
    ? availableAgents.agents.find((agent) => agent.id === expert.agent_slug)
    : undefined;

  async function submit() {
    if (form.submitting || !form.prompt.trim()) return;
    if (!authed) {
      router.navigate({ to: "/login" });
      return;
    }
    if (!worker.current) {
      form.setSubmitError(worker.message);
      return;
    }
    form.setSubmitting(true);
    form.setSubmitError(null);
    try {
      if (form.asGoal) {
        const codex = availableAgents.agents.find((agent) => agent.id === "codex-cli");
        if (!codex?.supportedModes.includes("acp")) {
          throw new Error("当前组织没有支持可视化对话的 Codex Worker");
        }
        const budget = parseTokenBudget(form.tokenBudget);
        const session = await createSession(
          codex,
          taskTitleForSubmit(form.prompt, undefined, "Codex Goal"),
          undefined,
          { mode: "acp" },
        );
        await setCodexGoal(session.id, {
          objective: buildCodexObjective(form.prompt, form.successCriteria),
          ...(budget !== undefined ? { tokenBudget: budget } : {}),
          status: form.goalMode,
        });
        if (form.projectName.trim()) await assignSessionProject(session.id, form.projectName);
        router.navigate({ to: "/sessions/$sessionId", params: { sessionId: session.id } });
        return;
      }

      const text = form.prompt.trim();
      if (expert && form.interactionMode === "acp") {
        if (!expertWorker) {
          throw new Error(`专家 ${expert.name} 对应的 Worker 不可用`);
        }
        const dispatched = await dispatchExpertBySlug(expert.slug, text, expertWorker);
        if (dispatched) {
          if (dispatched.kind === "session") {
            if (form.projectName.trim())
              await assignSessionProject(dispatched.sessionId, form.projectName);
            router.navigate({
              to: "/sessions/$sessionId",
              params: { sessionId: dispatched.sessionId },
            });
          } else {
            form.setSubmitError(dispatched.message);
          }
          return;
        }
      }
      const targetWorker = expertWorker ?? worker.current;
      if (!targetWorker) {
        throw new Error("当前 Worker 不可用");
      }
      if (!targetWorker.supportedModes.includes(form.interactionMode)) {
        throw new Error(`${targetWorker.name} 不支持当前交互方式`);
      }
      const session = await createSession(
        targetWorker,
        taskTitleForSubmit(text, expert?.name, targetWorker.name),
        messageWithExpertContext(text, expert?.name, expert?.description ?? expert?.prompt),
        { mode: form.interactionMode },
      );
      if (form.projectName.trim()) await assignSessionProject(session.id, form.projectName);
      router.navigate({
        to:
          form.interactionMode === "pty" ? "/sessions/$sessionId/terminal" : "/sessions/$sessionId",
        params: { sessionId: session.id },
      });
    } catch (error) {
      form.setSubmitError(error instanceof Error ? error.message : "创建任务失败");
    } finally {
      form.setSubmitting(false);
    }
  }

  return (
    <MobileFrame>
      <div className="flex flex-col">
        <NewTaskHeader asGoal={form.asGoal} onGoalChange={form.setAsGoal} />
        <div className="space-y-3 px-5 py-4 pb-2">
          <WorkerPicker
            disabled={form.asGoal}
            engines={availableAgents.agents}
            current={worker.current}
            message={worker.message}
            open={form.enginePickerOpen}
            selectedID={form.engineID}
            onOpenChange={form.setEnginePickerOpen}
            onSelect={form.selectWorker}
          />
          <InteractionModeSelector
            mode={form.interactionMode}
            supportedModes={worker.current?.supportedModes ?? []}
            disabled={form.asGoal || !worker.current}
            onChange={form.setInteractionMode}
          />
          <ExpertPicker
            authenticated={authed}
            disabled={form.asGoal || form.interactionMode === "pty"}
            experts={authed ? experts : []}
            current={expert}
            open={form.expertPickerOpen}
            selectedSlug={form.expertSlug}
            onOpenChange={form.setExpertPickerOpen}
            onSelect={(slug) => {
              form.setExpertSlug(slug);
              const selected = findSelectedExpert(experts, slug);
              if (selected) form.setEngineID(selected.agent_slug);
            }}
          />
          <ProjectPicker
            names={projectNames}
            open={form.projectPickerOpen}
            selectedName={form.projectName}
            onOpenChange={form.setProjectPickerOpen}
            onSelect={form.setProjectName}
          />
          <TaskPrompt
            asGoal={form.asGoal}
            authenticated={authed}
            currentExpert={expert}
            currentWorkerAvailable={Boolean(worker.current)}
            prompt={form.prompt}
            slashToken={form.slashToken}
            submitting={form.submitting}
            error={form.submitError}
            onPromptChange={form.onPromptChange}
            onSlashTokenChange={form.setSlashToken}
            onSlashPick={form.applySlash}
            onSubmit={() => void submit()}
          />
          {form.asGoal && (
            <NewTaskGoalPanel
              tokenBudget={form.tokenBudget}
              goalMode={form.goalMode}
              successCriteria={form.successCriteria}
              onTokenBudgetChange={form.setTokenBudget}
              onGoalModeChange={form.setGoalMode}
              onSuccessCriteriaChange={form.setSuccessCriteria}
            />
          )}
          <TaskShortcuts
            asGoal={form.asGoal}
            expert={expert}
            prompt={form.prompt}
            onPromptChange={form.onPromptChange}
          />
        </div>
      </div>
    </MobileFrame>
  );
}
