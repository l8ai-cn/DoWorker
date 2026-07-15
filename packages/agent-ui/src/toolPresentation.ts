import {
  FileCode2,
  FileText,
  Globe2,
  Image,
  Presentation,
  Search,
  TerminalSquare,
  Video,
  Wrench,
  type LucideIcon,
} from "lucide-react";

export interface ToolPresentation {
  icon: LucideIcon;
  inputLabel: string;
  label: string;
  outputLabel: string;
}

export function toolPresentation(name: string): ToolPresentation {
  const value = name.toLowerCase();
  if (matches(value, "shell", "terminal", "exec", "command")) {
    return present(TerminalSquare, "Command", "Command", "Output");
  }
  if (matches(value, "filechange", "file_change", "edit", "write", "patch")) {
    return present(FileCode2, "File change", "Change", "Result");
  }
  if (matches(value, "read", "cat", "open_file")) {
    return present(FileText, "Read file", "Path", "Content");
  }
  if (matches(value, "search", "find", "query", "grep")) {
    return present(Search, "Search", "Query", "Matches");
  }
  if (matches(value, "browser", "playwright", "navigate", "screenshot")) {
    return present(Globe2, "Browser", "Action", "Result");
  }
  if (matches(value, "image", "draw", "generate_image")) {
    return present(Image, "Image generation", "Prompt", "Result");
  }
  if (matches(value, "ppt", "slide", "presentation")) {
    return present(Presentation, "Presentation", "Request", "Result");
  }
  if (matches(value, "video", "render_video")) {
    return present(Video, "Video generation", "Request", "Result");
  }
  return present(Wrench, humanize(name), "Input", "Output");
}

function matches(value: string, ...needles: string[]) {
  return needles.some((needle) => value.includes(needle));
}

function present(
  icon: LucideIcon,
  label: string,
  inputLabel: string,
  outputLabel: string,
): ToolPresentation {
  return { icon, inputLabel, label, outputLabel };
}

function humanize(value: string) {
  const text = value.replace(/([a-z0-9])([A-Z])/g, "$1 $2").replace(/[_-]+/g, " ");
  return text.charAt(0).toUpperCase() + text.slice(1);
}
