export type KBSourceType = "git" | "feishu" | "dingtalk" | "google";

export const KB_SOURCE_OPTIONS: { value: KBSourceType; label: string; description: string }[] = [
  { value: "git", label: "Git 仓库", description: "纯 Git 知识库，手动维护或通过 Pod Ingest" },
  { value: "feishu", label: "飞书文档", description: "从飞书知识库 Wiki 空间单向同步到 raw/feishu/" },
  { value: "dingtalk", label: "钉钉文档", description: "从钉钉知识库工作区单向同步到 raw/dingtalk/" },
  { value: "google", label: "Google Drive", description: "从 Google Drive 文件夹单向同步到 raw/google/" },
];

export const SOURCE_LABELS: Record<string, string> = {
  git: "Git",
  feishu: "飞书",
  dingtalk: "钉钉",
  google: "Google",
};

export const SYNC_STATUS_LABELS: Record<string, string> = {
  idle: "待机",
  syncing: "同步中",
  synced: "已同步",
  failed: "失败",
};

export interface SourceFieldDef {
  key: string;
  label: string;
  secret?: boolean;
  placeholder?: string;
  help?: string;
}

export const SOURCE_FIELD_DEFS: Record<Exclude<KBSourceType, "git">, SourceFieldDef[]> = {
  feishu: [
    { key: "app_id", label: "App ID", placeholder: "cli_xxx" },
    { key: "app_secret", label: "App Secret", secret: true },
    { key: "space_id", label: "Wiki Space ID", placeholder: "知识空间 ID" },
  ],
  dingtalk: [
    { key: "app_key", label: "App Key", placeholder: "dingxxx" },
    { key: "app_secret", label: "App Secret", secret: true },
    {
      key: "workspace_id",
      label: "Workspace ID",
      placeholder: "知识库工作区 ID",
    },
    {
      key: "operator_id",
      label: "Operator Union ID",
      placeholder: "操作者 unionId",
      help: "钉钉 Wiki API 需要操作者 unionId，可在管理后台查看",
    },
  ],
  google: [
    {
      key: "access_token",
      label: "Access Token",
      secret: true,
      help: "需 drive.readonly 权限；Token 过期后需重新配置",
    },
    { key: "folder_id", label: "Folder ID", placeholder: "Google Drive 文件夹 ID" },
  ],
};

export type SourceConfigForm = Record<string, string>;

export function emptySourceConfig(sourceType: Exclude<KBSourceType, "git">): SourceConfigForm {
  const fields = SOURCE_FIELD_DEFS[sourceType];
  return Object.fromEntries(fields.map((f) => [f.key, ""]));
}

export function parseSourceConfigJson(json: string | undefined): SourceConfigForm {
  if (!json) return {};
  try {
    const parsed = JSON.parse(json) as Record<string, unknown>;
    const out: SourceConfigForm = {};
    for (const [k, v] of Object.entries(parsed)) {
      if (typeof v === "string") out[k] = v;
    }
    return out;
  } catch {
    return {};
  }
}

export function buildSourceConfigJson(
  sourceType: Exclude<KBSourceType, "git">,
  form: SourceConfigForm,
  existing?: SourceConfigForm,
): string {
  const fields = SOURCE_FIELD_DEFS[sourceType];
  const out: Record<string, string> = {};
  for (const field of fields) {
    const value = form[field.key]?.trim() ?? "";
    if (field.secret && (value === "" || value === "***")) {
      const prev = existing?.[field.key];
      if (prev && prev !== "***") out[field.key] = prev;
      continue;
    }
    out[field.key] = value;
  }
  return JSON.stringify(out);
}

export function validateSourceConfig(
  sourceType: Exclude<KBSourceType, "git">,
  form: SourceConfigForm,
  existing?: SourceConfigForm,
): string | null {
  const fields = SOURCE_FIELD_DEFS[sourceType];
  for (const field of fields) {
    const value = form[field.key]?.trim() ?? "";
    const hasExisting =
      field.secret &&
      existing?.[field.key] &&
      existing[field.key] !== "***";
    if (!value && !hasExisting) {
      return `请填写${field.label}`;
    }
  }
  return null;
}

export function isExternalSource(sourceType: string): sourceType is Exclude<KBSourceType, "git"> {
  return sourceType === "feishu" || sourceType === "dingtalk" || sourceType === "google";
}

export function syncStatusVariant(
  status: string,
): "default" | "secondary" | "destructive" | "outline" {
  if (status === "failed") return "destructive";
  if (status === "synced") return "default";
  if (status === "syncing") return "secondary";
  return "outline";
}
