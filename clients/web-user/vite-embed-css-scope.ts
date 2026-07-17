import postcss from "postcss";
import type { Plugin } from "vite";

const scopeSelector = ".do-worker-app";

function splitTopLevel(selectorList: string): string[] {
  const parts: string[] = [];
  let depth = 0;
  let current = "";
  for (const character of selectorList) {
    if (character === "(" || character === "[") depth++;
    if (character === ")" || character === "]") {
      depth = Math.max(0, depth - 1);
    }
    if (character === "," && depth === 0) {
      parts.push(current);
      current = "";
    } else {
      current += character;
    }
  }
  if (current.trim() !== "") parts.push(current);
  return parts;
}

function prefixSelector(selector: string): string {
  const value = selector.trim();
  if (value === "" || value.startsWith(scopeSelector)) return value;
  const rootMatch = value.match(/^(?::root|html|body)\b(.*)$/);
  return rootMatch ? `${scopeSelector}${rootMatch[1]}` : `${scopeSelector} ${value}`;
}

const scopePlugin = (): postcss.Plugin => ({
  postcssPlugin: "scope-do-worker",
  AtRule(atRule) {
    if (atRule.name !== "layer") return;
    if (!atRule.nodes) {
      atRule.remove();
      return;
    }
    atRule.replaceWith(atRule.nodes);
  },
  Rule(rule) {
    const parent = rule.parent;
    if (parent?.type === "atrule" && /keyframes$/i.test((parent as postcss.AtRule).name)) {
      return;
    }
    if (rule.selectors.every((selector) => selector.trim().startsWith(scopeSelector))) {
      return;
    }
    rule.selector = splitTopLevel(rule.selector).map(prefixSelector).join(", ");
  },
});
scopePlugin.postcss = true;

function scopeCss(css: string): string {
  return postcss([scopePlugin()]).process(css, { from: undefined }).css;
}

export function scopeDoWorkerCss(): Plugin {
  return {
    name: "scope-do-worker-css",
    enforce: "post",
    generateBundle(_options, bundle) {
      for (const file of Object.values(bundle)) {
        if (file.type !== "asset" || !file.fileName.endsWith(".css")) continue;
        const css =
          typeof file.source === "string" ? file.source : Buffer.from(file.source).toString("utf8");
        file.source = scopeCss(css);
      }
    },
  };
}
