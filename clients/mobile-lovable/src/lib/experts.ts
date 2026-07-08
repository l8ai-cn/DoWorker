// Expert Agent library — 通用助手场景下的"专家 Agent"目录

export type ExpertCategory =
  | "office"
  | "research"
  | "dev"
  | "life"
  | "creative"
  | "data";

export const CATEGORY_META: Record<
  ExpertCategory,
  { label: string; icon: string }
> = {
  office: { label: "办公", icon: "💼" },
  research: { label: "研究", icon: "🔬" },
  dev: { label: "开发", icon: "💻" },
  life: { label: "生活", icon: "🌿" },
  creative: { label: "创作", icon: "🎨" },
  data: { label: "数据", icon: "📊" },
};

export interface ExpertConnector {
  id: string;
  name: string;
  required: boolean;
  connected: boolean;
}

export interface Expert {
  id: string;
  name: string;
  avatar: string; // emoji
  accent: "primary" | "accent" | "info" | "success" | "warning";
  category: ExpertCategory;
  tagline: string;
  description: string;
  author: string;
  official: boolean;
  featured?: boolean;
  rating: number; // 0-5
  usageCount: number;
  capabilities: string[];
  connectors: ExpertConnector[];
  suggestedPrompts: string[];
  recentRuns?: number;
  lastUsedAt?: string;
}

export const experts: Expert[] = [
  {
    id: "email-butler",
    name: "邮件管家",
    avatar: "📧",
    accent: "primary",
    category: "office",
    tagline: "分类收件箱、起草回复、跟进未回邮件",
    description:
      "帮你在 5 分钟内清空收件箱：识别重要邮件、批量起草回复、追踪待跟进事项。支持 Gmail / Outlook。",
    author: "Lovable Official",
    official: true,
    featured: true,
    rating: 4.9,
    usageCount: 12480,
    capabilities: ["读取邮件", "起草回复", "标签整理", "联系人识别", "日历联动"],
    connectors: [
      { id: "gmail", name: "Gmail", required: true, connected: true },
      { id: "calendar", name: "Google Calendar", required: false, connected: false },
    ],
    suggestedPrompts: [
      "今天有什么重要邮件？总结成 3 条",
      "帮我起草回复给昨天 Sarah 的那封",
      "整理本周所有需要跟进的线程",
    ],
    recentRuns: 12,
    lastUsedAt: "2h 前",
  },
  {
    id: "meeting-notes",
    name: "会议纪要",
    avatar: "🎙️",
    accent: "info",
    category: "office",
    tagline: "转写会议音频、提取决策与 action items",
    description:
      "把 1 小时的会议录音变成 1 页可分享的纪要：关键决策、待办分配、下次会议议题。",
    author: "Lovable Official",
    official: true,
    featured: true,
    rating: 4.8,
    usageCount: 8930,
    capabilities: ["音频转写", "说话人识别", "决策提取", "Action items", "分享到 IM"],
    connectors: [
      { id: "zoom", name: "Zoom", required: false, connected: false },
      { id: "notion", name: "Notion", required: false, connected: true },
    ],
    suggestedPrompts: [
      "把这段录音整理成纪要发到项目频道",
      "从上周所有会议中提取给我的 TODO",
    ],
    recentRuns: 3,
    lastUsedAt: "昨天",
  },
  {
    id: "deep-researcher",
    name: "深度研究员",
    avatar: "🔬",
    accent: "accent",
    category: "research",
    tagline: "多轮检索 · 引用溯源 · 输出研究报告",
    description:
      "针对一个问题进行 20+ 次搜索与阅读，产出带引用的 markdown 报告，附对比表格与关键论点。",
    author: "Lovable Official",
    official: true,
    featured: true,
    rating: 4.9,
    usageCount: 15600,
    capabilities: ["Web 搜索", "论文检索", "引用溯源", "对比表格", "PDF 输出"],
    connectors: [
      { id: "web-search", name: "Web Search", required: true, connected: true },
    ],
    suggestedPrompts: [
      "调研 2025 年 AI Agent 协议现状（ACP / MCP / AG-UI）",
      "对比 Vercel AI SDK 与 LangChain 的适用场景",
      "整理最近 3 个月关于长上下文 RAG 的关键论文",
    ],
    recentRuns: 5,
    lastUsedAt: "3h 前",
  },
  {
    id: "data-analyst",
    name: "数据分析师",
    avatar: "📊",
    accent: "info",
    category: "data",
    tagline: "读表格 · 跑 SQL · 画图 · 洞察结论",
    description:
      "上传 CSV/Excel 或连接数据库，自动做清洗、聚合、可视化，并写出 3 条业务结论。",
    author: "Lovable Official",
    official: true,
    rating: 4.7,
    usageCount: 6820,
    capabilities: ["Excel / CSV", "SQL", "可视化", "异常检测", "报告"],
    connectors: [
      { id: "warehouse", name: "数据仓库", required: false, connected: false },
    ],
    suggestedPrompts: [
      "分析这份销售表，找出增长最快的品类",
      "给我看看上季度用户留存曲线",
    ],
  },
  {
    id: "ppt-master",
    name: "PPT 制作专家",
    avatar: "🖼️",
    accent: "warning",
    category: "creative",
    tagline: "从大纲到成稿 · 配图 · 品牌模板",
    description:
      "输入一段主题或文档，产出 10-20 页 PPT 大纲、配图建议与可编辑 pptx。",
    author: "Lovable Official",
    official: true,
    rating: 4.6,
    usageCount: 4210,
    capabilities: ["大纲生成", "配图", "品牌模板", "导出 pptx"],
    connectors: [],
    suggestedPrompts: [
      "给这份研究报告做一版 15 分钟的对外分享 PPT",
      "帮我准备周会汇报：本周进展 + 下周计划",
    ],
  },
  {
    id: "contract-review",
    name: "合同审查",
    avatar: "📜",
    accent: "warning",
    category: "office",
    tagline: "识别风险条款 · 对比模板 · 修改建议",
    description:
      "针对 NDA / SOW / 服务协议逐条审查，标出高风险条款并给出修订建议，输出对比版。",
    author: "Lovable Official",
    official: true,
    rating: 4.7,
    usageCount: 2140,
    capabilities: ["条款抽取", "风险打分", "模板对比", "修订建议"],
    connectors: [],
    suggestedPrompts: [
      "帮我看看这份 NDA 有没有需要修改的地方",
      "对比这份 SOW 和我们的标准模板",
    ],
  },
  {
    id: "codex-dev",
    name: "Codex 开发者",
    avatar: "⚡",
    accent: "primary",
    category: "dev",
    tagline: "写代码 · 修 bug · 跑测试 · 发 PR",
    description:
      "接入本地开发机，能读代码、跑测试、提交 PR。支持自动审批策略与工具调用白名单。",
    author: "OpenAI",
    official: true,
    featured: true,
    rating: 4.8,
    usageCount: 23400,
    capabilities: ["代码读写", "Shell", "Git", "测试运行", "PR 提交"],
    connectors: [
      { id: "git", name: "Git 仓库", required: true, connected: true },
      { id: "ci", name: "CI 系统", required: false, connected: true },
    ],
    suggestedPrompts: [
      "修一下 CI 里最新失败的那个 test",
      "给 /orders 加一版 rate-limit 中间件",
      "review 我最近的 PR 并给出改进建议",
    ],
    recentRuns: 42,
    lastUsedAt: "17 分钟前",
  },
  {
    id: "bug-hunter",
    name: "Bug 排查专家",
    avatar: "🐛",
    accent: "warning",
    category: "dev",
    tagline: "读日志 · 复现 · 二分定位 · 给修复方案",
    description:
      "从 Sentry / 日志出发，自动复现问题、二分 commit、给出最小修复 diff。",
    author: "Lovable Official",
    official: true,
    rating: 4.7,
    usageCount: 3820,
    capabilities: ["日志分析", "Sentry", "Git bisect", "最小复现"],
    connectors: [
      { id: "sentry", name: "Sentry", required: true, connected: false },
    ],
    suggestedPrompts: [
      "Safari 白屏问题为什么这周才出现？",
      "这个 500 报错帮我定位到具体代码行",
    ],
  },
  {
    id: "travel-planner",
    name: "旅行规划师",
    avatar: "✈️",
    accent: "success",
    category: "life",
    tagline: "行程 · 机票 · 酒店 · 预算控制",
    description:
      "根据你的时间、预算、偏好，产出一版可直接预订的行程单，含机票/酒店/景点。",
    author: "Lovable Official",
    official: true,
    rating: 4.6,
    usageCount: 5230,
    capabilities: ["机票检索", "酒店比价", "行程编排", "地图路线"],
    connectors: [],
    suggestedPrompts: [
      "国庆 5 天日本行，预算 1.5w，两人",
      "周末深圳周边亲子游，推荐 3 个方案",
    ],
  },
  {
    id: "daily-report",
    name: "日报生成",
    avatar: "📝",
    accent: "accent",
    category: "office",
    tagline: "从 IM / commits / 日历 自动汇总日报",
    description:
      "根据你今天的 commits、日历、Slack 消息，自动生成一段可直接发送的日报。",
    author: "Lovable Official",
    official: true,
    rating: 4.5,
    usageCount: 9100,
    capabilities: ["Git 历史", "日历", "Slack", "模板化输出"],
    connectors: [
      { id: "slack", name: "Slack", required: false, connected: false },
      { id: "git", name: "Git", required: false, connected: true },
    ],
    suggestedPrompts: [
      "生成我今天的日报，发到 #team-daily",
      "汇总本周工作发给我 leader",
    ],
    recentRuns: 8,
    lastUsedAt: "今天 09:15",
  },
];

export function getExpert(id: string): Expert | undefined {
  return experts.find((e) => e.id === id);
}

export function expertsByCategory(): Record<ExpertCategory, Expert[]> {
  const out = {} as Record<ExpertCategory, Expert[]>;
  (Object.keys(CATEGORY_META) as ExpertCategory[]).forEach((k) => (out[k] = []));
  for (const e of experts) out[e.category].push(e);
  return out;
}
