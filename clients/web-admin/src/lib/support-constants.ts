export const statusLabels: Record<string, string> = {
  open: "待处理",
  in_progress: "处理中",
  resolved: "已解决",
  closed: "已关闭",
};

export const statusVariants: Record<string, "default" | "secondary" | "destructive" | "outline" | "success" | "warning"> = {
  open: "destructive",
  in_progress: "warning",
  resolved: "success",
  closed: "secondary",
};

export const categoryLabels: Record<string, string> = {
  bug: "缺陷",
  feature_request: "功能请求",
  usage_question: "使用问题",
  account: "账号",
  other: "其他",
};

export const categoryVariants: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  bug: "destructive",
  feature_request: "default",
  usage_question: "secondary",
  account: "outline",
  other: "secondary",
};

export const priorityLabels: Record<string, string> = {
  low: "低",
  medium: "中",
  high: "高",
};

export const priorityVariants: Record<string, "default" | "secondary" | "destructive" | "outline" | "warning"> = {
  low: "secondary",
  medium: "warning",
  high: "destructive",
};
