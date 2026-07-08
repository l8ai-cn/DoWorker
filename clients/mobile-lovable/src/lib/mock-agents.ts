// Mock data reflecting ACP (Agent Client Protocol) shape.
import chartImg from "@/assets/mock-latency-chart.jpg";
import safariBlankImg from "@/assets/mock-safari-blank.jpg";

export { chartImg, safariBlankImg };

export type EventType =
  | "user_message"
  | "agent_message"
  | "agent_thought"
  | "tool_call"
  | "plan"
  | "permission_request"
  | "ask_user"
  | "phase"
  | "error";

export interface AskUserField {
  name: string;
  label: string;
  type: "text" | "textarea" | "select" | "radio" | "checkbox" | "number";
  options?: string[];
  placeholder?: string;
  required?: boolean;
  defaultValue?: string;
}
export interface AskUserForm {
  title: string;
  description?: string;
  fields: AskUserField[];
  submitLabel?: string;
}


export type ToolKind = "read" | "write" | "edit" | "shell" | "search" | "fetch" | "other";

export interface DiffHunk {
  header?: string;
  lines: { kind: "add" | "del" | "ctx" | "hunk"; text: string }[];
}

export interface PlanItem {
  text: string;
  status: "pending" | "in_progress" | "completed";
}

export interface AgentEvent {
  id: string;
  type: EventType;
  ts: string;
  title: string;
  detail?: string;
  // tool-call specifics
  tool?: string;
  toolKind?: ToolKind;
  status?: "pending" | "in_progress" | "completed" | "failed";
  duration?: string;
  // rich payloads
  command?: string;
  cwd?: string;
  output?: string; // shell stdout/stderr
  exitCode?: number;
  filePath?: string;
  additions?: number;
  deletions?: number;
  diff?: DiffHunk[];
  // plan
  plan?: PlanItem[];
  // search/fetch
  query?: string;
  results?: { title: string; url?: string; snippet?: string }[];
  // rich content on messages
  markdown?: string;
  images?: { src: string; caption?: string; alt?: string }[];
  attachments?: { name: string; kind: "image" | "file"; src?: string; note?: string }[];
  // ask_user dynamic form
  form?: AskUserForm;
  answer?: Record<string, string | boolean>;
  elicitationId?: string;
  // phase divider
  phaseIndex?: number;
  phaseTotal?: number;
  phaseEmoji?: string;
  phaseSummary?: string;
}


export type SessionStatus = "running" | "waiting_approval" | "completed" | "failed" | "idle";

export interface SessionMetrics {
  tokensIn: number;
  tokensOut: number;
  toolCalls: number;
  filesChanged: number;
  elapsed: string;
  cost?: string;
}

export interface AgentSession {
  id: string;
  projectId: string;
  title: string;
  agent: "Codex" | "Claude Code" | "Gemini CLI" | "Custom";
  branch: string;
  status: SessionStatus;
  updatedAt: string;
  eventCount: number;
  preview: string;
  metrics?: SessionMetrics;
  events: AgentEvent[];
}

export interface Project {
  id: string;
  name: string;
  repo: string;
  host: string;
  color: string;
  online: boolean;
  sessionIds: string[];
}

export const projects: Project[] = [
  { id: "p-api", name: "API Gateway", repo: "acme/api-gateway", host: "mac-studio-01", color: "primary", online: true, sessionIds: ["sx-9f2a", "sx-2c08"] },
  { id: "p-web", name: "Web App", repo: "acme/web", host: "linux-devbox", color: "accent", online: true, sessionIds: ["sx-mega", "sx-4b71"] },
  { id: "p-docs", name: "Docs & Marketing", repo: "acme/docs", host: "ci-runner-3", color: "info", online: false, sessionIds: ["sx-77e1", "sx-1122"] },
];

// --- Long-running realistic execution demo ---
const PHASE_TOTAL = 6;
const phase = (i: number, id: string, ts: string, title: string, emoji: string, summary: string): AgentEvent => ({
  id, type: "phase", ts, title, phaseIndex: i, phaseTotal: PHASE_TOTAL, phaseEmoji: emoji, phaseSummary: summary,
});

const MEGA_SESSION: AgentSession = {
  id: "sx-mega",
  projectId: "p-web",
  title: "将 analytics 管道从 Postgres 迁移到 ClickHouse（含双写、回填与灰度）",
  agent: "Claude Code",
  branch: "infra/clickhouse-migration",
  status: "running",
  updatedAt: "刚刚",
  eventCount: 46,
  preview: "Phase 5/6 — 灰度切换中，QPS 稳定，P95 ↓ 62%",
  metrics: { tokensIn: 184320, tokensOut: 42180, toolCalls: 31, filesChanged: 14, elapsed: "1h 42m", cost: "$2.47" },
  events: [
    { id: "u1", type: "user_message", ts: "09:12:00", title: "任务", detail: "把 events 表从 Postgres 迁到 ClickHouse。要求：不停机、双写一周、可回滚、Grafana 上 P95 下降至少 50%。" },

    phase(1, "p1", "09:12:04", "现状调研", "🔍", "梳理表结构、查询模式、上游写入路径"),
    { id: "t1", type: "agent_thought", ts: "09:12:06", title: "思考", detail: "先摸清 events 表规模、索引、慢查询和上下游依赖，再决定字段映射与分区策略。" },
    {
      id: "pl1", type: "plan", ts: "09:12:10", title: "全局计划",
      plan: [
        { text: "调研 events 表规模、索引、访问模式", status: "in_progress" },
        { text: "设计 ClickHouse schema（分区/排序键/物化列）", status: "pending" },
        { text: "搭双写通道（Kafka → PG + CH）", status: "pending" },
        { text: "回填历史数据（分片并发）", status: "pending" },
        { text: "灰度切读 5% → 50% → 100%", status: "pending" },
        { text: "下线 PG 写入路径 + 复盘 PR", status: "pending" },
      ],
    },
    {
      id: "sh1", type: "tool_call", ts: "09:12:18", title: "psql — 表体积与索引",
      tool: "shell", toolKind: "shell", status: "completed", duration: "820ms",
      command: "psql -c \"\\dt+ events\" -c \"\\di+ events*\" analytics", cwd: "~/ops",
      exitCode: 0,
      output: ` Schema |   Name   | Type  | Owner |    Size    | Description
--------+----------+-------+-------+------------+------
 public | events   | table | ops   |  742 GB    |
(1 row)

 idx_events_uid_ts       | 118 GB
 idx_events_kind         |  46 GB
 idx_events_ts_brin      | 2.1 GB`,
    },
    {
      id: "sh2", type: "tool_call", ts: "09:12:29", title: "pg_stat_statements — top 慢查询",
      tool: "shell", toolKind: "shell", status: "completed", duration: "1.4s",
      command: "psql -f ops/top_events_queries.sql", cwd: "~/ops", exitCode: 0,
      output: `                query                            | calls |  mean_ms | p95_ms
-------------------------------------------------+-------+----------+--------
 SELECT count(*) FROM events WHERE ...           |  912k |    1240  |  4180
 SELECT kind,count(*) FROM events GROUP BY 1 ... |   38k |    5420  | 14800
 SELECT * FROM events WHERE user_id=$1 ORDER ... |  4.2m |     82   |   410`,
    },
    {
      id: "sr1", type: "tool_call", ts: "09:12:41", title: "grep — 写入路径",
      tool: "search", toolKind: "search", status: "completed", duration: "180ms",
      query: "INSERT INTO events|events\\.insert",
      results: [
        { title: "services/ingest/src/writer.ts:42", snippet: "await db.query('INSERT INTO events ...', row)" },
        { title: "services/webhooks/src/handler.ts:118", snippet: "events.insert({ user_id, kind, payload })" },
        { title: "workers/backfill/index.ts:9", snippet: "for await (const chunk of source) events.insert(chunk)" },
      ],
    },
    { id: "am1", type: "agent_message", ts: "09:12:55", title: "现状小结",
      markdown: `**规模**：742 GB / 32 亿行 / 3 个写入点\n**热路径**：\`SELECT ... WHERE user_id=$1 ORDER BY ts DESC\`（4.2m 次/天）\n**痛点**：\`GROUP BY kind\` p95 = **14.8s**，索引膨胀到 118 GB。\n\nClickHouse 的 \`MergeTree\` + \`(user_id, ts)\` 排序键正好对上，预计 GROUP BY 掉到亚秒。` },

    phase(2, "p2", "09:14:10", "方案设计", "🧭", "ClickHouse schema、双写拓扑、回滚开关"),
    { id: "t2", type: "agent_thought", ts: "09:14:12", title: "分区策略权衡",
      detail: "按天分区 vs 按月：日分区利于回填并发但 part 数会爆；折中用月分区 + TTL 30 天冷存 → S3。" },
    {
      id: "ap1", type: "ask_user", ts: "09:14:24",
      title: "冷数据保留策略？",
      form: {
        title: "ClickHouse 冷数据放哪里？",
        description: "热分区留在本地 NVMe，30 天以上的分区做冷存。",
        fields: [
          { name: "cold", label: "冷存位置", type: "radio", options: ["S3 (推荐)", "GCS", "本地 HDD", "不冷存"], defaultValue: "S3 (推荐)", required: true },
          { name: "ttl_days", label: "热存天数", type: "number", defaultValue: "30", required: true },
          { name: "compression", label: "压缩", type: "select", options: ["ZSTD(3)", "ZSTD(9)", "LZ4"], defaultValue: "ZSTD(3)" },
        ],
        submitLabel: "确认",
      },
      answer: { cold: "S3 (推荐)", ttl_days: "30", compression: "ZSTD(3)" },
    },
    { id: "am2", type: "agent_message", ts: "09:15:01", title: "架构草图",
      markdown: `\`\`\`text
    ┌────────────┐   ┌──────────────┐   ┌────────────┐
ingest ─┤ Kafka topic├──▶│ PG writer    ├──▶│ Postgres   │
    │  events    │   └──────────────┘   └────────────┘
    │            │   ┌──────────────┐   ┌────────────┐
    │            ├──▶│ CH writer    ├──▶│ ClickHouse │
    └────────────┘   └──────────────┘   │ + S3 冷存  │
                                        └────────────┘
\`\`\`\n**回滚开关**：\`FEATURE_CH_READ_PCT=0..100\`（Consul 动态下发），任何时候可秒切回 PG。` },

    phase(3, "p3", "09:16:20", "基础设施 & Schema", "🏗️", "起集群、建表、部署 CH writer"),
    {
      id: "sh3", type: "tool_call", ts: "09:16:24", title: "terraform apply — 3-node CH cluster",
      tool: "shell", toolKind: "shell", status: "completed", duration: "4m 12s",
      command: "terraform apply -target=module.clickhouse -auto-approve", cwd: "~/infra",
      exitCode: 0,
      output: `module.clickhouse.aws_instance.ch[0]: Creating...
module.clickhouse.aws_instance.ch[1]: Creating...
module.clickhouse.aws_instance.ch[2]: Creating...
...
Apply complete! Resources: 14 added, 0 changed, 0 destroyed.
Outputs:
  ch_endpoint = "ch-prod.internal:9000"
  ch_zk       = "zk-1.internal:2181,zk-2.internal:2181,zk-3.internal:2181"`,
    },
    {
      id: "pr1", type: "permission_request", ts: "09:20:48", title: "写入 db/clickhouse/001_events.sql",
      tool: "fs.write", toolKind: "write", status: "pending",
      filePath: "db/clickhouse/001_events.sql", additions: 34, deletions: 0,
      detail: "创建 events 主表 + S3 冷存磁盘策略。执行前请审阅分区/排序键。",
      diff: [{
        header: "@@ +1,34 @@ db/clickhouse/001_events.sql",
        lines: [
          { kind: "add", text: "CREATE TABLE events ON CLUSTER prod (" },
          { kind: "add", text: "  ts        DateTime64(3, 'UTC') CODEC(Delta, ZSTD(3))," },
          { kind: "add", text: "  user_id   UInt64," },
          { kind: "add", text: "  kind      LowCardinality(String)," },
          { kind: "add", text: "  session   UUID," },
          { kind: "add", text: "  payload   String CODEC(ZSTD(3))," },
          { kind: "add", text: "  ingested  DateTime DEFAULT now()" },
          { kind: "add", text: ") ENGINE = ReplicatedMergeTree('/ch/{shard}/events','{replica}')" },
          { kind: "add", text: "PARTITION BY toYYYYMM(ts)" },
          { kind: "add", text: "ORDER BY (user_id, ts)" },
          { kind: "add", text: "TTL ts + INTERVAL 30 DAY TO DISK 's3_cold'" },
          { kind: "add", text: "SETTINGS storage_policy='hot_cold', index_granularity=8192;" },
        ],
      }],
    },
    {
      id: "sh4", type: "tool_call", ts: "09:21:10", title: "clickhouse-client — 执行 DDL",
      tool: "shell", toolKind: "shell", status: "completed", duration: "1.8s",
      command: "clickhouse-client --host ch-prod.internal -q \"$(cat db/clickhouse/001_events.sql)\"",
      cwd: "~/ops", exitCode: 0,
      output: `Ok.\nQuery id: 8f2e...\n0 rows in set. Elapsed: 1.732 sec.`,
    },

    phase(4, "p4", "09:23:00", "双写通道", "🔀", "上线 CH writer，灰度写入 & 一致性校验"),
    {
      id: "e-write", type: "tool_call", ts: "09:23:12", title: "改写 services/ingest/src/writer.ts",
      tool: "fs.edit", toolKind: "edit", status: "completed", duration: "180ms",
      filePath: "services/ingest/src/writer.ts", additions: 42, deletions: 8,
      diff: [{
        header: "@@ -30,10 +30,42 @@ export async function writeEvent(row)",
        lines: [
          { kind: "ctx", text: "export async function writeEvent(row: Event) {" },
          { kind: "del", text: "  await pg.query(INSERT_SQL, toPgRow(row));" },
          { kind: "add", text: "  // Dual-write. PG is source of truth until cutover." },
          { kind: "add", text: "  const pgP = pg.query(INSERT_SQL, toPgRow(row));" },
          { kind: "add", text: "  const chP = ch.insert('events', [toChRow(row)])" },
          { kind: "add", text: "    .catch((e) => { metrics.chWriteErr.inc(); log.warn({e}, 'ch write failed'); });" },
          { kind: "add", text: "  await pgP;              // block on PG only" },
          { kind: "add", text: "  void chP;               // CH failure must not affect ingest" },
          { kind: "add", text: "}" },
        ],
      }],
    },
    {
      id: "sh5", type: "tool_call", ts: "09:24:02", title: "部署 canary（1 pod）",
      tool: "shell", toolKind: "shell", status: "completed", duration: "38s",
      command: "kubectl -n prod set image deploy/ingest ingest=ghcr.io/acme/ingest:ch-dual@sha256:9a2c && kubectl rollout status deploy/ingest --timeout=120s",
      cwd: "~/ops", exitCode: 0,
      output: `deployment.apps/ingest image updated
Waiting for rollout to finish: 1 of 12 new replicas have been updated...
deployment "ingest" successfully rolled out`,
    },
    { id: "am3", type: "agent_message", ts: "09:25:20", title: "5 分钟观察", markdown: "canary 5 分钟：CH 写入 QPS = 4.1k，错误率 0.02%（3 起来自 payload > 1 MB，走了截断路径）。继续放量。" },
    {
      id: "sh6", type: "tool_call", ts: "09:31:04", title: "全量 rollout（12 pods）",
      tool: "shell", toolKind: "shell", status: "completed", duration: "2m 10s",
      command: "kubectl rollout resume deploy/ingest && kubectl rollout status deploy/ingest",
      cwd: "~/ops", exitCode: 0,
      output: `Waiting for rollout to finish: 4 of 12 new replicas have been updated...
Waiting for rollout to finish: 9 of 12 new replicas have been updated...
deployment "ingest" successfully rolled out`,
    },

    phase(5, "p5", "09:34:00", "历史回填 (32 亿行)", "📦", "月分片并发导入，checksum 校验"),
    {
      id: "sh-fail", type: "tool_call", ts: "09:34:22", title: "backfill 首次尝试 — 内存 OOM",
      tool: "shell", toolKind: "shell", status: "failed", duration: "6m 44s",
      command: "python workers/backfill.py --from=2023-01 --to=2025-12 --parallel=16", cwd: "~/ops",
      exitCode: 137,
      output: `[2025-07-05 09:34] shard 2023-01 → 42M rows OK
[2025-07-05 09:35] shard 2023-02 → 48M rows OK
[2025-07-05 09:38] shard 2023-05 → OOM (worker killed)
Traceback (most recent call last):
  File "workers/backfill.py", line 118, in _flush
    ch.insert('events', batch)
MemoryError: allocator ran out of memory (batch=500k rows, ~2.1 GB)
[FATAL] killed by cgroup oom-killer`,
    },
    { id: "err1", type: "error", ts: "09:41:08", title: "OOM at shard 2023-05",
      detail: "批大小 500k 行 × 2.1 GB 超过 worker cgroup 上限。降到 100k 行并按 payload 尺寸自适应。" },
    { id: "t3", type: "agent_thought", ts: "09:41:14", title: "复盘",
      detail: "payload 中位数 380 B，但 P99 是 42 KB。按行数切批是错的，应该按字节切。" },
    {
      id: "e-batch", type: "tool_call", ts: "09:41:30", title: "改进 backfill 分批算法",
      tool: "fs.edit", toolKind: "edit", status: "completed", duration: "210ms",
      filePath: "workers/backfill.py", additions: 18, deletions: 5,
      diff: [{
        header: "@@ -40,7 +40,20 @@ def _iter_batches(rows)",
        lines: [
          { kind: "del", text: "    for i in range(0, len(rows), 500_000):" },
          { kind: "del", text: "        yield rows[i:i+500_000]" },
          { kind: "add", text: "    # size-aware batching: cap at 128 MiB or 100k rows" },
          { kind: "add", text: "    buf, size = [], 0" },
          { kind: "add", text: "    for r in rows:" },
          { kind: "add", text: "        b = len(r['payload'])" },
          { kind: "add", text: "        if size + b > 128 * 1024 * 1024 or len(buf) >= 100_000:" },
          { kind: "add", text: "            yield buf; buf, size = [], 0" },
          { kind: "add", text: "        buf.append(r); size += b" },
          { kind: "add", text: "    if buf: yield buf" },
        ],
      }],
    },
    {
      id: "sh-bf", type: "tool_call", ts: "09:42:00", title: "backfill 第二次 — 全量导入",
      tool: "shell", toolKind: "shell", status: "in_progress",
      command: "python workers/backfill.py --from=2023-01 --to=2025-12 --parallel=8 --size-cap=128MiB",
      cwd: "~/ops",
      output: `[09:42:04] discovered 36 monthly shards (3.24B rows, 742 GB)
[09:44:12] shard 2023-01 done   42.1M rows /  9.4 GB / 128s ✓ checksum
[09:46:33] shard 2023-02 done   48.8M rows / 10.9 GB / 141s ✓ checksum
[09:49:01] shard 2023-03 done   51.2M rows / 11.7 GB / 148s ✓ checksum
[09:51:44] shard 2023-04 done   53.9M rows / 12.4 GB / 163s ✓ checksum
[09:54:30] shard 2023-05 done   55.1M rows / 12.9 GB / 166s ✓ checksum
[09:57:19] shard 2023-06 done   57.3M rows / 13.3 GB / 169s ✓ checksum
[10:00:12] shard 2023-07 done   59.0M rows / 14.1 GB / 173s ✓ checksum
[10:03:14] shard 2023-08 done   61.4M rows / 14.6 GB / 182s ✓ checksum
[10:06:22] shard 2023-09 done   63.8M rows / 15.0 GB / 188s ✓ checksum
[10:09:41] shard 2023-10 done   66.2M rows / 15.7 GB / 199s ✓ checksum
[10:13:08] shard 2023-11 done   68.5M rows / 16.4 GB / 207s ✓ checksum
[10:16:47] shard 2023-12 done   71.9M rows / 17.2 GB / 219s ✓ checksum
[10:20:39] shard 2024-01 done   74.6M rows / 17.9 GB / 232s ✓ checksum
[10:24:41] shard 2024-02 done   72.1M rows / 17.3 GB / 242s ✓ checksum
[10:28:55] running: 15/36  ████████░░░░░░░░░░░░░░  41%  ETA 42m`,
    },
    {
      id: "sql-cmp", type: "tool_call", ts: "10:29:20", title: "一致性校验 (row-count by day)",
      tool: "shell", toolKind: "shell", status: "completed", duration: "3.2s",
      command: "python ops/compare_counts.py --sample 200 --tolerance 0.0001",
      cwd: "~/ops", exitCode: 0,
      output: `sampled 200 days
  matched:  198  (99.00%)
  drift:      2  (0.19%, 0.14%)  → within tolerance
✅ consistency OK`,
    },

    phase(6, "p6", "10:32:00", "灰度切换 & 复盘", "🚦", "5% → 50% → 100% 切读，Grafana 观测"),
    {
      id: "ap2", type: "ask_user", ts: "10:32:12",
      title: "从 5% 开始灰度？",
      form: {
        title: "开始灰度切读",
        description: "读路径按用户 hash 分桶，可秒级回滚。",
        fields: [
          { name: "pct", label: "初始百分比", type: "radio", options: ["1%", "5%", "10%", "50%", "100%"], defaultValue: "5%", required: true },
          { name: "notify", label: "SLO 破线时", type: "select", options: ["自动回滚", "只告警", "不做操作"], defaultValue: "自动回滚" },
        ],
        submitLabel: "开始灰度",
      },
      answer: { pct: "5%", notify: "自动回滚" },
    },
    {
      id: "sh-consul", type: "tool_call", ts: "10:32:40", title: "consul kv — 灰度开关到 5%",
      tool: "shell", toolKind: "shell", status: "completed", duration: "180ms",
      command: "consul kv put feature/ch_read_pct 5", cwd: "~/ops", exitCode: 0,
      output: `Success! Data written to: feature/ch_read_pct`,
    },
    {
      id: "graf1", type: "tool_call", ts: "10:37:04", title: "Grafana — analytics-api P95",
      tool: "grafana.query", toolKind: "fetch", status: "completed", duration: "1.1s",
      results: [
        { title: "p95_latency_ms (5m avg)", snippet: "PG bucket: 1240 ms  |  CH bucket: 468 ms  ↓ 62%" },
        { title: "error_rate", snippet: "PG: 0.11%  |  CH: 0.08%" },
        { title: "cpu (analytics-api)", snippet: "no change (±2%)" },
      ],
    },
    { id: "am4", type: "agent_message", ts: "10:38:00", title: "5% 观察 5 分钟",
      markdown: "✅ 无回滚事件，CH 侧 P95 = **468 ms**（PG = 1240 ms，↓ **62%**），错误率没有升高。**准备放到 50%**。" },
    {
      id: "sh-50", type: "tool_call", ts: "10:43:11", title: "灰度提升到 50%",
      tool: "shell", toolKind: "shell", status: "completed", duration: "160ms",
      command: "consul kv put feature/ch_read_pct 50", cwd: "~/ops", exitCode: 0, output: "Success!",
    },
    { id: "t4", type: "agent_thought", ts: "10:53:22", title: "10 分钟观察", detail: "P95 稳定，CH replica lag < 200 ms，S3 冷存分区无写入错误。" },
    {
      id: "sh-100", type: "tool_call", ts: "10:53:50", title: "灰度提升到 100%",
      tool: "shell", toolKind: "shell", status: "in_progress",
      command: "consul kv put feature/ch_read_pct 100 && watch -n 5 ops/slo.sh",
      cwd: "~/ops",
      output: `Success!
[10:54:00] p95=452ms err=0.07% ok
[10:54:05] p95=449ms err=0.08% ok
[10:54:10] p95=461ms err=0.06% ok
[10:54:15] p95=458ms err=0.07% ok`,
    },
  ],
};



export const sessions: AgentSession[] = [
  {
    id: "sx-9f2a",
    projectId: "p-api",
    title: "Refactor auth middleware to use JWT rotation",
    agent: "Codex",
    branch: "feat/jwt-rotation",
    status: "waiting_approval",
    updatedAt: "2s",
    eventCount: 47,
    preview: "需要批准写入 src/middleware/auth.ts",
    metrics: { tokensIn: 12480, tokensOut: 3120, toolCalls: 9, filesChanged: 2, elapsed: "1m 24s", cost: "$0.08" },
    events: [
      {
        id: "e1",
        type: "user_message",
        ts: "14:22:01",
        title: "任务描述",
        detail: "Refactor auth middleware to use JWT rotation with 15min access token TTL, refresh token family, replay detection.",
      },
      {
        id: "e2",
        type: "agent_thought",
        ts: "14:22:04",
        title: "思考",
        detail: "先扫一遍 src/middleware/ 看现在的 session cookie 流程是怎么组织的，再决定改造粒度。JWT 轮换需要一个 refresh token store，Redis 是最轻的选择。",
      },
      {
        id: "e3",
        type: "plan",
        ts: "14:22:05",
        title: "执行计划",
        plan: [
          { text: "读取现有 auth 中间件与调用点", status: "completed" },
          { text: "引入 jose + ioredis 依赖", status: "completed" },
          { text: "实现 token family 轮换 & 复用检测", status: "in_progress" },
          { text: "更新 requireAuth / 登出流程", status: "pending" },
          { text: "补 3 个单元测试 + 一个集成测试", status: "pending" },
        ],
      },
      {
        id: "e4",
        type: "tool_call",
        ts: "14:22:07",
        title: "读取 src/middleware/auth.ts",
        tool: "fs.read",
        toolKind: "read",
        status: "completed",
        duration: "42ms",
        filePath: "src/middleware/auth.ts",
        detail: "182 行 · 4.1 KB · 最后修改 3 天前",
      },
      {
        id: "e5",
        type: "tool_call",
        ts: "14:22:11",
        title: "grep",
        tool: "search",
        toolKind: "search",
        status: "completed",
        duration: "310ms",
        query: "requireAuth\\(",
        results: [
          { title: "src/routes/api/users.ts:12", snippet: "app.get('/me', requireAuth(), ..." },
          { title: "src/routes/api/orders.ts:8", snippet: "app.post('/', requireAuth({ roles: ['user'] }), ..." },
          { title: "src/routes/api/admin.ts:4", snippet: "router.use(requireAuth({ roles: ['admin'] }))" },
        ],
      },
      {
        id: "e6",
        type: "tool_call",
        ts: "14:22:15",
        title: "bun add jose ioredis",
        tool: "shell",
        toolKind: "shell",
        status: "completed",
        duration: "3.2s",
        command: "bun add jose ioredis",
        cwd: "~/code/api-gateway",
        exitCode: 0,
        output: `bun add v1.1.30
 installed jose@5.9.6
 installed ioredis@5.4.1

 2 packages installed [3.19s]`,
      },
      {
        id: "e7",
        type: "agent_message",
        ts: "14:22:20",
        title: "依赖就绪，准备落盘变更",
        detail: "现有 requireAuth 有 3 个调用点，签名兼容不需要改。下面写入新的 auth 中间件。",
      },
      {
        id: "e8",
        type: "permission_request",
        ts: "14:22:22",
        title: "写入 src/middleware/auth.ts",
        tool: "fs.write",
        toolKind: "write",
        status: "pending",
        filePath: "src/middleware/auth.ts",
        additions: 142,
        deletions: 63,
        detail: "将替换现有 session 校验逻辑，保留 requireAuth 对外签名。",
        diff: [
          {
            header: "@@ -1,12 +1,18 @@ src/middleware/auth.ts",
            lines: [
              { kind: "del", text: "import { verifySession } from \"../lib/session\";" },
              { kind: "add", text: "import { SignJWT, jwtVerify } from \"jose\";" },
              { kind: "add", text: "import Redis from \"ioredis\";" },
              { kind: "ctx", text: "" },
              { kind: "add", text: "const ACCESS_TTL = 15 * 60;" },
              { kind: "add", text: "const REFRESH_TTL = 60 * 60 * 24 * 30;" },
              { kind: "add", text: "const redis = new Redis(process.env.REDIS_URL!);" },
              { kind: "ctx", text: "" },
              { kind: "add", text: "export async function issueTokenPair(userId: string) {" },
              { kind: "add", text: "  const jti = crypto.randomUUID();" },
              { kind: "add", text: "  const access = await new SignJWT({ sub: userId, jti })" },
              { kind: "add", text: "    .setProtectedHeader({ alg: \"HS256\" })" },
              { kind: "add", text: "    .setExpirationTime(\"15m\").sign(SECRET);" },
              { kind: "add", text: "  return { access, refresh: jti };" },
              { kind: "add", text: "}" },
            ],
          },
          {
            header: "@@ -44,9 +52,14 @@ export function requireAuth(opts?)",
            lines: [
              { kind: "ctx", text: "export function requireAuth(opts?: AuthOpts) {" },
              { kind: "ctx", text: "  return async (req, res, next) => {" },
              { kind: "del", text: "    const session = await verifySession(req.cookies.sid);" },
              { kind: "del", text: "    if (!session) return res.status(401).end();" },
              { kind: "add", text: "    const token = req.headers.authorization?.slice(7);" },
              { kind: "add", text: "    if (!token) return res.status(401).end();" },
              { kind: "add", text: "    const { payload } = await jwtVerify(token, SECRET);" },
              { kind: "add", text: "    if (await redis.sismember(\"revoked\", payload.jti))" },
              { kind: "add", text: "      return res.status(401).end();" },
              { kind: "ctx", text: "    next();" },
              { kind: "ctx", text: "  };" },
              { kind: "ctx", text: "}" },
            ],
          },
        ],
      },
    ],
  },
  {
    id: "sx-4b71",
    projectId: "p-web",
    title: "为 dashboard 添加 Playwright e2e",
    agent: "Claude Code",
    branch: "test/dashboard-e2e",
    status: "running",
    updatedAt: "12s",
    eventCount: 23,
    preview: "正在执行 bun test tests/dashboard.spec.ts...",
    metrics: { tokensIn: 5820, tokensOut: 1240, toolCalls: 4, filesChanged: 1, elapsed: "38s" },
    events: [
      { id: "e1", type: "user_message", ts: "14:19:44", title: "任务描述", detail: "为 /dashboard 页面添加 Playwright e2e，覆盖登录、筛选、导出。" },
      {
        id: "e2", type: "tool_call", ts: "14:20:02", title: "列出 tests/", tool: "fs.ls", toolKind: "read",
        status: "completed", duration: "12ms", filePath: "tests/",
        detail: "3 个文件：auth.spec.ts, home.spec.ts, playwright.config.ts",
      },
      {
        id: "e3", type: "tool_call", ts: "14:20:14", title: "bun test tests/dashboard.spec.ts",
        tool: "shell", toolKind: "shell", status: "in_progress",
        command: "bun test tests/dashboard.spec.ts", cwd: "~/code/web",
        output: `Running 6 tests using 2 workers

  ✓  1 dashboard › shows KPI cards (1.2s)
  ✓  2 dashboard › filters by date range (2.8s)
  ⣾  3 dashboard › exports CSV`,
      },
      {
        id: "e4", type: "ask_user", ts: "14:20:38",
        title: "导出 CSV 用哪种分隔符？",
        form: {
          title: "导出 CSV 用哪种分隔符？",
          description: "e2e 需要验证导出结果，请确认 CSV 的分隔与编码。",
          fields: [
            { name: "delimiter", label: "分隔符", type: "radio", options: [",", ";", "\\t"], defaultValue: ",", required: true },
            { name: "encoding", label: "编码", type: "select", options: ["UTF-8", "UTF-8 BOM", "GBK"], defaultValue: "UTF-8", required: true },
            { name: "includeHeader", label: "包含表头", type: "checkbox", defaultValue: "true" },
            { name: "note", label: "补充说明（可选）", type: "textarea", placeholder: "例如：Excel 打开需要 BOM" },
          ],
          submitLabel: "确认并继续",
        },
      },
    ],
  },
  {
    id: "sx-2c08",
    projectId: "p-api",
    title: "调查 checkout 页面在 Safari 上的白屏",
    agent: "Codex",
    branch: "bug/safari-checkout",
    status: "completed",
    updatedAt: "8m",
    eventCount: 61,
    preview: "已定位到 Intl.Segmenter polyfill 问题，PR #482 已提交",
    metrics: { tokensIn: 24160, tokensOut: 6840, toolCalls: 18, filesChanged: 3, elapsed: "8m 12s", cost: "$0.21" },
    events: [
      {
        id: "e1", type: "user_message", ts: "13:50:00",
        title: "任务描述",
        detail: "checkout 页面在 iOS Safari 16 上白屏，看看是什么问题。附一张用户反馈截图。",
        attachments: [
          { name: "safari-blank.jpg", kind: "image", src: safariBlankImg, note: "iOS 16.2 · iPhone 12" },
        ],
      },
      {
        id: "e2", type: "agent_thought", ts: "13:50:03",
        title: "思考",
        detail: "先看下 Sentry 上 Safari 的错误分布，再对比 build 输出。白屏一般是运行时抛错让整棵树卸载了。",
      },
      {
        id: "e3", type: "tool_call", ts: "13:50:07",
        title: "Sentry — Safari 16 错误 (24h)",
        tool: "sentry.search", toolKind: "fetch", status: "completed", duration: "1.4s",
        query: "browser.name:Safari browser.version:16.* environment:prod",
        results: [
          { title: "TypeError: undefined is not a constructor (Intl.Segmenter)", snippet: "412 events · 128 users · checkout/PriceSummary.tsx:47" },
          { title: "ChunkLoadError: Loading chunk 4821 failed", snippet: "9 events · 4 users" },
          { title: "AbortError: The operation was aborted", snippet: "3 events · 2 users" },
        ],
      },
      {
        id: "e4", type: "agent_message", ts: "13:50:12",
        title: "初步定位",
        markdown: `Sentry 里 **412 起** 报错都指向同一处：\n\n> \`TypeError: undefined is not a constructor\` — \`Intl.Segmenter\`\n\n影响的浏览器全部是 **iOS Safari 16.0–16.3**。iOS 16.4 才原生支持 \`Intl.Segmenter\`，白屏就是它抛错让 React 把整棵树卸载了。`,
      },
      {
        id: "e5", type: "tool_call", ts: "13:50:20",
        title: "读取 checkout/PriceSummary.tsx",
        tool: "fs.read", toolKind: "read", status: "completed", duration: "38ms",
        filePath: "src/checkout/PriceSummary.tsx",
        detail: "第 47 行使用 new Intl.Segmenter(locale, { granularity: 'grapheme' })",
      },
      {
        id: "e6", type: "tool_call", ts: "13:50:34",
        title: "拉取 caniuse 数据",
        tool: "fetch", toolKind: "fetch", status: "completed", duration: "612ms",
        query: "https://caniuse.com/mdn-javascript_builtins_intl_segmenter",
        results: [{ title: "Intl.Segmenter", snippet: "iOS Safari: 16.4+ · Chrome 87+ · Firefox 125+" }],
      },
      {
        id: "e7", type: "agent_thought", ts: "13:50:39",
        title: "选型",
        detail: "两个方案：(a) 替换成 Array.from(str) 手动分位；(b) 引入 `@formatjs/intl-segmenter` polyfill。文案里有 emoji 和中日韩，所以需要真正的 grapheme cluster，(b) 更安全。",
      },
      {
        id: "e8", type: "tool_call", ts: "13:50:52",
        title: "bun add @formatjs/intl-segmenter",
        tool: "shell", toolKind: "shell", status: "completed", duration: "2.8s",
        command: "bun add @formatjs/intl-segmenter",
        cwd: "~/code/web", exitCode: 0,
        output: `bun add v1.1.30
 installed @formatjs/intl-segmenter@11.7.7

 1 package installed [2.81s]`,
      },
      {
        id: "e9", type: "tool_call", ts: "13:51:04",
        title: "修改 src/polyfills.ts",
        tool: "fs.edit", toolKind: "edit", status: "completed", duration: "24ms",
        filePath: "src/polyfills.ts", additions: 6, deletions: 0,
        diff: [{
          header: "@@ -1,3 +1,9 @@ src/polyfills.ts",
          lines: [
            { kind: "ctx", text: "import \"core-js/actual/array/from\";" },
            { kind: "add", text: "" },
            { kind: "add", text: "if (typeof (Intl as any).Segmenter === \"undefined\") {" },
            { kind: "add", text: "  const mod = await import(\"@formatjs/intl-segmenter/polyfill-force\");" },
            { kind: "add", text: "  await import(\"@formatjs/intl-segmenter/locale-data/en\");" },
            { kind: "add", text: "  await import(\"@formatjs/intl-segmenter/locale-data/ja\");" },
            { kind: "add", text: "  void mod;" },
            { kind: "add", text: "}" },
          ],
        }],
      },
      {
        id: "e10", type: "tool_call", ts: "13:52:11",
        title: "bun run build && bun run preview",
        tool: "shell", toolKind: "shell", status: "completed", duration: "34s",
        command: "bun run build && bun run preview", cwd: "~/code/web", exitCode: 0,
        output: `vite v5.4.10 building for production...
✓ 1284 modules transformed.
dist/assets/index-a7c9f2.js   382.14 kB │ gzip: 118.67 kB
✓ built in 12.34s

  ➜  Local:   http://localhost:4173/`,
      },
      {
        id: "e11", type: "tool_call", ts: "13:52:58",
        title: "playwright — webkit @ /checkout",
        tool: "playwright.snapshot", toolKind: "fetch", status: "completed", duration: "4.1s",
        detail: "在 WebKit engine 下渲染 /checkout，DOM 挂载成功，无控制台错误。",
      },
      {
        id: "e12", type: "agent_message", ts: "13:57:44",
        title: "复盘 & PR",
        markdown: `## 修复完成 ✅

**根因**：\`Intl.Segmenter\` 在 iOS Safari <16.4 上抛 \`TypeError\`，导致 React 卸载整棵树。

**修复**：在 \`src/polyfills.ts\` 里做 feature-detect，缺失时按需加载 \`@formatjs/intl-segmenter\`（en/ja locale-data ≈ 18 KB gz）。

### 验证

| 环境 | Before | After |
| --- | --- | --- |
| iOS Safari 16.2 | 白屏 | 正常 |
| iOS Safari 17 | 正常 | 正常（未加载 polyfill）|
| Chrome 130 | 正常 | 正常（未加载 polyfill）|

WebKit engine 下的复现截图（修复前）：`,
        images: [
          { src: safariBlankImg, caption: "iOS Safari 16.2 上的白屏（修复前）", alt: "Safari blank screen" },
        ],
      },
      {
        id: "e13", type: "agent_message", ts: "13:58:02",
        title: "顺手看了下延迟",
        markdown: `checkout 页 P95 首屏也顺手看了一下 Grafana——修复前后无变化，polyfill 是动态 \`import()\`，主 bundle 没变大。`,
        images: [
          { src: chartImg, caption: "Grafana · checkout P95 latency (24h)", alt: "Latency chart" },
        ],
      },
      {
        id: "e14", type: "agent_message", ts: "13:58:22",
        title: "PR",
        markdown: `已提交 [**PR #482 — polyfill Intl.Segmenter for iOS Safari <16.4**](https://github.com/acme/web/pull/482)\n\n- \`src/polyfills.ts\` +6 \n- \`package.json\` +1 dep\n- \`CHANGELOG.md\` +2\n\n需要你 review 一下要不要把 \`zh-CN\` locale-data 也预加载。`,
      },
    ],
  },
  MEGA_SESSION,
  { id: "sx-77e1", projectId: "p-docs", title: "生成本季度 changelog", agent: "Custom", branch: "main", status: "failed", updatedAt: "1h", eventCount: 14, preview: "git log 解析失败：rate limit exceeded", events: [] },
  { id: "sx-1122", projectId: "p-docs", title: "把 pricing 页文案翻译为日语", agent: "Claude Code", branch: "i18n/ja", status: "idle", updatedAt: "3h", eventCount: 0, preview: "等待你下发任务", events: [] },
];

export function getSession(id: string) { return sessions.find((s) => s.id === id); }
export function getProject(id: string) { return projects.find((p) => p.id === id); }
export function getProjectSessions(id: string) { return sessions.filter((s) => s.projectId === id); }


export const statusMeta: Record<SessionStatus, { label: string; dotClass: string; textClass: string; ring: string }> = {
  running: { label: "运行中", dotClass: "bg-primary pulse-dot", textClass: "text-primary", ring: "ring-primary/30" },
  waiting_approval: { label: "待审批", dotClass: "bg-warning pulse-dot", textClass: "text-warning", ring: "ring-warning/30" },
  completed: { label: "已完成", dotClass: "bg-success", textClass: "text-success", ring: "ring-success/20" },
  failed: { label: "失败", dotClass: "bg-destructive", textClass: "text-destructive", ring: "ring-destructive/25" },
  idle: { label: "空闲", dotClass: "bg-muted-foreground/60", textClass: "text-muted-foreground", ring: "ring-border" },
};
