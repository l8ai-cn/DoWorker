import { DocNavigation } from "@/components/docs/DocNavigation";
import { DocsTable } from "@/components/docs/DocsTable";

const conceptRows = [
  { cells: ["Worker", "执行能力与运行环境配置", "被 Loop 或 Workflow 引用"] },
  { cells: ["目标 Loop", "为一个明确目标自主执行一次", "验证命令退出码为 0"] },
  { cells: ["Workflow", "按规则重复运行自动化任务", "Cron、API 或事件触发每次运行"] },
];

const loopFieldRows = [
  { cells: ["必填", "名称、执行 Worker、目标、验收标准、验证命令"] },
  { cells: ["可选边界", "最大迭代、Token 预算、总时长、无进展、同错、升级策略"] },
  { cells: ["明确不属于 Loop", "Cron、并发、回调、历史保留、跨运行会话"] },
];

const workflowFieldRows = [
  { cells: ["任务定义", "名称、Prompt 模板、Agent、Runner、仓库、模型资源"] },
  { cells: ["触发与调度", "API 触发、可选 Cron、Cron 表达式"] },
  { cells: ["运行治理", "执行模式、沙箱、会话保持、并发、超时、历史保留、回调"] },
];

export default function LoopAndWorkflowPage() {
  return (
    <div>
      <h1 className="mb-5 text-4xl font-bold">Loop 与 Workflow</h1>
      <p className="mb-10 max-w-3xl leading-relaxed text-muted-foreground">
        Worker、一次性目标 Loop 和可重复 Workflow 是三个独立概念。不要用 Cron 伪装 Loop，也不要用一次运行的成功声明代替验证。
      </p>

      <section className="mb-12">
        <h2 className="mb-4 text-2xl font-semibold">产品边界</h2>
        <DocsTable columns={[{ header: "概念" }, { header: "负责什么" }, { header: "结束或触发条件" }]} rows={conceptRows} />
      </section>

      <section className="mb-12">
        <h2 className="mb-3 text-2xl font-semibold">创建目标 Loop</h2>
        <p className="mb-4 leading-relaxed text-muted-foreground">
          Loop 选择已有 Worker 的不可变配置快照；Agent 完成后由 Runner 在同一工作区执行验证命令。只有退出码为 0 才会进入完成状态。
        </p>
        <DocsTable columns={[{ header: "字段类别" }, { header: "字段" }]} rows={loopFieldRows} />
      </section>

      <section className="mb-12">
        <h2 className="mb-3 text-2xl font-semibold">创建 Workflow</h2>
        <p className="mb-4 leading-relaxed text-muted-foreground">
          Workflow 是可复用的自动化定义，每次触发产生独立运行记录。它不追求一个跨多轮的单次目标。
        </p>
        <DocsTable columns={[{ header: "字段类别" }, { header: "字段" }]} rows={workflowFieldRows} />
      </section>

      <section className="mb-12 rounded-lg bg-surface-muted p-5">
        <h2 className="mb-2 text-lg font-semibold">设计依据</h2>
        <p className="text-sm leading-relaxed text-muted-foreground">
          目标 Loop 采用目标、完成条件、执行和独立验证的闭环。它参考 Codex Goals 与 Claude Code Goals 对长期目标和可验证完成条件的处理；周期性自动化则与 Codex Scheduled Tasks 对应，归入 Workflow。
        </p>
      </section>

      <DocNavigation />
    </div>
  );
}
