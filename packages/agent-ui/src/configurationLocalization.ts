const permissionLabels: Record<string, string> = {
  permissionMode: "权限",
  permission_mode: "权限",
};

export function localizeConfigurationLabel(id: string, fallback: string) {
  return (
    {
      ...permissionLabels,
      model: "模型",
      mode: "模式",
      effort: "推理强度",
      costControlMode: "成本控制",
    }[id] ?? fallback
  );
}

export function localizeConfigurationOption(
  id: string,
  value: string,
  fallback: string,
) {
  const permissionOptions = {
    default: "更改前询问",
    acceptEdits: "自动接受编辑",
    bypassPermissions: "无需确认",
    plan: "规划模式",
    bypass: "无需确认",
    ask_dangerous: "危险操作时询问",
    ask_any_write: "写入前询问",
  };
  const options: Record<string, Record<string, string>> = {
    permissionMode: permissionOptions,
    permission_mode: permissionOptions,
    mode: { default: "默认", auto: "自动执行", plan: "规划模式" },
    effort: { low: "低", medium: "中", high: "高", xhigh: "极高" },
    costControlMode: { on: "开启", off: "关闭" },
  };
  return options[id]?.[value] ?? fallback;
}

export function englishConfigurationLabel(id: string, fallback: string) {
  return id === "permissionMode" || id === "permission_mode"
    ? "Permissions"
    : fallback;
}

export function englishConfigurationOption(
  id: string,
  value: string,
  fallback: string,
) {
  if (id !== "permissionMode" && id !== "permission_mode") return fallback;
  return (
    {
      bypass: "Full access",
      ask_dangerous: "Ask for dangerous actions",
      ask_any_write: "Ask before writes",
    }[value] ?? fallback
  );
}
