import { localizeToolText } from "./toolLocalization";

export interface ToolActivityCount {
  count: number;
  label: string;
}

export function englishToolActivityGroupSummary(
  counts: ToolActivityCount[],
) {
  return counts.map(englishCount).join(" · ");
}

export function chineseToolActivityGroupSummary(
  counts: ToolActivityCount[],
) {
  return counts.map(chineseCount).join(" · ");
}

function englishCount({ count, label }: ToolActivityCount) {
  switch (label) {
    case "Command":
      return `Ran ${count} ${plural(count, "command")}`;
    case "File change":
      return `Changed ${count} ${plural(count, "file")}`;
    case "Read file":
      return `Read ${count} ${plural(count, "file")}`;
    case "Search":
      return `Ran ${count} ${plural(count, "search", "searches")}`;
    case "Browser":
      return `Used browser ${count} ${plural(count, "time")}`;
    case "Image generation":
      return `Generated ${count} ${plural(count, "image")}`;
    case "Presentation":
      return `Generated ${count} ${plural(count, "presentation")}`;
    case "Video generation":
      return `Generated ${count} ${plural(count, "video")}`;
    default:
      return `Used ${label} ${count} ${plural(count, "time")}`;
  }
}

function chineseCount({ count, label }: ToolActivityCount) {
  switch (label) {
    case "Command":
      return `运行了 ${count} 个命令`;
    case "File change":
      return `修改了 ${count} 个文件`;
    case "Read file":
      return `读取了 ${count} 个文件`;
    case "Search":
      return `执行了 ${count} 次搜索`;
    case "Browser":
      return `使用浏览器 ${count} 次`;
    case "Image generation":
      return `生成了 ${count} 张图片`;
    case "Presentation":
      return `生成了 ${count} 份演示文稿`;
    case "Video generation":
      return `生成了 ${count} 个视频`;
    default:
      return `使用 ${localizeToolText(label)} ${count} 次`;
  }
}

function plural(count: number, singular: string, pluralForm = `${singular}s`) {
  return count === 1 ? singular : pluralForm;
}
