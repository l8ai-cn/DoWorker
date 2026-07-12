import { Link, createFileRoute, useRouter } from "@tanstack/react-router";
import { ArrowLeft, Check, ChevronDown, ImagePlus, Loader2, Target, X, Zap } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { SlashMenu, SLASH_COMMANDS, detectSlashToken, type SlashCommand } from "@/components/slash-menu";
import { MobileFrame } from "@/components/mobile-frame";
import {
  NewTaskGoalPanel,
  buildCodexObjective,
  parseTokenBudget,
  type CodexGoalMode,
} from "@/components/new-task-goal-panel";
import { Lightbox } from "@/components/lightbox";
import { InteractionModeSelector } from "@/components/interaction-mode-selector";
import { useAvailableAgents } from "@/hooks/useAvailableAgents";
import { useLiveExperts } from "@/hooks/useLiveExperts";
import { useLiveProjects } from "@/hooks/useLiveProjects";
import { cn } from "@/lib/utils";
import { pageTitle } from "@/lib/app-brand";
import { useIsAuthed } from "@/hooks/useIsAuthed";
import { messageWithExpertContext, taskTitleForSubmit } from "@/lib/expert-agent-slugs";
import { dispatchExpertBySlug } from "@/lib/expert-run-dispatch";
import { assignSessionProject, createSession, type SessionInteractionMode } from "@/lib/sessions-api";
import { setCodexGoal } from "@/lib/codex-goal-api";
import { localProjectMeta } from "@/lib/projects-local";
import { projectNameFromId } from "@/lib/project-label";
import type { LiveExpert } from "@/lib/experts-api";
import { resolveMobileWorkerSelection } from "@/lib/mobile-worker-selection";

export const Route = createFileRoute("/new")({
  validateSearch: (s: Record<string, unknown>): { project?: string; expert?: string; prompt?: string } => ({
    project: typeof s.project === "string" ? s.project : undefined,
    expert: typeof s.expert === "string" ? s.expert : undefined,
    prompt: typeof s.prompt === "string" ? s.prompt : undefined,
  }),
  head: () => ({ meta: [{ title: pageTitle("下发新任务") }] }),
  component: NewTask,
});


const templates = [
  "修一下 CI 里最新失败的那个 test",
  "review 我最近的 PR 并给出改进建议",
  "把 README 翻译成日语",
];

const goalTemplates = [
  "保持主分支 CI 常绿",
  "持续处理依赖安全告警",
  "维护 README 与 API 文档同步",
];
const DEFAULT_PROJECT_NAMES = ["默认项目"];

function NewTask() {
  const router = useRouter();
  const search = Route.useSearch();
  const authed = useIsAuthed();
  const availableAgents = useAvailableAgents();
  const liveExperts = useLiveExperts();
  const liveProjects = useLiveProjects();
  const engines = availableAgents.agents;
  const projectNames = authed ? liveProjects.names : DEFAULT_PROJECT_NAMES;

  const [engineId, setEngineId] = useState<string>("codex-cli");
  const [enginePickerOpen, setEnginePickerOpen] = useState(false);
  const [expertSlug, setExpertSlug] = useState<string | undefined>(search.expert);
  const [expertPickerOpen, setExpertPickerOpen] = useState(false);
  const [prompt, setPrompt] = useState(search.prompt ?? "");
  const [projectName, setProjectName] = useState(
    search.project ? projectNameFromId(search.project) : (projectNames[0] ?? "默认项目"),
  );
  const [pickerOpen, setPickerOpen] = useState(false);
  const [images, setImages] = useState<{ id: string; url: string; name: string }[]>([]);
  const [asGoal, setAsGoal] = useState(false);
  const [interactionMode, setInteractionMode] = useState<SessionInteractionMode>("acp");
  const [tokenBudget, setTokenBudget] = useState("");
  const [goalMode, setGoalMode] = useState<CodexGoalMode>("active");
  const [successCriteria, setSuccessCriteria] = useState("");
  const [slashToken, setSlashToken] = useState<{ start: number; token: string } | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const workerSelection = resolveMobileWorkerSelection(
    engines,
    engineId,
    authed,
    availableAgents.loading,
    authed ? availableAgents.error : null,
  );
  const currentEngine = workerSelection.current;
  const currentExpert = liveExperts.items.find((e) => e.slug === expertSlug);
  const projectMeta = localProjectMeta(projectName);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (currentEngine && currentEngine.id !== engineId) {
      setEngineId(currentEngine.id);
    }
  }, [currentEngine, engineId]);

  useEffect(() => {
    if (!currentEngine || currentEngine.supportedModes.includes(interactionMode)) return;
    setInteractionMode(currentEngine.supportedModes[0]);
  }, [currentEngine, interactionMode]);

  useEffect(() => {
    if (projectNames.length > 0 && !projectNames.includes(projectName)) {
      setProjectName(projectNames[0]);
    }
  }, [projectNames, projectName]);

  useEffect(() => {
    if (!asGoal) return;
    setEngineId("codex-cli");
    setExpertSlug(undefined);
    setInteractionMode("acp");
  }, [asGoal]);

  useEffect(() => {
    if (interactionMode === "pty") setExpertSlug(undefined);
  }, [interactionMode]);

  const onPromptChange = (v: string, caret: number) => {
    setPrompt(v);
    setSlashToken(detectSlashToken(v, caret));
  };

  const applySlash = (cmd: SlashCommand) => {
    if (!slashToken) return;
    const before = prompt.slice(0, slashToken.start);
    const after = prompt.slice(slashToken.start + slashToken.token.length);
    const insert = cmd.hint?.startsWith(cmd.cmd + " ") ? cmd.cmd + " " : cmd.cmd + " ";
    const next = before + insert + after;
    setPrompt(next);
    setSlashToken(null);
    // side-effects for meta commands
    if (cmd.cmd === "/goal") setAsGoal(true);
    if (cmd.cmd === "/attach") fileInputRef.current?.click();
    requestAnimationFrame(() => {
      const el = textareaRef.current;
      if (!el) return;
      el.focus();
      const pos = (before + insert).length;
      el.setSelectionRange(pos, pos);
    });
  };

  const onPickImages = (files: FileList | null) => {
    if (!files) return;
    Array.from(files).slice(0, 6).forEach((file) => {
      if (!file.type.startsWith("image/")) return;
      const reader = new FileReader();
      reader.onload = () => {
        setImages((prev) => [
          ...prev,
          { id: `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`, url: reader.result as string, name: file.name },
        ]);
      };
      reader.readAsDataURL(file);
    });
  };

  const submit = async () => {
    if (submitting) return;
    if (asGoal ? !prompt.trim() : !prompt.trim() && images.length === 0) return;
    if (!authed) {
      router.navigate({ to: "/login" });
      return;
    }
    if (!currentEngine) {
      setSubmitError(workerSelection.message);
      return;
    }

    setSubmitting(true);
    setSubmitError(null);
    try {
      if (asGoal) {
        if (!engines.some((engine) => engine.id === "codex-cli")) {
          throw new Error("当前组织没有可用于目标模式的 Codex Worker");
        }
        if (!engines.find((engine) => engine.id === "codex-cli")?.supportedModes.includes("acp")) {
          throw new Error("当前组织的 Codex Worker 不支持可视化对话模式");
        }
        const objective = buildCodexObjective(prompt, successCriteria);
        const budget = parseTokenBudget(tokenBudget);
        const title = taskTitleForSubmit(prompt, undefined, "Codex Goal");
        const session = await createSession("codex-cli", title, undefined, { mode: "acp" });
        await setCodexGoal(session.id, {
          objective,
          ...(budget !== undefined ? { tokenBudget: budget } : {}),
          status: goalMode,
        });
        if (projectName.trim()) {
          await assignSessionProject(session.id, projectName).catch(() => undefined);
        }
        router.navigate({ to: "/sessions/$sessionId", params: { sessionId: session.id } });
        return;
      }

      const text = (prompt.trim() || "分析附带的图片") + (images.length > 0 ? ` [附带 ${images.length} 张图片]` : "");
      if (expertSlug && interactionMode === "acp") {
        const dispatched = await dispatchExpertBySlug(expertSlug, text);
        if (dispatched) {
          if (dispatched.kind === "session") {
            if (projectName.trim()) {
              await assignSessionProject(dispatched.sessionId, projectName).catch(() => undefined);
            }
            router.navigate({ to: "/sessions/$sessionId", params: { sessionId: dispatched.sessionId } });
          } else {
            setSubmitError(dispatched.message);
          }
          return;
        }
      }
      const agentSlug = currentExpert?.agent_slug ?? engineId;
      if (!currentEngine.supportedModes.includes(interactionMode)) {
        throw new Error(`${currentEngine.name} 不支持当前交互方式`);
      }
      const fullText = messageWithExpertContext(text, currentExpert?.name, currentExpert?.description ?? currentExpert?.prompt);
      const title = taskTitleForSubmit(text, currentExpert?.name, currentEngine.name);
      const session = await createSession(agentSlug, title, fullText, { mode: interactionMode });
      if (projectName.trim()) {
        await assignSessionProject(session.id, projectName).catch(() => undefined);
      }
      router.navigate({
        to: interactionMode === "pty" ? "/sessions/$sessionId/terminal" : "/sessions/$sessionId",
        params: { sessionId: session.id },
      });
    } catch (e) {
      setSubmitError(e instanceof Error ? e.message : "创建任务失败");
    } finally {
      setSubmitting(false);
    }
  };


  return (
    <MobileFrame>
      <div className="flex flex-col">
        {/* Header */}
        <header className="safe-top sticky top-0 z-30 border-b border-border/60 bg-background/85 px-4 pb-3 pt-3 backdrop-blur-xl">
          <div className="flex items-center gap-2">
            <Link to="/" className="flex h-8 w-8 items-center justify-center rounded-full hover:bg-surface">
              <ArrowLeft className="h-4 w-4" />
            </Link>
            <h1 className="flex-1 text-[14px] font-semibold">新任务</h1>
            <button
              onClick={() => setAsGoal((v) => !v)}
              className={cn(
                "flex items-center gap-1 rounded-full px-2.5 py-1 text-[11px] font-medium transition",
                asGoal
                  ? "bg-primary text-primary-foreground glow-primary"
                  : "bg-surface text-muted-foreground ring-1 ring-border/50 hover:text-foreground",
              )}
            >
              <Target className="h-3 w-3" />
              作为目标
            </button>
          </div>
        </header>

        <div className="space-y-3 px-5 py-4 pb-2">
          {/* Agent engine selector (required) */}
          <div>
            <div className="mb-1.5 flex items-center justify-between px-1">
              <p className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">Agent 工具</p>
              <span className="text-[10px] text-muted-foreground/70">必选</span>
            </div>
            <button
              onClick={() => setEnginePickerOpen((o) => !o)}
              disabled={asGoal || !currentEngine}
              className={cn(
                "flex w-full items-center gap-3 rounded-2xl bg-card p-3 ring-1 ring-border/50 transition hover:ring-primary/40",
                asGoal && "opacity-70",
              )}
            >
              {currentEngine ? (
                <>
                  <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/15 text-lg ring-1 ring-white/5">
                    {currentEngine.avatar}
                  </span>
                  <div className="min-w-0 flex-1 text-left">
                    <p className="truncate text-[13.5px] font-semibold">{currentEngine.name}</p>
                    <p className="truncate text-[10.5px] text-muted-foreground">
                      {currentEngine.vendor} · {currentEngine.desc}
                    </p>
                  </div>
                  <ChevronDown className={cn("h-4 w-4 text-muted-foreground transition-transform", enginePickerOpen && "rotate-180")} />
                </>
              ) : (
                <p className="min-w-0 flex-1 text-left text-[12px] text-muted-foreground">
                  {workerSelection.message}
                </p>
              )}
            </button>
            {enginePickerOpen && !asGoal && (
              <div className="mt-1 grid grid-cols-2 gap-1.5 rounded-2xl bg-card p-1.5 ring-1 ring-border/50 stream-in">
                {engines.map((e) => (
                  <button
                    key={e.id}
                    onClick={() => { setEngineId(e.id); setEnginePickerOpen(false); }}
                    className={cn(
                      "flex items-center gap-2 rounded-xl px-2 py-2 text-left transition",
                      e.id === engineId ? "bg-primary/10 ring-1 ring-primary/40" : "hover:bg-surface",
                    )}
                  >
                    <span className="flex h-7 w-7 items-center justify-center rounded-lg bg-surface text-sm">{e.avatar}</span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-[12px] font-medium">{e.name}</p>
                      <p className="truncate text-[10px] text-muted-foreground">{e.vendor}</p>
                    </div>
                    {e.id === engineId && <Check className="h-3 w-3 shrink-0 text-primary" />}
                  </button>
                ))}
              </div>
            )}
          </div>

          <InteractionModeSelector
            mode={interactionMode}
            supportedModes={currentEngine?.supportedModes ?? []}
            disabled={asGoal || !currentEngine}
            onChange={setInteractionMode}
          />

          {/* Expert selector (optional) */}
          <div>
            <div className="mb-1.5 flex items-center justify-between px-1">
              <p className="text-[10.5px] font-semibold uppercase tracking-wider text-muted-foreground">执行专家</p>
              <span className="text-[10px] text-muted-foreground/70">可选</span>
            </div>
            <button
              onClick={() => setExpertPickerOpen((o) => !o)}
              disabled={asGoal || interactionMode === "pty"}
              className={cn(
                "flex w-full items-center gap-3 rounded-2xl bg-card p-3 ring-1 ring-border/50 transition hover:ring-primary/40",
                asGoal && "opacity-70",
              )}
            >
              <span
                className={cn(
                  "flex h-9 w-9 items-center justify-center rounded-xl text-lg ring-1 ring-white/5",
                  currentExpert ? "bg-primary/15" : "bg-surface",
                )}
              >
                {"🤖"}
              </span>
              <div className="min-w-0 flex-1 text-left">
                <p className="truncate text-[13.5px] font-semibold">
                  {currentExpert?.name ?? "不使用专家"}
                </p>
                <p className="truncate text-[10.5px] text-muted-foreground">
                  {currentExpert?.description ?? "叠加一位专家来注入领域能力"}
                </p>
              </div>
              <ChevronDown className={cn("h-4 w-4 text-muted-foreground transition-transform", expertPickerOpen && "rotate-180")} />
            </button>
          </div>

            {expertPickerOpen && !asGoal && interactionMode === "acp" && (
            <div className="-mt-2 max-h-64 space-y-1 overflow-y-auto rounded-2xl bg-card p-1.5 ring-1 ring-border/50 stream-in">
              <button
                onClick={() => { setExpertSlug(undefined); setExpertPickerOpen(false); }}
                className={cn(
                  "flex w-full items-center gap-3 rounded-xl px-2.5 py-2 text-left transition",
                  !expertSlug ? "bg-primary/10" : "hover:bg-surface",
                )}
              >
                <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-surface text-base">✨</span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-[13px] font-medium">通用 Agent</p>
                  <p className="truncate text-[10.5px] text-muted-foreground">不指定专家</p>
                </div>
                {!expertSlug && <Check className="h-3.5 w-3.5 text-primary" />}
              </button>
              {(authed ? liveExperts.items : []).map((e: LiveExpert) => (
                <button
                  key={e.slug}
                  onClick={() => { setExpertSlug(e.slug); setExpertPickerOpen(false); }}
                  className={cn(
                    "flex w-full items-center gap-3 rounded-xl px-2.5 py-2 text-left transition",
                    e.slug === expertSlug ? "bg-primary/10" : "hover:bg-surface",
                  )}
                >
                  <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/15 text-base">🤖</span>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-[13px] font-medium">{e.name}</p>
                    <p className="truncate text-[10.5px] text-muted-foreground">{e.description ?? e.agent_slug}</p>
                  </div>
                  {e.slug === expertSlug && <Check className="h-3.5 w-3.5 text-primary" />}
                </button>
              ))}
              {authed && liveExperts.items.length === 0 && (
                <p className="px-2 py-3 text-center text-[11px] text-muted-foreground">
                  组织暂无专家 · <Link to="/experts" className="text-primary">专家库</Link>
                </p>
              )}
            </div>
          )}

          {/* Project selector — compact pill */}
          <button
            onClick={() => setPickerOpen((o) => !o)}
            className="flex w-full items-center gap-3 rounded-2xl bg-card p-3 ring-1 ring-border/50 transition hover:ring-primary/40"
          >
            <span className={cn(
              "flex h-9 w-9 items-center justify-center rounded-xl ring-1 ring-white/5 bg-primary/15",
            )}>
              <span className="h-2 w-2 rounded-full bg-success" />
            </span>
            <div className="min-w-0 flex-1 text-left">
              <p className="text-[10.5px] uppercase tracking-wider text-muted-foreground">目标项目</p>
              <p className="truncate text-[13.5px] font-semibold">{projectName}</p>
              {projectMeta?.repo && (
                <p className="truncate font-mono text-[10.5px] text-muted-foreground">{projectMeta.repo}</p>
              )}
            </div>
            <ChevronDown className={cn("h-4 w-4 text-muted-foreground transition-transform", pickerOpen && "rotate-180")} />
          </button>

          {pickerOpen && (
            <div className="-mt-2 space-y-1 rounded-2xl bg-card p-1.5 ring-1 ring-border/50 stream-in">
              {projectNames.map((name) => (
                <button
                  key={name}
                  onClick={() => { setProjectName(name); setPickerOpen(false); }}
                  className={cn(
                    "flex w-full items-center gap-3 rounded-xl px-2.5 py-2 text-left transition",
                    name === projectName ? "bg-primary/10" : "hover:bg-surface",
                  )}
                >
                  <span className="h-2 w-2 rounded-full bg-success" />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-[13px] font-medium">{name}</p>
                    {localProjectMeta(name)?.repo && (
                      <p className="truncate font-mono text-[10.5px] text-muted-foreground">
                        {localProjectMeta(name)?.repo}
                      </p>
                    )}
                  </div>
                  {name === projectName && <Check className="h-3.5 w-3.5 text-primary" />}
                </button>
              ))}
              <Link
                to="/projects/new"
                className="block rounded-xl px-2.5 py-2 text-center text-[11px] text-primary hover:bg-surface"
              >
                + 新建项目
              </Link>
            </div>
          )}

          {/* Prompt — text + image + slash + submit */}
          <div className="relative rounded-2xl bg-card p-3 ring-1 ring-border/50 focus-within:ring-primary/50">
            {images.length > 0 && (
              <div className="mb-2 flex gap-2 overflow-x-auto pb-1">
                {images.map((img) => (
                  <div key={img.id} className="relative h-16 w-16 shrink-0 overflow-hidden rounded-lg ring-1 ring-border/50">
                    <Lightbox
                      src={img.url}
                      alt={img.name}
                      caption={img.name}
                      thumbClassName="h-full w-full object-cover"
                    />
                    <button
                      onClick={() => setImages((p) => p.filter((i) => i.id !== img.id))}
                      className="absolute right-0.5 top-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-background/80 text-foreground ring-1 ring-border/60"
                      aria-label="移除图片"
                    >
                      <X className="h-2.5 w-2.5" />
                    </button>
                  </div>

                ))}
              </div>
            )}
            <textarea
              ref={textareaRef}
              rows={asGoal ? 4 : 3}
              value={prompt}
              onChange={(e) => onPromptChange(e.target.value, e.target.selectionStart ?? e.target.value.length)}
              onKeyUp={(e) => {
                const el = e.currentTarget;
                setSlashToken(detectSlashToken(el.value, el.selectionStart ?? el.value.length));
              }}
              onBlur={() => setTimeout(() => setSlashToken(null), 150)}
              placeholder={
                asGoal
                  ? "描述你希望持续达成的目标，例如：保持主分支 CI 常绿..."
                  : currentExpert
                    ? `告诉 ${currentExpert.name} 要做什么... （输入 / 唤起命令）`
                    : "想让 agent 做什么？（输入 / 唤起命令）"
              }
              autoFocus
              className="w-full resize-none bg-transparent text-[14px] leading-relaxed outline-none placeholder:text-muted-foreground"
            />
            {slashToken && (
              <SlashMenu
                token={slashToken.token}
                onPick={applySlash}
                className="absolute left-3 right-3 top-full mt-1 w-auto"
              />
            )}
            <div className="mt-2 flex items-center gap-2 border-t border-border/40 pt-2">
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                multiple
                className="hidden"
                onChange={(e) => { onPickImages(e.target.files); e.target.value = ""; }}
              />
              <button
                onClick={() => fileInputRef.current?.click()}
                className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-surface text-muted-foreground ring-1 ring-border/40 transition hover:text-foreground hover:ring-primary/40"
                aria-label="添加图片"
              >
                <ImagePlus className="h-4 w-4" />
              </button>
              <button
                onClick={() => {
                  const el = textareaRef.current;
                  if (!el) return;
                  const caret = el.selectionStart ?? prompt.length;
                  const next = prompt.slice(0, caret) + "/" + prompt.slice(caret);
                  setPrompt(next);
                  requestAnimationFrame(() => {
                    el.focus();
                    el.setSelectionRange(caret + 1, caret + 1);
                    setSlashToken({ start: caret, token: "/" });
                  });
                }}
                className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-surface font-mono text-[13px] text-muted-foreground ring-1 ring-border/40 hover:text-foreground hover:ring-primary/40"
                aria-label="斜杠命令"
              >
                /
              </button>
              <button
                onClick={submit}
                disabled={submitting || (!prompt.trim() && images.length === 0) || (authed && !currentEngine)}
                className={cn(
                  "flex h-9 min-w-0 flex-1 items-center justify-center gap-1.5 rounded-full bg-primary px-3 text-[13px] font-semibold text-primary-foreground transition active:scale-[0.98] disabled:opacity-40",
                  !submitting && "glow-primary",
                )}
              >
                {submitting ? (
                  <>
                    <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    创建中…
                  </>
                ) : asGoal ? (
                  <>
                    <Target className="h-3.5 w-3.5" />
                    保存目标
                  </>
                ) : (
                  <>
                    <Zap className="h-3.5 w-3.5" />
                    {authed ? "派发任务" : "登录后派发"}
                  </>
                )}
              </button>
            </div>
            {submitError && (
              <p className="mt-2 text-center text-xs text-destructive">{submitError}</p>
            )}
          </div>

          {asGoal && (
            <NewTaskGoalPanel
              tokenBudget={tokenBudget}
              goalMode={goalMode}
              successCriteria={successCriteria}
              onTokenBudgetChange={setTokenBudget}
              onGoalModeChange={setGoalMode}
              onSuccessCriteriaChange={setSuccessCriteria}
            />
          )}

          {/* Quick templates */}
          <div className="flex flex-wrap gap-1.5">
            {(asGoal
              ? goalTemplates
              : currentExpert?.prompt
                ? [currentExpert.prompt.slice(0, 60)]
                : templates
            ).map((t) => (
              <button
                key={t}
                onClick={() => setPrompt(t)}
                className="rounded-full bg-surface px-2.5 py-1 text-[11px] text-foreground/80 ring-1 ring-border/40 hover:ring-primary/40"
              >
                {t}
              </button>
            ))}
          </div>

          {!asGoal && (
            <div className="flex flex-wrap gap-1">
              {SLASH_COMMANDS.slice(0, 6).map((c) => (
                <button
                  key={c.cmd}
                  onClick={() => setPrompt((cur: string) => (cur ? cur + " " : "") + c.cmd + " ")}
                  className="flex items-center gap-1 rounded-full bg-surface/60 px-2 py-0.5 font-mono text-[10.5px] text-muted-foreground ring-1 ring-border/40 hover:text-primary hover:ring-primary/40"
                >
                  {c.emoji} {c.cmd}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </MobileFrame>
  );
}
