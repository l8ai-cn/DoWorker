const zhToolText: Record<string, string> = {
  Command: "命令",
  Output: "输出",
  "File change": "文件变更",
  Change: "变更",
  Result: "结果",
  "Read file": "读取文件",
  Path: "路径",
  Content: "内容",
  Search: "搜索",
  Query: "查询",
  Matches: "匹配结果",
  Browser: "浏览器",
  Action: "操作",
  "Image generation": "图片生成",
  Prompt: "提示词",
  Presentation: "演示文稿",
  Request: "请求",
  "Video generation": "视频生成",
  Input: "输入",
};

export function localizeToolText(value: string) {
  return zhToolText[value] ?? value;
}

export function englishFileChangeVerb(kind: string) {
  switch (kind.toLowerCase()) {
    case "add":
      return "Added";
    case "delete":
      return "Deleted";
    default:
      return "Updated";
  }
}

export function localizeFileChangeVerb(kind: string) {
  switch (kind.toLowerCase()) {
    case "add":
      return "新增";
    case "delete":
      return "删除";
    default:
      return "更新";
  }
}
